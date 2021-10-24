package main

import (
	"context"
	"disysminiproject2/service"
	"log"

	"google.golang.org/protobuf/types/known/emptypb"
)

type ChittyChatServer struct {
	service.UnimplementedChittychatServer
}

func (ChittyChatServer) Publish(context context.Context, message *service.Message) (*emptypb.Empty, error) {
	log.Printf("Received message: %s", message.Message)

	return &emptypb.Empty{}, nil
}
