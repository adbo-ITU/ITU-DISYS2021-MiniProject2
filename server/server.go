package main

import (
	"disysminiproject2/service"
	"fmt"
	"log"
	"sync"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ChittyChatServer struct {
	service.UnimplementedChittychatServer

	clients   map[string]service.Chittychat_ChatSessionServer
	usernames map[string]string
	mutex     sync.Mutex
	clock     service.VectorClock
}

func (c *ChittyChatServer) addClient(id string, conn service.Chittychat_ChatSessionServer) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, ok := c.clients[id]; ok {
		return fmt.Errorf("user id already exists: %s", id)
	}

	c.clients[id] = conn
	c.clock[id] = 0

	return nil
}

func (c *ChittyChatServer) addUsername(uid string, conn service.Chittychat_ChatSessionServer) error {

	message := service.UserMessage{Message: c.newMessage(""), User: c.usernames[uid], Event: service.UserMessage_SET_USERNAME}
	conn.Send(&message)

	msg, err := conn.Recv()

	if e, errOk := status.FromError(err); errOk && err != nil && e.Code() == codes.Canceled {
		log.Printf("[%v] User exited\n", uid)
		c.broadcastMessage("", uid, service.UserMessage_DISCONNECT)
		return nil
	}
	if err != nil {
		log.Printf("[%v] Error on receive: %v\n", uid, err)
		c.broadcastMessage("", uid, service.UserMessage_ERROR)
		return err
	}

	username := msg.Content

	if containsUsername(username, c.usernames) {
		log.Printf("User tried to join with existing username: %s\n", username)
		message := service.UserMessage{Message: c.newMessage(""), User: c.usernames[uid], Event: service.UserMessage_INVALID_USERNAME}
		conn.Send(&message)
		return fmt.Errorf("username already exists: %s", username)
	}

	c.usernames[uid] = username
	return nil
}

func (c *ChittyChatServer) removeClient(id string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.clients, id)
	delete(c.clock, id)
	delete(c.usernames, id)
}

// func (c *ChittyChatServer) getClient(id string) (service.Chittychat_ChatSessionServer, error) {
// 	c.mutex.Lock()
// 	defer c.mutex.Unlock()

// 	if _, ok := c.clients[id]; !ok {
// 		return nil, fmt.Errorf("user id does not exist: %s", id)
// 	}

// 	return c.clients[id], nil
// }

func (c *ChittyChatServer) getAllClients() map[string]service.Chittychat_ChatSessionServer {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	clone := make(map[string]service.Chittychat_ChatSessionServer)
	for k, v := range c.clients {
		clone[k] = v
	}
	return clone
}

func containsUsername(username string, usernames map[string]string) bool {

	for _, v := range usernames {
		if v == username {
			return true
		}
	}
	return false
}

func (c *ChittyChatServer) ChatSession(stream service.Chittychat_ChatSessionServer) error {

	log.Println("New user joined")
	uid := uuid.Must(uuid.NewRandom()).String()[0:4]

	c.incrementOwnClock()
	err := c.addClient(uid, stream)
	if err != nil {
		log.Printf("Client join error: %v\n", err)
		return err
	}
	defer c.removeClient(uid)

	c.incrementOwnClock()
	err = stream.Send(&service.UserMessage{Message: c.newMessage(""), User: uid, Event: service.UserMessage_SET_UID})
	if err != nil {
		log.Printf("Failed to send back UID: %v\n", err)
		return err
	}

	err = c.addUsername(uid, stream)
	if err != nil {
		return err
	}

	c.broadcastMessage("", uid, service.UserMessage_JOIN)

	for {
		msg, err := stream.Recv()

		if e, errOk := status.FromError(err); errOk && err != nil && e.Code() == codes.Canceled {
			log.Printf("[%v] User exited\n", uid)
			c.broadcastMessage("", uid, service.UserMessage_DISCONNECT)
			return nil
		}
		if err != nil {
			log.Printf("[%v] Error on receive: %v\n", uid, err)
			c.broadcastMessage("", uid, service.UserMessage_ERROR)
			return err
		}

		c.mutex.Lock()
		c.clock = service.MergeClocks(c.clock, msg.Clock)
		c.mutex.Unlock()
		c.incrementOwnClock()

		fmtClock := service.FormatVectorClockAsString(c.clock)
		log.Printf("[%v] %s %s", c.usernames[uid], fmtClock, msg.Content)
		c.broadcastMessage(msg.Content, uid, service.UserMessage_MESSAGE)
	}
}

func (c *ChittyChatServer) broadcastMessage(content string, uid string, event service.UserMessage_EventType) {
	c.incrementOwnClock()
	for k, v := range c.getAllClients() {
		if k != uid {
			message := service.UserMessage{Message: c.newMessage(content), User: c.usernames[uid], Event: event}
			v.Send(&message)
		}
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
