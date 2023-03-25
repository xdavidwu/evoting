DOCKER := docker

all: evoting-server evoting-client evotingctl

gen: proto/voting.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/voting.proto

evoting-server: gen
	go build ./cmd/evoting-server

evoting-client: gen
	go build ./cmd/evoting-client

evotingctl: gen
	go build ./cmd/evotingctl

containers:
	$(DOCKER) build -f Containerfile -t evotingctl --target evotingctl .
	$(DOCKER) build -f Containerfile -t evoting-server --target evoting-server .
	$(DOCKER) build -f Containerfile -t evoting-client --target evoting-client .

.PHONY: gen evoting-server evoting-client evotingctl containers
