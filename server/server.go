package main

import (
	"disysminiproject2/service"
	"fmt"
	"log"
	"sync"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Rewrite to publish/broadcast TODO:
/*
	* Rewrite to publish/broadcast TODO:
		- broadcastServiceMessage should perhaps also use new broadcast call, such that messages
		  are not sent to all clients, but are put into the channels for broadcasting to all
		- Look into the formatting of user messages, it looks weird
*/

type ChittyChatServer struct {
	service.UnimplementedChittychatServer

	clients map[string]service.Chittychat_BroadcastServer
	mutex   sync.Mutex
	clock   service.VectorClock

	incomingMessages     chan *service.UserMessage
	manyIncomingMessages map[string]chan *service.UserMessage
}

func NewServer() ChittyChatServer {
	return ChittyChatServer{
		clients:              make(map[string]service.Chittychat_BroadcastServer),
		clock:                make(map[string]uint32),
		incomingMessages:     make(chan *service.UserMessage, 1000),
		manyIncomingMessages: make(map[string](chan *service.UserMessage)),
	}
}

func (c *ChittyChatServer) Publish(stream service.Chittychat_PublishServer) error {
	// RECEIVE MESSAGES AND PASS TO BROADCAST

	// This is to prevent segmentation violations on user exit, as the msg received will be nil
	var cachedUID string

	for {
		msg, err := stream.Recv()
		log.Println("Recevied published message from client")

		// Store the uid of the publishing user
		if msg != nil {
			cachedUID = msg.User
		}

		if e, errOk := status.FromError(err); errOk && err != nil && e.Code() == codes.Canceled {
			log.Printf("[%v] User exited\n", cachedUID)
			c.broadcastServiceMessage("", cachedUID, service.UserMessage_DISCONNECT)
			return nil
		}
		if err != nil {
			log.Printf("[%v] Error on receive: %v\n", msg.User, err)
			c.broadcastServiceMessage("", msg.User, service.UserMessage_ERROR)
			return err
		}

		c.mutex.Lock()
		c.clock = service.MergeClocks(c.clock, msg.Message.Clock)
		c.mutex.Unlock()
		c.incrementOwnClock()

		fmtClock := service.FormatVectorClockAsString(c.clock)
		log.Printf("[%v] %s %s", msg.User, fmtClock, msg.Message)

		// Throw the message into the incoming messages channel
		// c.incomingMessages <- msg
		for _, channel := range c.manyIncomingMessages {
			channel <- msg
		}
	}

	return status.Errorf(codes.Unimplemented, "method Publish not implemented")
}

func provisionClient(stream service.Chittychat_BroadcastServer, server *ChittyChatServer, uid string) error {
	return stream.Send(&service.UserMessage{Message: server.newMessage(""), User: uid, Event: service.UserMessage_SET_UID})
}

func (c *ChittyChatServer) Broadcast(_ *emptypb.Empty, stream service.Chittychat_BroadcastServer) error {
	// Each broadcast call will live in its own goroutine - serving one client each
	// https://github.com/grpc/grpc-go/blob/master/Documentation/concurrency.md#servers

	// Communication: messages clients needs to go to all clientss

	log.Println("New user joined")
	uid := uuid.Must(uuid.NewRandom()).String()[0:4]

	// Create the communications channel to receive incoming messages destined for the belonging client
	c.manyIncomingMessages[uid] = make(chan *service.UserMessage, 20)

	c.incrementOwnClock()
	err := c.addClient(uid, stream)
	if err != nil {
		log.Printf("Client join error: %v\n", err)
	}
	defer c.removeClient(uid)

	c.incrementOwnClock()

	// provision the client means to set the client up with the correct configuration: setting its id
	err = provisionClient(stream, c, uid)
	if err != nil {
		log.Printf("Failed to send back UID: %v\n", err)
	}

	c.broadcastServiceMessage("", uid, service.UserMessage_JOIN)

	for {
		msg := <-c.manyIncomingMessages[uid]

		c.mutex.Lock()
		c.clock = service.MergeClocks(c.clock, msg.Message.Clock)
		c.mutex.Unlock()
		c.incrementOwnClock()

		fmtClock := service.FormatVectorClockAsString(c.clock)
		log.Printf("[%v] %s %s", msg.User, fmtClock, msg.Message.Content)
		// c.broadcastMessage(msg.Message.Content, msg.User, service.UserMessage_MESSAGE)

		// we're creating a message and making sure we're sending the information about which user it is coming from
		message := service.UserMessage{Message: c.newMessage(msg.Message.Content), User: msg.User, Event: service.UserMessage_MESSAGE}
		err = c.clients[uid].Send(&message)

		if err != nil {
			// lets just do this so that so the go linter thinks we can reach end of this function
			break
		}
	}

	// return status.Errorf(codes.Unimplemented, "method Broadcast not implemented")
	return nil
}

func (c *ChittyChatServer) ChatSession(stream service.Chittychat_ChatSessionServer) error {
	log.Println("New user joined")
	uid := uuid.Must(uuid.NewRandom()).String()[0:4]

	c.incrementOwnClock()
	err := c.addClient(uid, stream)
	if err != nil {
		log.Printf("Client join error: %v\n", err)
	}
	defer c.removeClient(uid)

	c.incrementOwnClock()
	err = stream.Send(&service.UserMessage{Message: c.newMessage(""), User: uid, Event: service.UserMessage_SET_UID})
	if err != nil {
		log.Printf("Failed to send back UID: %v\n", err)
	}

	c.broadcastServiceMessage("", uid, service.UserMessage_JOIN)

	for {
		msg, err := stream.Recv()

		if e, errOk := status.FromError(err); errOk && err != nil && e.Code() == codes.Canceled {
			log.Printf("[%v] User exited\n", uid)
			c.broadcastServiceMessage("", uid, service.UserMessage_DISCONNECT)
			return nil
		}
		if err != nil {
			log.Printf("[%v] Error on receive: %v\n", uid, err)
			c.broadcastServiceMessage("", uid, service.UserMessage_ERROR)
			return err
		}

		c.mutex.Lock()
		c.clock = service.MergeClocks(c.clock, msg.Clock)
		c.mutex.Unlock()
		c.incrementOwnClock()

		fmtClock := service.FormatVectorClockAsString(c.clock)
		log.Printf("[%v] %s %s", uid, fmtClock, msg.Content)
		c.broadcastServiceMessage(msg.Content, uid, service.UserMessage_MESSAGE)
	}
}

func (c *ChittyChatServer) getAllClients() map[string]service.Chittychat_BroadcastServer {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	clone := make(map[string]service.Chittychat_BroadcastServer)
	for k, v := range c.clients {
		clone[k] = v
	}
	return clone
}

func (c *ChittyChatServer) addClient(id string, conn service.Chittychat_BroadcastServer) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, ok := c.clients[id]; ok {
		return fmt.Errorf("user id already exists: %s", id)
	}
	c.clients[id] = conn
	c.clock[id] = 0
	return nil
}

func (c *ChittyChatServer) removeClient(id string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.clients, id)
	delete(c.clock, id)
}

func (c *ChittyChatServer) broadcastServiceMessage(content string, uid string, event service.UserMessage_EventType) {
	c.incrementOwnClock()
	for _, v := range c.getAllClients() {
		message := service.UserMessage{Message: c.newMessage(content), User: uid, Event: event}
		v.Send(&message)
	}
}

func (c *ChittyChatServer) newMessage(content string) *service.Message {
	return &service.Message{Clock: c.clock, Content: content}
}

func (c *ChittyChatServer) incrementOwnClock() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.clock["server"]++
}

// func (c *ChittyChatServer) getClient(id string) (service.Chittychat_ChatSessionServer, error) {
// 	c.mutex.Lock()
// 	defer c.mutex.Unlock()

// 	if _, ok := c.clients[id]; !ok {
// 		return nil, fmt.Errorf("user id does not exist: %s", id)
// 	}

// 	return c.clients[id], nil
// }
