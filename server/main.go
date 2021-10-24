package main

import (
	"disysminiproject2/service"
	"log"
	"net"

	"google.golang.org/grpc"
)

func main() {
	listener, err := net.Listen("tcp", "localhost:3333")
	if err != nil {
		log.Fatalf("Error while attempting to listen on port 3333: %v", err)
	}

	log.Println("Started server")
	server := grpc.NewServer()
	srv := ChittyChatServer{
		clients: make(map[string]service.Chittychat_ChatSessionServer),
	}
	service.RegisterChittychatServer(server, &srv)
	server.Serve(listener)
}
