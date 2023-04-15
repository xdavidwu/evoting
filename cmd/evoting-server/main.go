package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net"
	"os"
	"path"
	"time"
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
CREATE TABLE IF NOT EXISTS 'challenges' ('id' INTEGER PRIMARY KEY AUTOINCREMENT, 'name' TEXT, 'value' TEXT);
CREATE TABLE IF NOT EXISTS 'elections' ('id' INTEGER PRIMARY KEY AUTOINCREMENT, 'name' TEXT UNIQUE, 'end_date' TEXT);
CREATE TABLE IF NOT EXISTS 'election_groups' ('id' INTEGER PRIMARY KEY AUTOINCREMENT, 'election_id' INTEGER, 'group' TEXT, FOREIGN KEY('election_id') REFERENCES elections('id'));
CREATE TABLE IF NOT EXISTS 'election_choices' ('id' INTEGER PRIMARY KEY AUTOINCREMENT, 'election_id' INTEGER, 'choice' TEXT, 'votes' INTEGER DEFAULT 0, FOREIGN KEY('election_id') REFERENCES elections('id'));
CREATE TABLE IF NOT EXISTS 'election_voted' ('id' INTEGER PRIMARY KEY AUTOINCREMENT, 'election_id' INTEGER, 'user' TEXT, FOREIGN KEY('election_id') REFERENCES elections('id'))`
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
	key sodium.SignKP
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

type token struct {
	Sub	string
	Exp	time.Time
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

	for rows.Next() {
		var c string
		err = rows.Scan(&c)
		if err != nil {
			panic(err)
		}
		m := sodium.Bytes([]byte(c))
		err = m.SignVerifyDetached(sodium.Signature{Bytes: req.Response.Value}, key)
		if err == nil {
			rows.Close()
			s.db.Exec("DELETE FROM 'challenges' WHERE value = $1", c)
			j, _ := json.Marshal(token{Sub: *req.Name.Name, Exp: time.Now().Add(time.Hour)})
			tok := sodium.Bytes(j)
			token := tok.Sign(s.key.SecretKey)
			return &pb.AuthToken{Value: token}, nil
		}
	}
	rows.Close()

	return nil, status.Error(codes.Unauthenticated, "unknown signature")
}

func (s eVotingServer) verifyToken(t *pb.AuthToken) (string, error) {
	m := sodium.Bytes(t.Value)
	b, err := m.SignOpen(s.key.PublicKey)
	if err != nil {
		return "", err
	}
	var token token
	err = json.Unmarshal(b, &token)
	if err != nil {
		return "", err
	}
	if time.Now().Before(token.Exp) {
		return token.Sub, nil
	}
	return "", errors.New("token expired")
}

func (s eVotingServer) CreateElection(_ context.Context, e *pb.Election) (*pb.Status, error) {
	_, err := s.verifyToken(e.Token)
	if err != nil {
		status := pb.CreateElectionUnauthn
		return &pb.Status{Code: &status}, nil
	}

	if len(e.Choices) == 0 || len(e.Groups) == 0 {
		status := pb.CreateElectionNoSpec
		return &pb.Status{Code: &status}, nil
	}

	dateBytes, err := e.EndDate.AsTime().MarshalText()
	if err != nil {
		panic(err)
	}
	rows, err := s.db.Query("INSERT INTO 'elections' ('name', 'end_date') VALUES ($1, $2); SELECT last_insert_rowid()", e.Name, string(dateBytes))
	if err != nil {
		status := pb.CreateElectionUnknown
		return &pb.Status{Code: &status}, nil
	}
	var id int64
	rows.Scan(&id)
	rows.Close()

	for _, g := range(e.Groups) {
		_, err = s.db.Exec("INSERT INTO 'election_groups' ('election_id', 'group') VALUES ($1, $2)", id, g)
		if err != nil {
			panic(err)
		}
	}

	for _, c := range(e.Choices) {
		_, err = s.db.Exec("INSERT INTO 'election_choices' ('election_id', 'choice') VALUES ($1, $2)", id, c)
		if err != nil {
			panic(err)
		}
	}

	status := pb.CreateElectionSuccess
	return &pb.Status{Code: &status}, nil
}

func (s eVotingServer) CastVote(_ context.Context, v *pb.Vote) (*pb.Status, error) {
	user, err := s.verifyToken(v.Token)
	if err != nil {
		status := pb.CastVoteUnauthn
		return &pb.Status{Code: &status}, nil
	}

	rows, err := s.db.Query("SELECT id, end_date FROM 'elections' WHERE name = $1", *v.ElectionName)
	if err != nil {
		panic(err)
	}
	if !rows.Next() {
		rows.Close()
		status := pb.CastVoteNotFound
		return &pb.Status{Code: &status}, nil
	}
	var (
		id int64
		timeStr string
	)
	rows.Scan(&id, &timeStr)
	rows.Close()
	var endTime time.Time
	err = endTime.UnmarshalText([]byte(timeStr))
	if err != nil {
		panic(err)
	}
	if endTime.Before(time.Now()) {
		return nil, status.Error(codes.Unavailable, "vote ended")
	}

	rows, err = s.db.Query("SELECT [group] FROM 'users' WHERE name = $1", user)
	if err != nil {
		panic(err)
	}
	if !rows.Next() {
		rows.Close()
		status := pb.CastVoteUnauthn
		return &pb.Status{Code: &status}, nil
	}
	var group string
	rows.Scan(&group)
	rows.Close()

	rows, err = s.db.Query("SELECT [group] FROM 'election_groups' WHERE election_id = $1 AND [group] = $2", id, group)
	if err != nil {
		panic(err)
	}
	if !rows.Next() {
		rows.Close()
		status := pb.CastVoteUnauthz
		return &pb.Status{Code: &status}, nil
	}
	rows.Close()

	var choiceId int64
	rows, err = s.db.Query("SELECT id FROM 'election_choices' WHERE election_id = $1 AND choice = $2", id, *v.ChoiceName)
	if err != nil {
		panic(err)
	}
	if !rows.Next() {
		rows.Close()
		return nil, status.Error(codes.NotFound, "no such choice")
	}
	rows.Scan(&choiceId)
	rows.Close()

	rows, err = s.db.Query("SELECT id FROM 'election_voted' WHERE election_id = $1 AND user = $2", id, user)
	if err != nil {
		panic(err)
	}
	if rows.Next() {
		rows.Close()
		status := pb.CastVoteAlready
		return &pb.Status{Code: &status}, nil
	}
	rows.Close()

	_, err = s.db.Exec(`UPDATE 'election_choices' SET votes = votes + 1 WHERE id = $1;
INSERT INTO 'election_voted' ('election_id', 'user') VALUES ($2, $3)`, choiceId, id, user)
	if err != nil {
		panic(err)
	}
	status := pb.CastVoteSuccess
	return &pb.Status{Code: &status}, nil
}

func (s eVotingServer) GetResult(_ context.Context, e *pb.ElectionName) (*pb.ElectionResult, error) {
	rows, err := s.db.Query("SELECT id, end_date FROM 'elections' WHERE name = $1", *e.Name)
	if err != nil {
		panic(err)
	}
	if !rows.Next() {
		rows.Close()
		status := pb.GetResultNotFound
		return &pb.ElectionResult{Status: &status}, nil
	}

	var (
		id int64
		timeStr string
	)
	rows.Scan(&id, &timeStr)
	rows.Close()

	var endTime time.Time
	err = endTime.UnmarshalText([]byte(timeStr))
	if err != nil {
		panic(err)
	}
	if endTime.After(time.Now()) {
		status := pb.GetResultNotYet
		return &pb.ElectionResult{Status: &status}, nil
	}

	var res []*pb.VoteCount
	rows, err = s.db.Query("SELECT choice, votes FROM 'election_choices' WHERE election_id = $1", id)
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var (
			choice string
			votes int32
		)
		rows.Scan(&choice, &votes)
		res = append(res, &pb.VoteCount{ChoiceName: &choice, Count: &votes})
	}
	log.Print(res)
	rows.Close()
	status := pb.GetResultSuccess
	return &pb.ElectionResult{Status: &status, Counts: res}, nil
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

	serverPrivPath := path.Join(dataDir, "key")
	serverPubPath := serverPrivPath + ".pub"

	serverPriv, errPriv := os.ReadFile(serverPrivPath)
	serverPub, errPub := os.ReadFile(serverPubPath)

	kp := sodium.SignKP{
		PublicKey: sodium.SignPublicKey{Bytes: serverPub},
		SecretKey: sodium.SignSecretKey{Bytes: serverPriv},
	}

	if errPriv != nil || errPub != nil {
		kp = sodium.MakeSignKP()

		err = os.WriteFile(serverPrivPath, kp.SecretKey.Bytes, 0600)
		if err != nil {
			log.Fatalf("Unable to write secret key: %v", err)
		}

		err = os.WriteFile(serverPubPath, kp.PublicKey.Bytes, 0600)
		if err != nil {
			log.Fatalf("Unable to write public key: %v", err)
		}
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
	pb.RegisterEVotingServer(voteServer, &eVotingServer{keysDir: keysDir, db: db, key: kp})

	go registServer.Serve(registLn)
	voteServer.Serve(voteLn)
}
