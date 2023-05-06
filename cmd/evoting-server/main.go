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
	"os/exec"
	"path"
	"sort"
	"time"

	"github.com/jamesruan/sodium"
	pb "github.com/xdavidwu/evoting/proto"
	"github.com/xdavidwu/evoting/store"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	_ "modernc.org/sqlite"
)

var (
	registAddr	= flag.String("registration-listen", "localhost:1234", "Listen address for registration")
	voteAddr	= flag.String("vote-listen", "0.0.0.0:5678", "Listen address for voting")
	syncAddr	= flag.String("sync-listen", "0.0.0.0:5679", "Listen address for syncing")
	primaryAddr	= flag.String("join-primary", "", "Join a primary for primary-backup setup")
	primaryAction	= flag.String("set-primary", "", "Shell command for setting up networking as a primary")
	primary	string	= ""
	nodes	[]string	= []string{}
	dbPath	= ""
)

const (
	dbSchema = `CREATE TABLE IF NOT EXISTS 'users' ('name' TEXT PRIMARY KEY, 'group' TEXT);
CREATE TABLE IF NOT EXISTS 'challenges' ('id' INTEGER PRIMARY KEY AUTOINCREMENT, 'name' TEXT, 'value' TEXT);
CREATE TABLE IF NOT EXISTS 'elections' ('id' INTEGER PRIMARY KEY AUTOINCREMENT, 'name' TEXT UNIQUE, 'end_date' TEXT);
CREATE TABLE IF NOT EXISTS 'election_groups' ('id' INTEGER PRIMARY KEY AUTOINCREMENT, 'election_id' INTEGER, 'group' TEXT, FOREIGN KEY('election_id') REFERENCES elections('id'));
CREATE TABLE IF NOT EXISTS 'election_choices' ('id' INTEGER PRIMARY KEY AUTOINCREMENT, 'election_id' INTEGER, 'choice' TEXT, 'votes' INTEGER DEFAULT 0, FOREIGN KEY('election_id') REFERENCES elections('id'));
CREATE TABLE IF NOT EXISTS 'election_voted' ('id' INTEGER PRIMARY KEY AUTOINCREMENT, 'election_id' INTEGER, 'user' TEXT, FOREIGN KEY('election_id') REFERENCES elections('id'))`
	dbReset = `PRAGMA writable_schema = 1;DELETE FROM sqlite_master;PRAGMA writable_schema = 0;VACUUM;PRAGMA integrity_check;`
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
		syncToBackups()
		os.WriteFile(path.Join(s.keysDir, *v.Name), v.PublicKey, 0600)
		syncKeyToBackups(*v.Name, v.PublicKey)
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
	syncToBackups()
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
	syncToBackups()

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
			syncToBackups()
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
		log.Println("invalid signature")
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
	rows.Next()
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
	syncToBackups()

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
		log.Printf("no rule for %d, %s", id, group)
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
	syncToBackups()
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

type syncServer struct {
	pb.UnimplementedSyncServer
	keysDir string
	dbPath	string
	serverPub	string
	serverPriv	string
	db	*sql.DB
}

func dumpDB(dbPath string) string {
	// FIXME: is there a better way?
	bytes, err := exec.Command("sqlite3", dbPath, ".dump").Output()
	if err != nil {
		panic(err)
	}
	return dbReset + string(bytes)
}

func (s syncServer) dumpKeys() []*pb.Key {
	list := []*pb.Key{}
	files, err := os.ReadDir(s.keysDir)
	if err != nil {
		panic(err)
	}
	for _, f := range files {
		if !f.Type().IsRegular() {
			continue
		}
		name := f.Name()
		bytes, err := os.ReadFile(path.Join(s.keysDir, name))
		if err != nil {
			continue
		}
		list = append(list, &pb.Key{Name: &name, Key: bytes})
	}
	return list
}

func notifyNodesChanged() {
	list := []*pb.NodeIdentifier{}
	for _, node := range nodes {
		list = append(list, &pb.NodeIdentifier{Address: &node})
	}
	res := &pb.NodesList{Primary: &pb.NodeIdentifier{Address: &primary}, Nodes: list}
	for _, node := range nodes {
		conn, err := grpc.Dial(node, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			//TODO
			continue
		}
		defer conn.Close()
		client := pb.NewSyncClient(conn)
		client.NodesChanged(context.Background(), res);
	}
}

func (s syncServer) Join(_ context.Context, newNode *pb.NodeIdentifier) (*pb.Dump, error) {
	if *syncAddr != primary {
		return nil, status.Error(codes.Unavailable, "i'm not primary")
	}
	nodes = append(nodes, *newNode.Address)
	sort.Strings(nodes)
	notifyNodesChanged()
	dump := dumpDB(s.dbPath)
	log.Print(dump)
	serverPriv, _ := os.ReadFile(s.serverPriv)
	serverPub, _ := os.ReadFile(s.serverPub)
	return &pb.Dump{
		Content: &dump,
		Keys: s.dumpKeys(),
		ServerPub: serverPub,
		ServerPriv: serverPriv,
	}, nil
}

func (syncServer) NodesChanged(_ context.Context, newNodes *pb.NodesList) (*pb.Empty, error) {
	if *syncAddr == primary {
		panic("join loops back")
	}
	log.Println("nodes list update")
	primary = *newNodes.Primary.Address
	list := []string{}
	for _, n := range newNodes.Nodes {
		list = append(list, *n.Address)
	}
	nodes = list
	return &pb.Empty{}, nil
}

func (s syncServer) Sql(_ context.Context, req *pb.SqlRequest) (*pb.Empty, error) {
	if *syncAddr == primary {
		panic("i'm the one who primaries")
	}
	s.db.Exec(*req.Command);
	return &pb.Empty{}, nil
}

func (s syncServer) NewKey(_ context.Context, key *pb.Key) (*pb.Empty, error) {
	os.WriteFile(path.Join(s.keysDir, *key.Name), key.Key, 0600)
	return &pb.Empty{}, nil
}

func (syncServer) Ping(context.Context, *pb.Empty) (*pb.Empty, error) {
	return &pb.Empty{}, nil
}

func syncToBackups() {
	// FIXME do just the new queries instead
	dump := dumpDB(dbPath)
	for _, node := range nodes {
		conn, err := grpc.Dial(node, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			//TODO
			continue
		}
		defer conn.Close()
		client := pb.NewSyncClient(conn)
		client.Sql(context.Background(), &pb.SqlRequest{Command: &dump});
	}
}

func syncKeyToBackups(n string, k []byte) {
	for _, node := range nodes {
		conn, err := grpc.Dial(node, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			continue
		}
		defer conn.Close()
		client := pb.NewSyncClient(conn)
		client.NewKey(context.Background(), &pb.Key{Name: &n, Key: k});
	}
}

func waitForPrimeTime() {
	for {
		time.Sleep(time.Second)
		thisPrimary := primary
		conn, err := grpc.Dial(thisPrimary, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			client := pb.NewSyncClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			_, err = client.Ping(ctx, &pb.Empty{})
			cancel()
			if err == nil {
				continue
			}
		}
		log.Print("primary dead?")
		// time to shine?
		idx := -1
		for i, n := range nodes {
			if n == *syncAddr {
				idx = i
			}
		}
		if idx == -1 {
			log.Fatal("cannot find myself in nodes list")
		}
		time.Sleep(time.Duration(idx) * time.Second)
		if primary == thisPrimary {
			log.Print("making me primary")
			primary = *syncAddr
			nodes = append(nodes[:idx], nodes[idx + 1:]...)
			notifyNodesChanged()
			// do some networking stuff
			bytes, _ := exec.Command("/bin/sh", "-c", *primaryAction).CombinedOutput()
			log.Printf("set-primary: %s", string(bytes))
			return
		}
	}
}

func syncFromPrimary(keysDir, pub, priv string, db *sql.DB) {
	os.RemoveAll(keysDir)
	os.MkdirAll(keysDir, 0700)
	conn, err := grpc.Dial(*primaryAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("cannot dial primary: %v", err)
	}
	client := pb.NewSyncClient(conn)
	state, err := client.Join(context.Background(), &pb.NodeIdentifier{Address: syncAddr})
	if err != nil {
		log.Fatalf("cannot join: %v", err)
	}
	_, err = db.Exec(*state.Content)
	if err != nil {
		log.Fatalf("cannot sync db: %v", err)
	}
	for _, key := range state.Keys {
		os.WriteFile(path.Join(keysDir, *key.Name), key.Key, 0600)
	}
	os.WriteFile(pub, state.ServerPub, 0600)
	os.WriteFile(priv, state.ServerPriv, 0600)
	primary = *primaryAddr
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

	dbPath = path.Join(dataDir, "db.sqlite")
	db, err := sql.Open("sqlite", dbPath)
	defer db.Close()
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	_, err = db.Exec(dbSchema)
	if err != nil {
		log.Printf("cannot init db: %v", err)
	}

	syncLn, err := net.Listen("tcp", *syncAddr)
	if err != nil {
		log.Fatalf("failed to listen %s: %v", *syncAddr, err)
	}
	sServer := grpc.NewServer()
	pb.RegisterSyncServer(sServer, &syncServer{keysDir: keysDir, dbPath: dbPath, db: db, serverPub: serverPubPath, serverPriv: serverPrivPath})
	go sServer.Serve(syncLn)

	if *primaryAddr != "" {
		syncFromPrimary(keysDir, serverPubPath, serverPrivPath, db)
	} else {
		primary = *syncAddr
		bytes, _ := exec.Command("/bin/sh", "-c", *primaryAction).CombinedOutput()
		log.Printf("set-primary: %s", string(bytes))
	}
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

	if *primaryAddr != "" { // backup
		go waitForPrimeTime()
	}

	go registServer.Serve(registLn)
	voteServer.Serve(voteLn)
}
