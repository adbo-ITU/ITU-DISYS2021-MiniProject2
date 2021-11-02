image = disys-miniproject2-group-pointofgoreturn
run_container = docker run -it --rm -v '$(CURDIR):/app' --network="host" $(image)

.PHONY: docker-image
docker-image:
	docker build -t $(image) .

.PHONY: docker-proto
docker-proto:
	$(run_container) bash -c 'make proto'

.PHONY: docker-server
docker-server:
	$(run_container) bash -c 'make server'

.PHONY: docker-client
docker-client:
	$(run_container) bash -c 'make client'

.PHONY: proto
proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./service/chittychat.proto;

.PHONY: client
client:
	go run client/*.go

.PHONY: server
server:
	go run server/*.go
