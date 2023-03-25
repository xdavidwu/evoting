all: evoting-server evoting-client evotingctl

gen: proto/voting.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/voting.proto

evoting-server: gen
	go build ./cmd/evoting-server

evoting-client: gen
	go build ./cmd/evoting-client

evotingctl: gen
	go build ./cmd/evotingctl

.PHONY: gen evoting-server evoting-client evotingctl
