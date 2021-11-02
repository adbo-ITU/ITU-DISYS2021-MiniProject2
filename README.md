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

TBA.

### Example client log (user "Alpha")

TBA.

### Example client log (user "Bravo")

TBA.
