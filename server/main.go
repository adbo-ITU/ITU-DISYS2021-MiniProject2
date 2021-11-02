package main

import (
	"disysminiproject2/service"
	"log"
	"net"

	"google.golang.org/grpc"
)

func main() {
	// We start the server on the localhost's port 3333
	listener, err := net.Listen("tcp", "localhost:3333")
	if err != nil {
		log.Fatalf("Error while attempting to listen on port 3333: %v", err)
	}

	log.Println("Started server")
	// Instantiate the server
	server := grpc.NewServer()
	srv := NewServer()
	// Initialise the server clock
	srv.clock["server"] = 0
	service.RegisterChittychatServer(server, &srv)
	// We connect the server to the TCP listener
	server.Serve(listener)
}
