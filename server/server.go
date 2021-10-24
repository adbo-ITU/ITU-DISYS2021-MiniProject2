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

	clients map[string]service.Chittychat_ChatSessionServer
	mutex   sync.Mutex
}

func (c *ChittyChatServer) addClient(id string, conn service.Chittychat_ChatSessionServer) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, ok := c.clients[id]; ok {
		return fmt.Errorf("user id already exists: %s", id)
	}
	c.clients[id] = conn
	return nil
}

func (c *ChittyChatServer) removeClient(id string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.clients, id)
}

func (c *ChittyChatServer) getAllClients() map[string]service.Chittychat_ChatSessionServer {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	clone := make(map[string]service.Chittychat_ChatSessionServer)
	for k, v := range c.clients {
		clone[k] = v
	}
	return clone
}

func (c *ChittyChatServer) ChatSession(stream service.Chittychat_ChatSessionServer) error {
	log.Println("New user joined")
	uid := uuid.Must(uuid.NewRandom()).String()

	err := c.addClient(uid, stream)
	if err != nil {
		log.Printf("Client join error: %v\n", err)
	}

	defer c.removeClient(uid)

	c.broadcastMessage(newMessage(""), uid, service.UserMessage_JOIN)

	for {
		msg, err := stream.Recv()

		if e, errOk := status.FromError(err); errOk && err != nil && e.Code() == codes.Canceled {
			log.Printf("[%v] User exited\n", uid)
			c.broadcastMessage(newMessage(""), uid, service.UserMessage_DISCONNECT)
			return nil
		}
		if err != nil {
			log.Printf("[%v] Error on receive: %v\n", uid, err)
			c.broadcastMessage(newMessage(""), uid, service.UserMessage_ERROR)
			return err
		}

		log.Printf("[%v] %s", uid, msg.Content)
		c.broadcastMessage(msg, uid, service.UserMessage_MESSAGE)

	}
}

func (c *ChittyChatServer) broadcastMessage(msg *service.Message, user string, event service.UserMessage_EventType) {
	message := service.UserMessage{Message: msg, User: user, Event: event}

	for k, v := range c.getAllClients() {
		if k != user {
			v.Send(&message)
		}
	}
}

func newMessage(content string) *service.Message {
	return &service.Message{Clock: 0, Content: content}
}
