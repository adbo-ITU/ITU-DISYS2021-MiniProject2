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
2021/11/02 12:22:12 Started serverd
2021/11/02 12:22:17 New user joined: 260364a4-b23a-4f98-aacc-0cdb376c0c04
2021/11/02 12:22:17 [A] <A: 1, server: 1> 
2021/11/02 12:22:17 260364a4-b23a-4f98-aacc-0cdb376c0c04 was assigned with username A
2021/11/02 12:22:23 New user joined: ecf566ce-7ab2-4b8b-a281-e00db0420a43
2021/11/02 12:22:23 [B] <A: 0, B: 1, server: 5> 
2021/11/02 12:22:23 ecf566ce-7ab2-4b8b-a281-e00db0420a43 was assigned with username B
2021/11/02 12:22:28 New user joined: 6924f5a8-7b5e-44f5-a767-d78e589a9cc7
2021/11/02 12:22:28 [C] <A: 0, B: 0, C: 1, server: 11> 
2021/11/02 12:22:28 6924f5a8-7b5e-44f5-a767-d78e589a9cc7 was assigned with username C
2021/11/02 12:22:33 New user joined: 0d294bb0-92a9-425b-a6a0-ce4ba22791d0
2021/11/02 12:22:33 [D] <A: 0, B: 0, C: 1, D: 1, server: 19> 
2021/11/02 12:22:33 0d294bb0-92a9-425b-a6a0-ce4ba22791d0 was assigned with username D
2021/11/02 12:22:39 [A] <A: 9, B: 1, C: 1, D: 1, server: 29> Hello
2021/11/02 12:22:45 [B] User exited
2021/11/02 12:22:52 [C] <A: 9, C: 7, D: 1, server: 39> Hello hello
2021/11/02 12:22:54 [A] User exited
```

### Example client log (user "Alpha")

TBA.

### Example client log (user "Bravo")

TBA.

### Example client log (user "Charlie")

TBA.
