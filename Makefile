.PHONY: proto
proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./service/chittychat.proto;


.PHONY: client
client:
	go run client/*.go

.PHONY: server
server:
	go run server/*.go
