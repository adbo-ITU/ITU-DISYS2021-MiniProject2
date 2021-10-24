package main

import (
	"context"
	"disysminiproject2/service"
	"fmt"
	"log"
	"strconv"
	"time"

	"google.golang.org/grpc"
)

func main() {
	address := "127.0.0.1:3333"

	fmt.Print("Connecting.. ")
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("There were an error: %v", err)
	}
	fmt.Println("Done!")

	defer conn.Close()
	client := service.NewChittychatClient(conn)

	context := context.Background()

	fmt.Print("Sending message.. ")

	stream, err := client.ChatSession(context)
	if err != nil {
		log.Fatal("Failed to join chat room")
	}

	for i := 0; i < 10; i++ {
		message := service.Message{Clock: 0, Message: strconv.Itoa(i)}
		stream.Send(&message)
		time.Sleep(1000 * time.Millisecond)
	}

	fmt.Println("Done!")
}
