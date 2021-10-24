package main

import (
	"context"
	"disysminiproject2/service"
	"fmt"
	"os"
	"time"

	"google.golang.org/grpc"
)

var (
	clock uint64 = 0
)

func main() {
	address := "127.0.0.1:3333"

	fmt.Print("Connecting.. ")
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		fmt.Printf("There were an error: %v", err)
		os.Exit(1)
	}
	fmt.Println("Done!")

	defer conn.Close()
	client := service.NewChittychatClient(conn)

	context, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Print("Sending message.. ")
	message := service.Message{Clock: clock, Message: "Hello, World!"}
	client.Publish(context, &message)
	fmt.Println("Done!")
	time.Sleep(5 * time.Second)
	fmt.Println("Ending now!")
}
