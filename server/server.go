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

func (c *ChittyChatServer) ChatSession(stream service.Chittychat_ChatSessionServer) error {
	log.Println("New user joined")
	uid := uuid.Must(uuid.NewRandom()).String()

	err := c.addClient(uid, stream)
	if err != nil {
		log.Printf("Client join error: %v\n", err)
	}

	defer c.removeClient(uid)

	for {
		msg, err := stream.Recv()

		if e, errOk := status.FromError(err); errOk && err != nil && e.Code() == codes.Canceled {
			log.Printf("[%v] User exited\n", uid)
			return nil
		}
		if err != nil {
			log.Printf("[%v] Error on receive: %v\n", uid, err)
			return err
		}

		log.Printf("[%v] %s", uid, msg.Message)
	}
}
