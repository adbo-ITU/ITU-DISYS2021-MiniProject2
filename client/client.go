package main

import (
	"context"
	"disysminiproject2/service"
	"fmt"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

//Create new client with separate channels for in- and outgoing messages
func StartClient(conn grpc.ClientConnInterface, messageSends <-chan string, messageReceives chan<- *service.UserMessage) {
	client := service.NewChittychatClient(conn)

	context := context.Background()
	inStream, err := client.Broadcast(context, &emptypb.Empty{}) //Abort in case of error in incoming stream
	if err != nil {
		log.Fatalf("Failed to receive chat message broadcast. Error message: %s", err)
	}

	//Set stream for outgoing messages and abort in case of error
	outStream, err := client.Publish(context)
	if err != nil {
		log.Fatalf("Failed to open up for publishing chat messages. Error message: %s", err)
	}

	// Notifies the server about the typed username
	setUsername(outStream)

	go listenForMessages(inStream, messageReceives) //Begin continously awaiting incoming messages
	go messageSender(outStream, messageSends)       //Begin continously awaiting outgoing messages
}

//Send username to server
func setUsername(outStream service.Chittychat_PublishClient) {
	incrementOwnClock()
	message := service.Message{Clock: clock, Content: ""}
	userMessage := service.UserMessage{User: username, Event: service.UserMessage_SET_USERNAME, Message: &message}
	log.Print(FormatMessageContent(&userMessage))
	outStream.Send(&userMessage)
}

func incrementOwnClock() {
	//Update Lamport timestamp by incrementing own clock before sending message
	clockMutex.Lock()
	defer clockMutex.Unlock()

	clock[username]++
}

//Continously listen for outgoing user messages
func messageSender(stream service.Chittychat_PublishClient, messages <-chan string) {
	for {
		text := <-messages

		incrementOwnClock()

		message := service.Message{Clock: clock, Content: text}
		userMessage := service.UserMessage{User: username, Event: service.UserMessage_MESSAGE, Message: &message}
		log.Printf("Sending message to sever with contents: %s, clock: %s", text, service.FormatVectorClockAsString(clock))
		stream.Send(&userMessage) //Send message to server through publish
	}
}

//Continously listen for incoming messages to print in user UI
func listenForMessages(stream service.Chittychat_BroadcastClient, messagesChannel chan<- *service.UserMessage) {
	log.Println("Starting to listen for messages from the chat server")
	for {
		//Abort program in case of error on incoming broadcast stream
		msg, err := stream.Recv()
		if err != nil {
			log.Fatalf("Failed to receive message, error: %s", err)
		}

		log.Printf("Client received message: %s\n", FormatMessageContent(msg))

		//Merge clocks according to Lamport timestamps and increase own clock on receive
		clockMutex.Lock()
		clock = service.MergeClocks(clock, msg.Message.Clock)
		clock[username]++
		clockMutex.Unlock()

		//Only print relevant messages to client
		if msg.Event != service.UserMessage_SET_USERNAME {
			messagesChannel <- msg
		}
		msg.Message.Clock = clock

		if msg.Event == service.UserMessage_DISCONNECT {
			// We don't want to keep dead users around, so we remove it from the
			// local clock map.
			clockMutex.Lock()
			delete(clock, msg.User)
			clockMutex.Unlock()
		}
	}
}

//Format incomming messages for print in user terminal based on message event type
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
