package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"github.com/chzyer/readline"
	pb "github.com/xdavidwu/evoting/proto"
)

var (
	addr	= flag.String("server", "localhost:5678", "Server address")
	keyFile	= flag.String("key", "/dev/null", "Secret key file")
	name	= flag.String("name", "foo", "Voter name")
)

func obtainToken(client pb.EVotingClient, _ []byte) (*pb.AuthToken, error) {
	vname := &pb.VoterName{Name: name}
	challenge, err := client.PreAuth(context.Background(), vname)
	if err != nil {
		return nil, err
	}
	// TODO sign
	return client.Auth(context.Background(), &pb.AuthRequest{
		Name: vname,
		Response: &pb.Response{Value: challenge.Value},
	})
}

func shellHelp(out io.Writer) {
	fmt.Fprint(out, `Commands:
  create:             Create an election
  vote ELECTION NAME: Vote for NAME on ELECTION
  result ELECTION:    Query ELECTION result
  exit, quit, q:      Exit`)
}

func main() {
	flag.Parse()

	key, err := os.ReadFile(*keyFile)
	if err != nil {
		log.Fatalf("Unable to read secret key: %v", err)
	}

	connection, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("fail to connect to server: %v", err)
	}
	defer connection.Close()
	client := pb.NewEVotingClient(connection)

	token, err := obtainToken(client, key)
	if err != nil {
		log.Fatalf("cannot authenticate: %v", err)
	}

	l, err := readline.New("evoting> ")
	if err != nil {
		panic(err)
	}
	defer l.Close()

	log.SetOutput(l.Stderr())

	stdout := l.Stdout()

	for {
		line, err := l.Readline()
		if err != nil {
			if err == readline.ErrInterrupt || err == io.EOF {
				break
			}
			panic(err)
		}

		args := strings.Split(strings.TrimSpace(line), " ")
		if len(args) == 0 {
			continue
		}

		switch args[0] {
		case "exit":
			goto exit
		case "quit":
			goto exit
		case "q":
			goto exit
		case "create":
			// TODO input
			status, err := client.CreateElection(context.Background(), &pb.Election{
				Token: token,
			})
			if err != nil {
				log.Fatalf("cannot create election: %v", err)
			}
			log.Printf("Create: %d", status.Code)

		case "vote":
			if len(args) != 3 {
				log.Println("Invalid number of arguments for vote")
				shellHelp(stdout)
				break
			}
			status, err := client.CastVote(context.Background(), &pb.Vote{
				ElectionName: &args[1],
				ChoiceName: &args[2],
				Token: token,
			})
			if err != nil {
				log.Fatalf("cannot cast vote: %v", err)
			}
			log.Printf("Vote: %d", status.Code)

		case "result":
			if len(args) != 2 {
				log.Println("Invalid number of arguments for result")
				shellHelp(stdout)
				break
			}
			result, err := client.GetResult(context.Background(), &pb.ElectionName{Name: &args[1]})
			if err != nil {
				log.Fatalf("cannot query result: %v", err)
			}
			// TODO check status
			for _, r := range(result.Counts) {
				fmt.Printf("%s:\t%d", *r.ChoiceName, r.Count)
			}

		default:
			shellHelp(stdout)
		}
	}
exit:
}
