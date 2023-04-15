package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
	"github.com/chzyer/readline"
	"github.com/jamesruan/sodium"
	pb "github.com/xdavidwu/evoting/proto"
	"github.com/xdavidwu/evoting/store"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	addr	= flag.String("server", "localhost:5678", "Server address")
	keyFile	= flag.String("key", path.Join(store.ClientDataDir(), "key"), "Secret key file")
	name	= flag.String("name", "foo", "Voter name")
)

const (
	usage	= `Usage: %s [FLAGS...] [keygen]

keygen: Generate key pair and exit

Flags:
`
	shellUsage	= `Commands:
  create NAME:             Create an election
  vote ELECTION NAME: Vote for NAME on ELECTION
  result ELECTION:    Query ELECTION result
  exit, quit, q:      Exit
`
	shellPrompt	= "evoiting> "
)

type clientState struct {
	client	pb.EVotingClient
	key	sodium.SignSecretKey
	token	*pb.AuthToken
}

func obtainToken(s *clientState) {
	vname := &pb.VoterName{Name: name}
	challenge, err := s.client.PreAuth(context.Background(), vname)
	if err != nil {
		log.Fatalf("fail to call PreAuth: %v", err)
	}

	m := sodium.Bytes(challenge.Value)
	sig := m.SignDetached(s.key)

	token, err := s.client.Auth(context.Background(), &pb.AuthRequest{
		Name: vname,
		Response: &pb.Response{Value: sig.Bytes},
	})
	s.token = token
	if err != nil {
		log.Fatalf("fail to call Auth: %v", err)
	}
}

func retryWithAuth[T interface{}](s clientState, f func(clientState) T, retries func(T) bool) T {
	v := f(s)
	if retries(v) {
		obtainToken(&s)
		return f(s)
	}
	return v
}

func ask(l *readline.Instance, prompt string) string {
	l.HistoryDisable()
	l.SetPrompt(prompt)
	res, err := l.Readline()
	if err != nil {
		panic(err)
	}
	l.SetPrompt(shellPrompt)
	l.HistoryEnable()
	return strings.TrimSpace(res)
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), usage, os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	args := flag.Args()
	if len(args) == 1 && args[0] == "keygen" {
		dataDir := store.ClientDataDir()
		err := os.MkdirAll(dataDir, 0700)
		if err != nil {
			log.Printf("Unable to create data dir %s: %v", dataDir, err)
		}

		kp := sodium.MakeSignKP()

		err = os.WriteFile(*keyFile, kp.SecretKey.Bytes, 0600)
		if err != nil {
			log.Fatalf("Unable to write secret key: %v", err)
		}

		pub := *keyFile + ".pub"
		err = os.WriteFile(pub, kp.PublicKey.Bytes, 0600)
		if err != nil {
			log.Fatalf("Unable to write public key: %v", err)
		}

		log.Printf("Key pair generated at %s, %s", *keyFile, pub)
		return
	}

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
		key: sodium.SignSecretKey{
			Bytes: key,
		},
	}
	obtainToken(&s)

	l, err := readline.New(shellPrompt)
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
			if len(args) != 2 {
				log.Println("Invalid number of arguments for create")
				fmt.Fprint(stdout, shellUsage)
				break
			}

			// TODO input
			var t time.Time
			for {
				timeStr := ask(l, "ending time (format as in " + time.DateTime + "): ")
				t, err = time.ParseInLocation(time.DateTime, timeStr, time.Local)
				if err == nil {
					break
				}
			}

			var ng int
			for {
				ngStr := ask(l, "number of groups to allow: ")
				ng, err = strconv.Atoi(ngStr)
				if err == nil {
					break
				}
			}

			groups := make([]string, ng)
			for i := 0; i < ng; i++ {
				groups[i] = ask(l, " group name: ")
			}

			var nc int
			for {
				ncStr := ask(l, "number of choices: ")
				nc, err = strconv.Atoi(ncStr)
				if err == nil {
					break
				}
			}

			choices := make([]string, nc)
			for i := 0; i < nc; i++ {
				choices[i] = ask(l, " choice: ")
			}

			status := retryWithAuth(s, func(s clientState) *pb.Status {
				status, err := s.client.CreateElection(context.Background(), &pb.Election{
					Name: &args[1],
					Groups: groups,
					Choices: choices,
					EndDate: timestamppb.New(t),
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
				fmt.Fprint(stdout, shellUsage)
				break
			}

			status := retryWithAuth(s, func(s clientState) *pb.Status {
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
				fmt.Fprint(stdout, shellUsage)
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
			for _, r := range result.Counts {
				fmt.Printf("%s:\t%d", *r.ChoiceName, r.Count)
			}
		default:
			fmt.Fprint(stdout, shellUsage)
		}
	}
exit:
}
