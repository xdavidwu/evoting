package main

import (
	"context"
	"flag"
	"log"
	"net"
	"google.golang.org/grpc"
	pb "github.com/xdavidwu/evoting/proto"
)

var (
	registAddr	= flag.String("registration-listen", "localhost:1234", "Listen address for registration")
	voteAddr	= flag.String("vote-listen", "0.0.0.0:5678", "Listen address for voting")
)

type registrationServer struct {
	pb.UnimplementedRegistrationServer
}

type eVotingServer struct {
	pb.UnimplementedEVotingServer
}

func (eVotingServer) PreAuth(_ context.Context, name *pb.VoterName) (*pb.Challenge, error) {
	log.Printf("PreAuth: %s", *name.Name)
	return &pb.Challenge{Value: []byte{}}, nil
}

func (eVotingServer) Auth(_ context.Context, req *pb.AuthRequest) (*pb.AuthToken, error) {
	log.Printf("Auth: %s", *req.Name.Name)
	return &pb.AuthToken{Value: []byte{}}, nil
}

func main() {
	flag.Parse()

	registLn, err := net.Listen("tcp", *registAddr)
	if err != nil {
		log.Fatalf("failed to listen %s: %v", *registAddr, err)
	}
	voteLn, err := net.Listen("tcp", *voteAddr)
	if err != nil {
		log.Fatalf("failed to listen %s: %v", *voteAddr, err)
	}

	registServer := grpc.NewServer()
	voteServer := grpc.NewServer()

	pb.RegisterRegistrationServer(registServer, &registrationServer{})
	pb.RegisterEVotingServer(voteServer, &eVotingServer{})

	go registServer.Serve(registLn)
	voteServer.Serve(voteLn)
}
