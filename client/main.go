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

	stream, err := client.ChatSession(context)
	if err != nil {
		log.Fatal("Failed to join chat room")
	}

	go listenForMessages(stream)

	for i := 0; i < 10; i++ {
		message := service.Message{Clock: 0, Content: strconv.Itoa(i)}
		stream.Send(&message)
		log.Printf("%v You: %s\n", message.Clock, message.Content)
		time.Sleep(1000 * time.Millisecond)
	}
}

func listenForMessages(stream service.Chittychat_ChatSessionClient) {
	for {
		msg, err := stream.Recv()
		if err != nil {
			log.Fatal("Failed to receive message")
		}

		switch msg.Event {
		case service.UserMessage_MESSAGE:
			log.Printf("%v %s: %s\n", msg.Message.Clock, msg.User, msg.Message.Content)
		case service.UserMessage_DISCONNECT:
			log.Printf("%v %s disconnected\n", msg.Message.Clock, msg.User)
		case service.UserMessage_JOIN:
			log.Printf("%v %s joined\n", msg.Message.Clock, msg.User)
		case service.UserMessage_ERROR:
			log.Printf("%v %s crashed\n", msg.Message.Clock, msg.User)
		}
	}
}
