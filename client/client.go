package main

import (
	"context"
	"disysminiproject2/service"
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func StartClient(conn grpc.ClientConnInterface, messageSends <-chan string, messageReceives chan<- *service.UserMessage) {
	client := service.NewChittychatClient(conn)

	context := context.Background()
	inStream, err := client.Broadcast(context, &emptypb.Empty{})
	if err != nil {
		log.Fatal("Failed to receive chat message broadcast")
	}

	outStream, err := client.Publish(context)
	if err != nil {
		log.Fatal("Failed to open up for publishing chat messages")
	}

	setUsername(outStream) //THis is just to send username to server

	go listenForMessages(inStream, messageReceives)
	go messageSender(outStream, messageSends)
}

func setUsername(outStream service.Chittychat_PublishClient) {
	message := service.Message{Clock: clock, Content: ""}
	userMessage := service.UserMessage{User: username, Event: service.UserMessage_SET_USERNAME, Message: &message}
	outStream.Send(&userMessage)
}

func messageSender(stream service.Chittychat_PublishClient, messages <-chan string) {
	for {
		log.Println("Starting listener for sending messages (sending messages on <Enter>)")
		text := <-messages
		log.Println("Received message, will send to server")

		// Increment clock count
		clockMutex.Lock()
		clock[username]++
		clockMutex.Unlock()

		message := service.Message{Clock: clock, Content: text}
		userMessage := service.UserMessage{User: username, Event: service.UserMessage_MESSAGE, Message: &message}
		log.Printf("Sending message to sever with contents: %s, clock: %s", text, service.FormatVectorClockAsString(clock))
		stream.Send(&userMessage)
	}
}

func listenForMessages(stream service.Chittychat_BroadcastClient, messagesChannel chan<- *service.UserMessage) {
	log.Println("Starting to listen for messages from the chat server")
	for {
		msg, err := stream.Recv()
		if err != nil {
			log.Fatal("Failed to receive message")
		}

		log.Printf("Listener received message with content: %s\n", msg.Message.Content)

		clockMutex.Lock()
		clock = service.MergeClocks(clock, msg.Message.Clock)
		clock[username]++
		clockMutex.Unlock()

		if msg.Event != service.UserMessage_SET_USERNAME { //Is SET_USERNAME ever necessary? It has a formatting case, but is never printed?
			messagesChannel <- msg
		}
		msg.Message.Clock = clock

		log.Printf("%s\n", FormatMessageContent(msg))
		if msg.Event == service.UserMessage_DISCONNECT {
			// We don't want to keep dead users around, so we remove it from the
			// local clock map.
			clockMutex.Lock()
			delete(clock, msg.User)
			clockMutex.Unlock()
		}
	}
}

func FormatMessageContent(msg *service.UserMessage) string {
	fmtClock := service.FormatVectorClockAsString(msg.Message.Clock)
	switch msg.Event {
	case service.UserMessage_MESSAGE:
		return fmt.Sprintf("[%s] %s\n%s", msg.User, fmtClock, msg.Message.Content)
	case service.UserMessage_DISCONNECT:
		return fmt.Sprintf("%s\n%s disconnected from the chat", fmtClock, msg.User)
	case service.UserMessage_JOIN:
		return fmt.Sprintf("%s\n%s joined the chat", fmtClock, msg.User)
	case service.UserMessage_ERROR:
		return fmt.Sprintf("%s\n%s crashed and left the chat", fmtClock, msg.User)
	case service.UserMessage_SET_USERNAME:
		return fmt.Sprintf("%s\nset username to %s", fmtClock, msg.User)
	}
	return ""
}
