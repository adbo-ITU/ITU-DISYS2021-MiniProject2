# ITU Distributed Systems 2021, Mini Project 2

## Run the program

The necessary commands have been defined in the `Makefile` and can be run easily with `make`. If you are on Windows, `make` should be very easy to install. If you do not wish to do that, you can go into the `Makefile` and just execute the commands yourself.

To run the program, you can use these commands (which will all use Docker):

1. Build the Docker image: `make docker-image`
2. Build the gRPC protocul buffer files with `protoc` (inside Docker): `make docker-proto` (likely not necessary, because the files are already in the repository)
3. Run the server (with Docker): `make docker-server`
4. Run the client (with Docker): `make docker-client`

To avoid using Docker and just run the program outside of docker, use the non-Docker commands: `make proto`, `make server`, `make client`.

## Using the program

### Server

The server does not need you to do anything other than start it.

### Client

To join the chat room, start the client and enter your username on demand (no whitespace allowed):

```text
Connecting.. Done!
Please enter your wanted username:
```

Once it as been entered (click `Enter`), you should see a chat window saying you have joined:

![Picture of join message and GUI](https://i.imgur.com/hHhlA6S.png)

The purple box shows all received messages. The blue box is your input field, limited to at maximum 128 characters. To type, simply type with normal letters, punctuation, spaces etc. To delete the latest character, use backspace. To send the message, press `Enter`. Example of sent message and a message being typed:

![Picture of typing and sending messages](https://i.imgur.com/VOimqAj.png)

The messages follow this format:

```text
[USER WHO SENT GOES HERE] <VECTOR CLOCK GOES HERE>
MESSAGE GOES HERE
```

You can join with different clients, and you can see that the vector clock and usernames should respond appropriately.

In order to scroll up and down in the messages box, simply use the UP and DOWN arrow keys. You *may* have to click it multiple times before anything happens due to a bug.

## Logs

The server writes its logs to stdout. The client uses the terminal for a GUI, so the logs are written to a file instead - this file is called `client.log`.

The following three subsections show log output for a chat session where two users, *Alpha* and *Bravo*, chat.

### Example server log

```text
2021/11/02 12:51:23 Started server
2021/11/02 12:51:28 New user joined: 609d93af-4f59-4bb2-8aff-fe3015b0fedc
2021/11/02 12:51:28 [A] <A: 1, server: 1> 
2021/11/02 12:51:28 609d93af-4f59-4bb2-8aff-fe3015b0fedc was assigned with username A
2021/11/02 12:51:35 New user joined: 07112cb1-4d4f-48c0-a6e5-4aee777369de
2021/11/02 12:51:35 [B] <A: 0, B: 1, server: 5> 
2021/11/02 12:51:35 07112cb1-4d4f-48c0-a6e5-4aee777369de was assigned with username B
2021/11/02 12:51:42 New user joined: 8e637654-a716-407b-8c61-106b4a0a4f56
2021/11/02 12:51:42 [C] <A: 0, B: 0, C: 1, server: 11> 
2021/11/02 12:51:42 8e637654-a716-407b-8c61-106b4a0a4f56 was assigned with username C
2021/11/02 12:51:46 [A] <A: 7, B: 1, C: 1, server: 19> Hello
2021/11/02 12:51:55 New user joined: faf44288-a837-4243-9ef4-79f6495f72ef
2021/11/02 12:51:55 [D] <A: 7, B: 1, C: 1, D: 1, server: 23> 
2021/11/02 12:51:55 faf44288-a837-4243-9ef4-79f6495f72ef was assigned with username D
2021/11/02 12:52:02 [B] User exited
2021/11/02 12:52:14 [D] <A: 7, C: 1, D: 4, server: 38> Hello hello
2021/11/02 12:52:21 [A] User exited
```

### Example client log (user "Alpha")

```text
2021/11/02 12:51:28 Starting the system
2021/11/02 12:51:28 <A: 1>
set username to A
2021/11/02 12:51:28 Starting to listen for messages from the chat server
2021/11/02 12:51:28 Starting to listen for chat messages from the server
2021/11/02 12:51:28 Client received message: Participant A joined Chitty-Chat at Lamport time <A: 2, server: 4>
2021/11/02 12:51:28 Participant A joined Chitty-Chat at Lamport time <A: 2, server: 4>
2021/11/02 12:51:35 Client received message: <A: 3, B: 1, server: 6>
set username to B
2021/11/02 12:51:35 Client received message: Participant B joined Chitty-Chat at Lamport time <A: 4, B: 1, server: 10>
2021/11/02 12:51:35 Participant B joined Chitty-Chat at Lamport time <A: 4, B: 1, server: 10>
2021/11/02 12:51:42 Client received message: <A: 5, B: 1, C: 1, server: 12>
set username to C
2021/11/02 12:51:42 Client received message: Participant C joined Chitty-Chat at Lamport time <A: 6, B: 1, C: 1, server: 18>
2021/11/02 12:51:42 Participant C joined Chitty-Chat at Lamport time <A: 6, B: 1, C: 1, server: 18>
2021/11/02 12:51:46 Received UI event for <Enter>
2021/11/02 12:51:46 Sending message to sever with contents: Hello, clock: <A: 7, B: 1, C: 1, server: 18>
2021/11/02 12:51:46 Client received message: [A] <A: 8, B: 1, C: 1, server: 22>
Hello
2021/11/02 12:51:46 [A] <A: 8, B: 1, C: 1, server: 22>
Hello
2021/11/02 12:51:55 Client received message: <A: 9, B: 1, C: 1, D: 1, server: 27>
set username to D
2021/11/02 12:51:55 Client received message: Participant D joined Chitty-Chat at Lamport time <A: 10, B: 1, C: 1, D: 1, server: 29>
2021/11/02 12:51:55 Participant D joined Chitty-Chat at Lamport time <A: 10, B: 1, C: 1, D: 1, server: 29>
2021/11/02 12:52:02 Client received message: Participant B left Chitty-Chat at Lamport time <A: 11, B: 1, C: 1, D: 1, server: 36>
2021/11/02 12:52:02 Participant B left Chitty-Chat at Lamport time <A: 11, C: 1, D: 1, server: 36>
2021/11/02 12:52:14 Client received message: [D] <A: 12, C: 1, D: 4, server: 40>
Hello hello
2021/11/02 12:52:14 [D] <A: 12, C: 1, D: 4, server: 40>
Hello hello
2021/11/02 12:52:21 Received UI event for program exit

```

### Example client log (user "Bravo")

```text
2021/11/02 12:51:35 Starting the system
2021/11/02 12:51:35 <B: 1>
set username to B
2021/11/02 12:51:35 Starting to listen for messages from the chat server
2021/11/02 12:51:35 Client received message: Participant B joined Chitty-Chat at Lamport time <A: 0, B: 2, server: 10>
2021/11/02 12:51:35 Starting to listen for chat messages from the server
2021/11/02 12:51:35 Participant B joined Chitty-Chat at Lamport time <A: 0, B: 2, server: 10>
2021/11/02 12:51:42 Client received message: <A: 0, B: 3, C: 1, server: 13>
set username to C
2021/11/02 12:51:42 Client received message: Participant C joined Chitty-Chat at Lamport time <A: 0, B: 4, C: 1, server: 17>
2021/11/02 12:51:42 Participant C joined Chitty-Chat at Lamport time <A: 0, B: 4, C: 1, server: 17>
2021/11/02 12:51:46 Client received message: [A] <A: 7, B: 5, C: 1, server: 22>
Hello
2021/11/02 12:51:46 [A] <A: 7, B: 5, C: 1, server: 22>
Hello
2021/11/02 12:51:55 Client received message: <A: 7, B: 6, C: 1, D: 1, server: 24>
set username to D
2021/11/02 12:51:55 Client received message: Participant D joined Chitty-Chat at Lamport time <A: 7, B: 7, C: 1, D: 1, server: 32>
2021/11/02 12:51:55 Participant D joined Chitty-Chat at Lamport time <A: 7, B: 7, C: 1, D: 1, server: 32>
2021/11/02 12:52:02 Received UI event for program exit
```

### Example client log (user "Charlie")

```text
2021/11/02 12:51:42 Starting the system
2021/11/02 12:51:42 <C: 1>
set username to C
2021/11/02 12:51:42 Starting to listen for messages from the chat server
2021/11/02 12:51:42 Client received message: Participant C joined Chitty-Chat at Lamport time <A: 0, B: 0, C: 2, server: 17>
2021/11/02 12:51:42 Starting to listen for chat messages from the server
2021/11/02 12:51:42 Participant C joined Chitty-Chat at Lamport time <A: 0, B: 0, C: 2, server: 17>
2021/11/02 12:51:46 Client received message: [A] <A: 7, B: 1, C: 3, server: 20>
Hello
2021/11/02 12:51:46 [A] <A: 7, B: 1, C: 3, server: 20>
Hello
2021/11/02 12:51:55 Client received message: <A: 7, B: 1, C: 4, D: 1, server: 28>
set username to D
2021/11/02 12:51:55 Client received message: Participant D joined Chitty-Chat at Lamport time <A: 7, B: 1, C: 5, D: 1, server: 31>
2021/11/02 12:51:55 Participant D joined Chitty-Chat at Lamport time <A: 7, B: 1, C: 5, D: 1, server: 31>
2021/11/02 12:52:02 Client received message: Participant B left Chitty-Chat at Lamport time <A: 7, B: 1, C: 6, D: 1, server: 37>
2021/11/02 12:52:02 Participant B left Chitty-Chat at Lamport time <A: 7, C: 6, D: 1, server: 37>
2021/11/02 12:52:14 Client received message: [D] <A: 7, C: 7, D: 4, server: 41>
Hello hello
2021/11/02 12:52:14 [D] <A: 7, C: 7, D: 4, server: 41>
Hello hello
2021/11/02 12:52:21 Client received message: Participant A left Chitty-Chat at Lamport time <A: 7, C: 8, D: 4, server: 43>
2021/11/02 12:52:21 Participant A left Chitty-Chat at Lamport time <C: 8, D: 4, server: 43>
```

### Example client log (user "Delta")

```text
2021/11/02 12:51:55 Starting the system
2021/11/02 12:51:55 <D: 1>
set username to D
2021/11/02 12:51:55 Starting to listen for messages from the chat server
2021/11/02 12:51:55 Client received message: Participant D joined Chitty-Chat at Lamport time <A: 7, B: 1, C: 1, D: 2, server: 31>
2021/11/02 12:51:55 Starting to listen for chat messages from the server
2021/11/02 12:51:55 Participant D joined Chitty-Chat at Lamport time <A: 7, B: 1, C: 1, D: 2, server: 31>
2021/11/02 12:52:02 Client received message: Participant B left Chitty-Chat at Lamport time <A: 7, B: 1, C: 1, D: 3, server: 35>
2021/11/02 12:52:02 Participant B left Chitty-Chat at Lamport time <A: 7, C: 1, D: 3, server: 35>
2021/11/02 12:52:14 Received UI event for <Enter>
2021/11/02 12:52:14 Sending message to sever with contents: Hello hello, clock: <A: 7, C: 1, D: 4, server: 35>
2021/11/02 12:52:14 Client received message: [D] <A: 7, C: 1, D: 5, server: 39>
Hello hello
2021/11/02 12:52:14 [D] <A: 7, C: 1, D: 5, server: 39>
Hello hello
2021/11/02 12:52:21 Client received message: Participant A left Chitty-Chat at Lamport time <A: 7, C: 1, D: 6, server: 44>
2021/11/02 12:52:21 Participant A left Chitty-Chat at Lamport time <C: 1, D: 6, server: 44>

```