package main

import (
	"disysminiproject2/service"
	"fmt"
	"log"
	"os"
	"sync"

	ui "github.com/gizak/termui/v3"
	"google.golang.org/grpc"
)

// This is dirty
var (
	uid        string
	clock      service.VectorClock
	clockMutex sync.Mutex
)

func main() {

	f, err := os.OpenFile("other.log", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)

	log.Println("Starting the system")
	address := "127.0.0.1:3333"

	fmt.Print("Connecting.. ")
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("There were an error: %v", err)
	}
	fmt.Println("Done!")
	defer conn.Close()

	// Create the UI
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	// going to listen to this channel later to stop main thread from exiting
	systemExitChannel := make(chan (bool))

	theUI := NewUI()
	StartClient(conn, theUI.messageStream, theUI.chatEvents)
	theUI.uiEvents = ui.PollEvents()
	theUI.Render()

	go theUI.HandleUIEvents(systemExitChannel)
	go theUI.HandleChatMessages()

	// Prevent program exit before something is sent on this
	<-systemExitChannel
}
