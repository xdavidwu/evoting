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

type clientState struct {
	client	pb.EVotingClient
	key	[]byte
	token	*pb.AuthToken
}

func obtainToken(s clientState) {
	vname := &pb.VoterName{Name: name}
	challenge, err := s.client.PreAuth(context.Background(), vname)
	if err != nil {
		log.Fatalf("fail to call PreAuth: %v", err)
	}
	// TODO sign
	token, err := s.client.Auth(context.Background(), &pb.AuthRequest{
		Name: vname,
		Response: &pb.Response{Value: challenge.Value},
	})
	s.token = token
	if err != nil {
		log.Fatalf("fail to call Auth: %v", err)
	}
}

func retryWithAuth[T interface{}](s clientState, f func(clientState) T, retries func(T) bool) T {
	v := f(s)
	if retries(v) {
		obtainToken(s)
		return f(s)
	}
	return v
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
	s := clientState{
		client: client,
		key: key,
	}
	obtainToken(s)

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
			var status *pb.Status
			retryWithAuth(s, func(s clientState) *pb.Status {
				status, err = s.client.CreateElection(context.Background(), &pb.Election{
					Token: s.token,
				})
				if err != nil {
					log.Fatalf("cannot create election: %v", err)
				}
				return status
			}, func(status *pb.Status) bool {
				return *status.Code == pb.CreateElectionUnauthn
			})
			if err = pb.CreateElectionToError(status); err != nil {
				log.Printf("fail to create election: %v", err)
			}
		case "vote":
			if len(args) != 3 {
				log.Println("Invalid number of arguments for vote")
				shellHelp(stdout)
				break
			}

			var status *pb.Status
			retryWithAuth(s, func(s clientState) *pb.Status {
				status, err := s.client.CastVote(context.Background(), &pb.Vote{
					ElectionName: &args[1],
					ChoiceName: &args[2],
					Token: s.token,
				})
				if err != nil {
					log.Fatalf("cannot cast vote: %v", err)
				}
				return status
			}, func(status *pb.Status) bool {
				return *status.Code == pb.CastVoteUnauthn
			})
			if err = pb.CastVoteToError(status); err != nil {
				log.Printf("fail to cast vote: %v", err)
			}
		case "result":
			if len(args) != 2 {
				log.Println("Invalid number of arguments for result")
				shellHelp(stdout)
				break
			}
			result, err := s.client.GetResult(context.Background(), &pb.ElectionName{Name: &args[1]})
			if err != nil {
				log.Fatalf("cannot query result: %v", err)
			}
			result, err = pb.GetResultToError(result)
			if err != nil {
				log.Printf("failed to query result: %v", err)
			}
			for _, r := range(result.Counts) {
				fmt.Printf("%s:\t%d", *r.ChoiceName, r.Count)
			}
		default:
			shellHelp(stdout)
		}
	}
exit:
}
