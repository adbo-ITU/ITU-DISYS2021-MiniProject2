package main

import (
	"disysminiproject2/service"
	"fmt"
	"log"
	"sync"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// This is the struct that holds all global server state
type ChittyChatServer struct {
	service.UnimplementedChittychatServer

	clients   map[string]service.Chittychat_BroadcastServer // map from UID to broadcast stream
	usernames map[string]string                             // map from UID to username
	mutex     sync.Mutex
	clock     service.VectorClock

	// This map contains one channel for each user, and each channel is fed with
	// each new incoming message (i.e. this is how users receive broadcasts).
	manyIncomingMessages map[string]chan *service.UserMessage
}

// Initialises a new server instance
func NewServer() ChittyChatServer {
	return ChittyChatServer{
		clients:              make(map[string]service.Chittychat_BroadcastServer),
		usernames:            make(map[string]string),
		clock:                make(service.VectorClock),
		manyIncomingMessages: make(map[string](chan *service.UserMessage)),
	}
}

// Clients connect to this endpoint with a stream, in which they send their
// messages that they want published. The server then takes these messages and
// feeds them to each user's broadcast channel. Worth noting is that gRPC opens
// this function for each endpoint call (in a goroutine), and the function call
// is kept alive until the user disconnects again.
func (c *ChittyChatServer) Publish(stream service.Chittychat_PublishServer) error {
	// Cache the username, as the last msg received on client disconnect will be nil
	var cachedUsername string

	// We loop forever, listening for new messages from the currently connected
	// client.
	for {
		msg, err := stream.Recv()

		// Store the username of the publishing user
		if msg != nil {
			cachedUsername = msg.User
		}

		// The "Canceled" status code is sent when the client disconnects. We
		// handle users leaving the room by them simply disconnecting.
		if e, errOk := status.FromError(err); errOk && err != nil && e.Code() == codes.Canceled {
			log.Printf("[%v] User exited\n", cachedUsername)
			c.broadcastServiceMessage("", cachedUsername, service.UserMessage_DISCONNECT)
			c.removeClient(cachedUsername) //Remove client in case of disconnect
			return nil
		}
		// This handles any other error - we have never seen this get fired.
		if err != nil {
			log.Printf("[%v] Error on receive: %v\n", msg.User, err)
			c.broadcastServiceMessage("", msg.User, service.UserMessage_ERROR)
			c.removeClient(cachedUsername) //Remove client in case of client crash
			return err
		}

		// The vector clock is a shared resource, so we must lock it upon access
		// We merge the two clocks and increment the server's clock as per the
		// algorithm from the lecture.
		c.mutex.Lock()
		c.clock = service.MergeClocks(c.clock, msg.Message.Clock)
		c.mutex.Unlock()
		c.incrementOwnClock()

		// We log the current message
		fmtClock := service.FormatVectorClockAsString(c.clock)
		log.Printf("[%v] %s %s", msg.User, fmtClock, msg.Message.Content)

		// Broadcast the received message to all other clients (and in fact the
		// sender themselves)
		for _, channel := range c.manyIncomingMessages {
			channel <- msg
		}
	}
}

func (c *ChittyChatServer) Broadcast(_ *emptypb.Empty, stream service.Chittychat_BroadcastServer) error {
	// Each broadcast call will live in its own goroutine - serving one client each
	// https://github.com/grpc/grpc-go/blob/master/Documentation/concurrency.md#servers

	// We generate a new UUID for each client
	uid := uuid.Must(uuid.NewRandom()).String()
	log.Printf("New user joined: %s\n", uid)

	// Create the communications channel to receive incoming messages destined
	// for the belonging client
	c.manyIncomingMessages[uid] = make(chan *service.UserMessage, 20)

	// We expect the next message to be one where the new client sets their
	// username (as typed in by the user), but this is not reliable. That's
	// because a message could be sent between the user joining and the user
	// sending their username - an unfortunate consequence of bad communication
	// with lecturers and TAs that led to some needed (poorly done) refactoring.
	usernameMsg := <-c.manyIncomingMessages[uid]
	if usernameMsg.Event != service.UserMessage_SET_USERNAME {
		log.Printf("Failed to initialise user %s\n", uid)
	}
	username := usernameMsg.User
	log.Printf("%s was assigned with username %s\n", uid, username)

	// Same as in Publish
	c.mutex.Lock()
	c.clock = service.MergeClocks(c.clock, usernameMsg.Message.Clock) //Clients clock added alreadt at clients join
	c.mutex.Unlock()
	c.incrementOwnClock()

	// We officially add the new client to our global server state
	err := c.addClient(uid, username, stream)
	if err != nil {
		log.Printf("Client join error: %v\n", err)
	}
	defer c.removeClient(uid)

	// We notify all other users that a new client has joined
	c.broadcastServiceMessage("", username, service.UserMessage_JOIN)

	// For this specific client, forever listen to new broadcasts being sent to
	// its communication channel.
	for {
		msg := <-c.manyIncomingMessages[uid]

		// Upon a new message, again merge and increment clock
		c.mutex.Lock()
		c.clock = service.MergeClocks(c.clock, msg.Message.Clock)
		c.mutex.Unlock()
		c.incrementOwnClock()

		// Set the clock of the message to the updated clock, to avoid stale clocks being passed around
		msg.Message.Clock = c.clock

		// Actually send that message over the network to the connected client
		err = c.clients[uid].Send(msg)
		if err != nil {
			// Something must have gone wrong with sending, just drop the client
			return fmt.Errorf("something went wrong, dropping client: %v", err)
		}
	}
}

func (c *ChittyChatServer) addClient(id string, username string, conn service.Chittychat_BroadcastServer) error {
	// This touches global server state (shared resource), so we must lock
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if _, ok := c.clients[id]; ok {
		return fmt.Errorf("user id already exists: %s", id)
	}
	// We add the user into the relevant maps
	c.clients[id] = conn
	c.usernames[id] = username
	c.clock[username] = 0
	return nil
}

// Function to clean up the actions that addClient performed
func (c *ChittyChatServer) removeClient(id string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.clients, id)
	delete(c.clock, id)
	delete(c.usernames, id)
}

// Utility function to broadcast a message. It handles incrementing the server
// clock and writes the new message to all client channels.
func (c *ChittyChatServer) broadcastServiceMessage(content string, username string, event service.UserMessage_EventType) {
	// Note: this can't be rewritten at the momment, to also use the messaging channels to broadcast messages, as
	// some information is lost somehow, so clients don't register that they are told that someone joins or exits
	c.incrementOwnClock()
	message := service.UserMessage{Message: c.newMessage(content), User: username, Event: event}
	for _, channel := range c.manyIncomingMessages {
		channel <- &message
	}
}

// Utility function to make a Message struct (containing the current server
// clock) from a simple string.
func (c *ChittyChatServer) newMessage(content string) *service.Message {
	return &service.Message{Clock: c.clock, Content: content}
}

// Utility function to access the shared clock resource and increment the server
// clock.
func (c *ChittyChatServer) incrementOwnClock() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.clock["server"]++
}
