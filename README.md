# evoting

## Build

### Container-based builds

Build in containers, to containers.

Requirements:

- make
- docker or podman

```
make containers
```

This will build three containers, {evotingctl,evoting-server,evoting-client}.

To use podman instead of docker:

```
make containers DOCKER=podman
```

### Native builds

Requirements:

- make
- go
- protoc
- google/protobuf/timestamp.proto
	- On Alpine Linux, this is in protobuf-dev package
- gRPC/ Protobuf Go tools
	- `protoc-gen-go`, `protoc-gen-go-grpc` in PATH
	- To install: `go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28 && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2`
	- Make sure GOPATH is in PATH

```
make all
```

This will produce {evotingctl,evoting-server,evoting-client} binaries.
