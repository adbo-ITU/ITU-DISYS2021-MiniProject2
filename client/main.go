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

// Client structure with own username and local copy of vectorclock
var (
	username   string
	clock      service.VectorClock
	clockMutex sync.Mutex
)

//Create new client
func main() {

	//Prompt user for wanted username
	fmt.Print("Please enter your wanted username: ")
	fmt.Scan(&username)

	//Ensure file is closed after client disconnects by use of defer
	// Create a file name or the clients log that is infixed with a unit timestamp
	// in nanoseconds, to hopefully avoid shared/overwritten log files
	logFilename := fmt.Sprintf("client-%s.log", username)
	os.Remove(logFilename) // Clean up from a previous client run
	f, err := os.OpenFile(logFilename, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("error opening file: %s", err)
	}
	defer f.Close()
	log.SetOutput(f)

	log.Println("Starting the system")
	address := "127.0.0.1:3333" //Set address depending on local port

	//Try connecting to server and abort program in case of conenction error
	fmt.Print("Connecting.. ")
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("There were an error: %s", err)
	}
	fmt.Println("Done!")
	defer conn.Close()

	//Make clock map, that contains clients local copy of the vector clock, and add the user
	clock = make(service.VectorClock)
	clock[username] = 0

	chatEvents := make(chan (*service.UserMessage), 1000) //Create channel for incomming messages
	messageStream := make(chan (string))                  //Create channel for outgoing messages
	StartClient(conn, messageStream, chatEvents)

	// Create the UI
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %s", err)
	}
	defer ui.Close()

	// going to listen to this channel later to stop main thread from exiting
	systemExitChannel := make(chan (bool))

	theUI := NewUI(chatEvents, messageStream)
	theUI.uiEvents = ui.PollEvents()
	theUI.Render()

	//Ensure that UI continously handles events and messages
	go theUI.HandleUIEvents(systemExitChannel)
	go theUI.HandleChatMessages()

	// Prevent program exit before something is sent on this
	<-systemExitChannel
}
