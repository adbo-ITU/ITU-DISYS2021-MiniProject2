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

type ChittyChatServer struct {
	service.UnimplementedChittychatServer

	clients map[string]service.Chittychat_BroadcastServer
	mutex   sync.Mutex
	clock   service.VectorClock

	manyIncomingMessages map[string]chan *service.UserMessage
}

func NewServer() ChittyChatServer {
	return ChittyChatServer{
		clients:              make(map[string]service.Chittychat_BroadcastServer),
		clock:                make(map[string]uint32),
		manyIncomingMessages: make(map[string](chan *service.UserMessage)),
	}
}

func (c *ChittyChatServer) Publish(stream service.Chittychat_PublishServer) error {
	// Cache the UID, as the last msg received on client disconnect will be nil
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

		// Broadcast the received message to all other clients (and in fact the sender themselves)
		for _, channel := range c.manyIncomingMessages {
			channel <- msg
		}
	}

	return nil
}

func (c *ChittyChatServer) Broadcast(_ *emptypb.Empty, stream service.Chittychat_BroadcastServer) error {
	// Each broadcast call will live in its own goroutine - serving one client each
	// https://github.com/grpc/grpc-go/blob/master/Documentation/concurrency.md#servers

	// Communication: messages clients needs to go to all clients

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

		err = c.clients[uid].Send(msg)
		if err != nil {
			// Something must have gone wrong with sending, just drop the client
			break
		}
	}

	return nil
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
	// Note: this can't be rewritten at the momment, to also use the messaging channels to broadcast messages, as
	// some information is lost somehow, so clients don't register that they are told that someone joins or exits
	c.incrementOwnClock()
	message := service.UserMessage{Message: c.newMessage(content), User: uid, Event: event}
	for _, channel := range c.manyIncomingMessages {
		channel <- &message
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

func provisionClient(stream service.Chittychat_BroadcastServer, server *ChittyChatServer, uid string) error {
	return stream.Send(&service.UserMessage{Message: server.newMessage(""), User: uid, Event: service.UserMessage_SET_UID})
}
