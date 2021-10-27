.PHONY: proto
proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./service/chittychat.proto;


.PHONY: run-client
run-client:
	go run client/*.go

.PHONY: run-server
run-server:
	go run server/*.go