package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"net"
	"os"
	"path"
	"google.golang.org/grpc"
	pb "github.com/xdavidwu/evoting/proto"
	"github.com/xdavidwu/evoting/store"
	_ "modernc.org/sqlite"
)

var (
	registAddr	= flag.String("registration-listen", "localhost:1234", "Listen address for registration")
	voteAddr	= flag.String("vote-listen", "0.0.0.0:5678", "Listen address for voting")
)

const (
	dbSchema = `CREATE TABLE IF NOT EXISTS 'users' ('name' TEXT PRIMARY KEY, 'group' TEXT);`
)

type registrationServer struct {
	pb.UnimplementedRegistrationServer
	keysDir string
	db *sql.DB
}

func (s registrationServer) RegisterVoter(_ context.Context, v *pb.Voter) (*pb.Status, error) {
	status := pb.RegisterVoterSuccess
	_, err := s.db.Exec("INSERT INTO 'users' ('name', 'group') VALUES ($1, $2)", v.Name, v.Group)
	if err != nil {
		status = pb.RegisterVoterExists
	} else {
		os.WriteFile(path.Join(s.keysDir, *v.Name), v.PublicKey, 0600)
	}
	return &pb.Status{Code: &status}, nil
}

func (s registrationServer) UnregisterVoter(_ context.Context, v *pb.VoterName) (*pb.Status, error) {
	rows, err := s.db.Query("SELECT 'group' FROM 'users' WHERE name = $1", v.Name)
	if err != nil {
		panic(err)
	}
	if !rows.Next() {
		rows.Close()
		status := pb.UnregisterVoterNotFound
		return &pb.Status{Code: &status}, nil
	}
	rows.Close()

	os.Remove(path.Join(s.keysDir, *v.Name))
	s.db.Exec("DELETE FROM 'users' WHERE name = $1", v.Name)
	status := pb.UnregisterVoterSuccess
	return &pb.Status{Code: &status}, nil
}

type eVotingServer struct {
	pb.UnimplementedEVotingServer
	keysDir string
	db *sql.DB
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

	dataDir := store.ServerDataDir()
	err := os.MkdirAll(dataDir, 0700)
	if err != nil {
		log.Fatalf("failed to create data dir %s: %v", dataDir, err)
	}

	keysDir := path.Join(dataDir, "keys")
	err = os.MkdirAll(keysDir, 0700)
	if err != nil {
		log.Fatalf("failed to create keys dir %s: %v", keysDir, err)
	}

	dbPath := path.Join(dataDir, "db.sqlite")
	db, err := sql.Open("sqlite", dbPath)
	defer db.Close()
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	_, err = db.Exec(dbSchema)
	if err != nil {
		log.Printf("cannot init db: %v", err)
	}

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

	pb.RegisterRegistrationServer(registServer, &registrationServer{keysDir: keysDir, db: db})
	pb.RegisterEVotingServer(voteServer, &eVotingServer{keysDir: keysDir, db: db})

	go registServer.Serve(registLn)
	voteServer.Serve(voteLn)
}
