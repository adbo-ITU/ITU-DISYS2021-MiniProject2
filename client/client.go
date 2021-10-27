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

	msg, err := inStream.Recv()
	if msg.Event != service.UserMessage_SET_UID || err != nil {
		log.Fatalf("Failed to get uid: %v", err)
	}
	uid = msg.User
	clock = msg.Message.Clock
	go listenForMessages(inStream, messageReceives)
	go messageSender(outStream, messageSends)
}

func messageSender(stream service.Chittychat_PublishClient, messages <-chan string) {
	for {
		log.Println("Starting listener for sending messages (sending messages on <Enter>)")
		text := <-messages
		log.Println("Received message, will send to server")

		// Increment clock count
		clockMutex.Lock()
		clock[uid]++
		clockMutex.Unlock()

		message := service.Message{Clock: clock, Content: text}
		userMessage := service.UserMessage{User: uid, Event: service.UserMessage_MESSAGE, Message: &message}
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
		clock[uid]++
		fmtClock := service.FormatVectorClockAsString(clock)
		clockMutex.Unlock()
		messagesChannel <- msg
		switch msg.Event {
		case service.UserMessage_SET_UID:
			log.Printf("%v set uid to %s\n", fmtClock, msg.User)
			uid = msg.User
		case service.UserMessage_MESSAGE:
			log.Printf("%v %s: %s\n", fmtClock, msg.User, msg.Message.Content)
		case service.UserMessage_DISCONNECT:
			log.Printf("%v %s disconnected\n", fmtClock, msg.User)
			// We don't want to keep dead users around, so we remove it from the
			// local clock map.
			clockMutex.Lock()
			delete(clock, msg.User)
			clockMutex.Unlock()
		case service.UserMessage_JOIN:
			log.Printf("%v %s joined\n", fmtClock, msg.User)
		case service.UserMessage_ERROR:
			log.Printf("%v %s crashed\n", fmtClock, msg.User)
		}
	}
}

func FormatMessageContent(msg *service.UserMessage) string {
	switch msg.Event {
	case service.UserMessage_SET_UID:
		return fmt.Sprintf("Set uid to %v", msg.User)
	case service.UserMessage_MESSAGE:
		return fmt.Sprintf("[%s] %s", msg.User, msg.Message.Content)
	case service.UserMessage_DISCONNECT:
		return fmt.Sprintf("%s disconnected from the chat", msg.User)
	case service.UserMessage_JOIN:
		return fmt.Sprintf("%s joined the chat", msg.User)
	case service.UserMessage_ERROR:
		return fmt.Sprintf("%s crashed and left the chat", msg.User)
	}
	return ""
}
