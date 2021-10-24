package main

import (
	"context"
	"disysminiproject2/service"
	"fmt"
	"log"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/stats"
)

var (
	clock      = make(map[string]uint64)
	serverData = ChittyChatServer{}
)

// at holde høje om en client disconnecter
// https://stackoverflow.com/questions/39825671/grpc-go-how-to-know-in-server-side-when-client-closes-the-connection

func unaryInterceptor(ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (interface{}, error) {
	// This function is fired when a new connection comes, but before the client gets their function call processed
	// https://blog.gopheracademy.com/advent-2017/go-grpc-beyond-basics/
	log.Println("New connection received")

	return handler(ctx, req)
}

func disconnectDetector(input <-chan struct{}) {
	for {
		select {
		case <-input:
			fmt.Println("Client disconnected")
		}
	}

}

// handler, statshandler stuff from https://github.com/grpc/grpc-go/issues/3634
// used to detect when client disconnects
type Handler struct {
}

func (h *Handler) TagRPC(context.Context, *stats.RPCTagInfo) context.Context {
	log.Println("TagRPC")
	return context.Background()
}

// HandleRPC processes the RPC stats.
func (h *Handler) HandleRPC(context.Context, stats.RPCStats) {
	log.Println("HandleRPC")
}

func (h *Handler) TagConn(context.Context, *stats.ConnTagInfo) context.Context {

	log.Println("Tag Conn")
	return context.Background()
}

// HandleConn processes the Conn stats.
func (h *Handler) HandleConn(c context.Context, s stats.ConnStats) {
	switch s.(type) {
	case *stats.ConnEnd:
		log.Println("get connEnd")
		//fmt.Printf("client %d disconnected", s.userIdMap[ctx.Value("user_counter")])
		break
	}
}

func main() {
	listener, err := net.Listen("tcp", "localhost:3333")
	if err != nil {
		fmt.Printf("Error while attempting to listen on port 3333: %v", err)
		os.Exit(1)
	}

	var options []grpc.ServerOption
	options = append(options, grpc.UnaryInterceptor(unaryInterceptor)) // used to catch when clients connect
	options = append(options, grpc.StatsHandler(&Handler{}))           // used when clients disconnects

	log.Println("Starting server")

	server := grpc.NewServer(options...)
	service.RegisterChittychatServer(server, serverData)
	server.Serve(listener)
}
