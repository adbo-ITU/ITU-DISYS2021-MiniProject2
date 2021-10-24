package main

import (
	"disysminiproject2/service"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ChittyChatServer struct {
	service.UnimplementedChittychatServer
}

func (c *ChittyChatServer) ChatSession(stream service.Chittychat_ChatSessionServer) error {
	log.Println("New user joined")

	for {
		msg, err := stream.Recv()

		if e, errOk := status.FromError(err); errOk && err != nil && e.Code() == codes.Canceled {
			log.Println("User exited")
			return nil
		}
		if err != nil {
			log.Fatalf("Error on receive: %v", err)
			return err
		}

		log.Println(msg.Message)
	}
}
