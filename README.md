# ITU Distributed Systems 2021, Mini Project 2

## Run the program

The necessary commands have been defined in the `Makefile` and can be run easily with `make`. If you are on Windows, `make` should be very easy to install. If you do not wish to do that, you can go into the `Makefile` and just execute the commands yourself.

To run the program, you can use these commands (which will all use Docker):

1. Build the Docker image: `make docker-image`
2. Build the gRPC protocul buffer files with `protoc` (inside Docker): `make docker-proto`
3. Run the server (with Docker): `make docker-server`
4. Run the client (with Docker): `make docker-client`

To avoid using Docker and just run the program outside of docker, use the non-Docker commands: `make proto`, `make server`, `make client`.
