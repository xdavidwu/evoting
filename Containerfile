ARG ALPINE=3.17

FROM alpine:$ALPINE as build
RUN apk add make go protoc protobuf-dev
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28 &&\
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
COPY . evoting
RUN env PATH="$PATH:$(go env GOPATH)/bin" make all -C evoting

FROM alpine:$ALPINE as evotingctl
COPY --from=build evoting/evotingctl /usr/local/bin
CMD /usr/local/bin/evotingctl

FROM alpine:$ALPINE as evoting-client
COPY --from=build evoting/evoting-client /usr/local/bin
CMD /usr/local/bin/evoting-client

FROM alpine:$ALPINE as evoting-server
COPY --from=build evoting/evoting-server /usr/local/bin
CMD /usr/local/bin/evoting-server
