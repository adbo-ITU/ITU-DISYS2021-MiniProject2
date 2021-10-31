package main

import (
	"context"
	"disysminiproject2/service"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"
)

// This is dirty
var (
	uid        string
	username   string
	clock      service.VectorClock
	clockMutex sync.Mutex
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

	msg, err := stream.Recv()
	if msg.Event != service.UserMessage_SET_UID || err != nil {
		log.Fatalf("Failed to get uid: %v", err)
	}

	uid = msg.User
	clock = msg.Message.Clock

	clockMutex.Lock()
	clock[uid]++
	clockMutex.Unlock()

	msg, err = stream.Recv()
	if msg.Event != service.UserMessage_SET_USERNAME || err != nil {
		log.Fatalf("Failed to get usernamerequest: %v", err)
	}

	clockMutex.Lock()
	clock[uid]++
	fmtClock := service.FormatVectorClockAsString(clock)
	clockMutex.Unlock()

	fmt.Printf("%v\nPlease enter wanted username: ", fmtClock)
	fmt.Scan(&username)

	message := service.Message{Clock: clock, Content: username}
	stream.Send(&message)

	go listenForMessages(stream)

	for i := 0; i < 10; i++ {
		clockMutex.Lock()
		clock[uid]++
		clockMutex.Unlock()
		message := service.Message{Clock: clock, Content: strconv.Itoa(i)}
		stream.Send(&message)
		log.Printf("%v You (%s): %s\n", service.FormatVectorClockAsString(message.Clock), uid, message.Content)
		time.Sleep(1000 * time.Millisecond)
	}

	fmt.Println("Press Enter to exit")
	fmt.Scanln()
}

func listenForMessages(stream service.Chittychat_ChatSessionClient) {
	for {
		msg, err := stream.Recv()
		if err != nil {
			log.Fatal("Failed to receive message")
		}

		clockMutex.Lock()
		clock = service.MergeClocks(clock, msg.Message.Clock)
		clock[uid]++
		fmtClock := service.FormatVectorClockAsString(clock)
		clockMutex.Unlock()

		switch msg.Event {
		case service.UserMessage_SET_UID:
			log.Printf("%v set uid to %s\n", fmtClock, msg.User)
			uid = msg.User
		case service.UserMessage_INVALID_USERNAME:
			log.Printf("%v Usernmae %s already taken. Please rejoin with new name\n", fmtClock, username)
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
