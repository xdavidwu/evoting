package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"flag"
	"log"
	"net"
	"os"
	"path"
	"github.com/jamesruan/sodium"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pb "github.com/xdavidwu/evoting/proto"
	"github.com/xdavidwu/evoting/store"
	_ "modernc.org/sqlite"
)

var (
	registAddr	= flag.String("registration-listen", "localhost:1234", "Listen address for registration")
	voteAddr	= flag.String("vote-listen", "0.0.0.0:5678", "Listen address for voting")
)

const (
	dbSchema = `CREATE TABLE IF NOT EXISTS 'users' ('name' TEXT PRIMARY KEY, 'group' TEXT);
CREATE TABLE IF NOT EXISTS 'challenges' ('id' INTEGER PRIMARY KEY AUTOINCREMENT, 'name' TEXT, 'value' TEXT);`
	challengeBytes = 16
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
	rows, err := s.db.Query("SELECT [group] FROM 'users' WHERE name = $1", v.Name)
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

func (s eVotingServer) PreAuth(_ context.Context, name *pb.VoterName) (*pb.Challenge, error) {
	rows, err := s.db.Query("SELECT 'group' FROM 'users' WHERE name = $1", name.Name)
	if err != nil {
		panic(err)
	}
	if !rows.Next() {
		return nil, status.Error(codes.Unauthenticated, "voter not registered") 
	}
	rows.Close()

	var c [challengeBytes]byte
	_, err = rand.Read(c[:])
	if err != nil {
		panic(err)
	}
	var challenge [challengeBytes * 2]byte
	hex.Encode(challenge[:], c[:])
	_, err = s.db.Exec("INSERT INTO 'challenges' ('name', 'value') VALUES ($1, $2)", name.Name, string(challenge[:]))
	if err != nil {
		panic(err)
	}

	return &pb.Challenge{Value: challenge[:]}, nil
}

func (s eVotingServer) Auth(_ context.Context, req *pb.AuthRequest) (*pb.AuthToken, error) {
	b, err := os.ReadFile(path.Join(s.keysDir, *req.Name.Name))
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "voter not registered")
	}
	key := sodium.SignPublicKey{Bytes: sodium.Bytes(b)}
	rows, err := s.db.Query("SELECT value FROM 'challenges' WHERE name = $1", req.Name.Name)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for rows.Next() {
		var c string
		err = rows.Scan(&c)
		if err != nil {
			panic(err)
		}
		m := sodium.Bytes([]byte(c))
		err = m.SignVerifyDetached(sodium.Signature{Bytes: req.Response.Value}, key)
		if err == nil {
			return &pb.AuthToken{Value: []byte{}}, nil
		}
	}

	return nil, status.Error(codes.Unauthenticated, "unknown signature")
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
