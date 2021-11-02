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

	clients   map[string]service.Chittychat_BroadcastServer // map from UID to broadcast stream
	usernames map[string]string                             // map from UID to username
	mutex     sync.Mutex
	clock     service.VectorClock

	manyIncomingMessages map[string]chan *service.UserMessage
}

func NewServer() ChittyChatServer {
	return ChittyChatServer{
		clients:              make(map[string]service.Chittychat_BroadcastServer),
		usernames:            make(map[string]string),
		clock:                make(map[string]uint32),
		manyIncomingMessages: make(map[string](chan *service.UserMessage)),
	}
}

func (c *ChittyChatServer) Publish(stream service.Chittychat_PublishServer) error {
	// Cache the username, as the last msg received on client disconnect will be nil
	var cachedUsername string

	for {
		msg, err := stream.Recv()

		// Store the uid of the publishing user
		if msg != nil {
			cachedUsername = msg.User
		}

		if e, errOk := status.FromError(err); errOk && err != nil && e.Code() == codes.Canceled {
			log.Printf("[%v] User exited\n", cachedUsername)
			c.broadcastServiceMessage("", cachedUsername, service.UserMessage_DISCONNECT)
			c.removeClient(cachedUsername) //Added this as diconnected users remined in clock
			return nil
		}
		if err != nil {
			log.Printf("[%v] Error on receive: %v\n", msg.User, err)
			c.broadcastServiceMessage("", msg.User, service.UserMessage_ERROR)
			c.removeClient(cachedUsername) //Added this, as disconnected users remained
			return err
		}

		c.mutex.Lock()
		c.clock = service.MergeClocks(c.clock, msg.Message.Clock)
		c.mutex.Unlock()
		c.incrementOwnClock()

		fmtClock := service.FormatVectorClockAsString(c.clock)
		log.Printf("[%v] %s %s", msg.User, fmtClock, msg.Message.Content)

		// Broadcast the received message to all other clients (and in fact the sender themselves)
		for _, channel := range c.manyIncomingMessages {
			channel <- msg
		}
	}
}

func (c *ChittyChatServer) Broadcast(_ *emptypb.Empty, stream service.Chittychat_BroadcastServer) error {
	// Each broadcast call will live in its own goroutine - serving one client each
	// https://github.com/grpc/grpc-go/blob/master/Documentation/concurrency.md#servers

	// Communication: messages clients needs to go to all clients

	uid := uuid.Must(uuid.NewRandom()).String()
	log.Printf("New user joined: %s\n", uid)

	// Create the communications channel to receive incoming messages destined for the belonging client
	c.manyIncomingMessages[uid] = make(chan *service.UserMessage, 20)

	usernameMsg := <-c.manyIncomingMessages[uid]
	if usernameMsg.Event != service.UserMessage_SET_USERNAME {
		log.Printf("Failed to initialise user %s\n", uid)
	}
	username := usernameMsg.User
	log.Printf("%s was assigned with username %s\n", uid, username)

	c.mutex.Lock() //Added the merging of clocks here, client clock should exist
	c.clock = service.MergeClocks(c.clock, usernameMsg.Message.Clock)
	c.mutex.Unlock()
	c.incrementOwnClock()

	err := c.addClient(uid, username, stream)
	if err != nil {
		log.Printf("Client join error: %v\n", err)
	}
	defer c.removeClient(uid)

	c.broadcastServiceMessage("", username, service.UserMessage_JOIN)

	for {
		msg := <-c.manyIncomingMessages[uid]

		c.mutex.Lock()
		c.clock = service.MergeClocks(c.clock, msg.Message.Clock)
		c.mutex.Unlock()
		c.incrementOwnClock()

		err = c.clients[uid].Send(msg)
		if err != nil {
			// Something must have gone wrong with sending, just drop the client
			break
		}
	}

	return nil
}

// func (c *ChittyChatServer) initUser(uid string, stream service.Chittychat_BroadcastServer) (string, error) {
// 	if usernameMessage.Event == service.UserMessage_SET_USERNAME {
// 		alreadyExists := false
// 		for _, username := range c.usernames {
// 			if username == usernameMessage.User {
// 				alreadyExists = true
// 				break
// 			}
// 		}

// 		eventType := service.UserMessage_SET_USERNAME
// 		if alreadyExists || usernameMessage.User == "server" {
// 			eventType = service.UserMessage_INVALID_USERNAME
// 		}
// 		c.incrementOwnClock()
// 		msg := service.UserMessage{Event: eventType, User: usernameMessage.User, Message: &service.Message{Content: "", Clock: c.clock}}
// 		stream.Send(&msg)
// 	} else {
// 		return "", fmt.Errorf("user failed to give username as first action")
// 	}
// 	return usernameMessage.User, nil
// }

func (c *ChittyChatServer) addClient(id string, username string, conn service.Chittychat_BroadcastServer) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, ok := c.clients[id]; ok {
		return fmt.Errorf("user id already exists: %s", id)
	}
	c.clients[id] = conn
	c.usernames[id] = username
	c.clock[username] = 0
	return nil
}

func (c *ChittyChatServer) removeClient(id string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.clients, id)
	delete(c.clock, id)
	delete(c.usernames, id) //Added this delete
}

func (c *ChittyChatServer) broadcastServiceMessage(content string, username string, event service.UserMessage_EventType) {
	// Note: this can't be rewritten at the momment, to also use the messaging channels to broadcast messages, as
	// some information is lost somehow, so clients don't register that they are told that someone joins or exits
	c.incrementOwnClock()
	message := service.UserMessage{Message: c.newMessage(content), User: username, Event: event}
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
