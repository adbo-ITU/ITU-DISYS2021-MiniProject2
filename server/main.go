package main

import (
	"disysminiproject2/service"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
)

func main() {
	listener, err := net.Listen("tcp", "localhost:3333")
	if err != nil {
		log.Fatalf("Error while attempting to listen on port 3333: %v", err)
		os.Exit(1)
	}

	log.Println("Starting server")
	server := grpc.NewServer()
	service.RegisterChittychatServer(server, &ChittyChatServer{})
	server.Serve(listener)
}
