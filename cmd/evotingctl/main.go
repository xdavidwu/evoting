package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "github.com/xdavidwu/evoting/proto"
)

var (
	addr = flag.String("server", "localhost:1234", "server address")
	help = `Usage: %s [GLOBAL FLAGS]... SUBCOMMAND

Subcommands:
  register NAME GROUP PUBLIC_KEY_FILE
  unregister NAME

Global flags:
`
)

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), help, os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	connection, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("fail to connect to server: %v", err)
	}
	defer connection.Close()
	client := pb.NewRegistrationClient(connection)

	args := flag.Args()
	if len(args) < 1 {
		flag.Usage()
		log.Fatal("SUBCOMMAND not provided")
	}
	switch args[0] {
	case "register":
		if len(args) != 4 {
			flag.Usage()
			log.Fatal("Invalid numer of arguments for register")
		}
		key, err := os.ReadFile(args[3])
		if err != nil {
			log.Fatalf("fail to read public key: %v", err)
		}

		status, err := client.RegisterVoter(context.Background(), &pb.Voter{
			Name: &args[1],
			Group: &args[2],
			PublicKey: key,
		})
		if err != nil {
			log.Fatalf("fail to register: %v", err)
		}
		if err = pb.RegisterVoterToError(status); err != nil {
			log.Fatalf("fail to register: %v", err)
		}
	case "unregister":
		if len(args) != 2 {
			flag.Usage()
			log.Fatal("Invalid numer of arguments for unregister")
		}
		status, err := client.UnregisterVoter(context.Background(), &pb.VoterName{Name: &args[1]})
		if err != nil {
			log.Fatalf("fail to unregister: %v", err)
		}
		if err = pb.UnregisterVoterToError(status); err != nil {
			log.Fatalf("fail to unregister: %v", err)
		}
	default:
		log.Fatalf("unknown subcommand %s", args[0])
	}
}
