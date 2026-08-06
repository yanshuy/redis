package server

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

type Config struct {
	port       string
	Dir        string
	Dbfilename string
}

type Server struct {
	Config            Config
	Role              string
	replicaof         string
	ReplicationId     string
	ReplicationOffset int
	Store             *store.RedisStore
}

var Global *Server

func NewServer(config Config, store *store.RedisStore) *Server {
	return &Server{
		Config:            config,
		Store:             store,
		ReplicationId:     generateRandID(),
		ReplicationOffset: 0,
	}
}

func Init() *Server {
	config := NewConfig()
	store, err := store.InitializeStore(config.Dir, config.Dbfilename)
	if err != nil {
		log.Fatal(err)
	}

	Global = NewServer(config, store)
	replicaof := *replicaofFlag
	if replicaof == "" {
		Global.Role = MASTER
	} else {
		Global.Role = SLAVE
		Global.replicaof = replicaof
	}
	return Global
}

func (s *Server) Run(handleConn func(net.Conn)) {
	port := s.Config.port

	l, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		log.Fatal("Failed to bind to port 6379")
	}

	if s.Role == SLAVE {
		err := s.HandshakeMaster()
		if err != nil {
			log.Fatal(err)
		}
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}
		go handleConn(conn)
	}
}

func HandleConnection(conn net.Conn, cmdChan chan<- *client.Client) {
	defer conn.Close()

	c := client.NewClient(conn)
	go c.WriteLoop()

	err := ReadRequests(c, cmdChan)
	if err != nil {
		log.Println("error reading", err)
	}
}

var (
	dirFlag       = flag.String("dir", "tmp", "Directory for RDB persistence")
	dbFileFlag    = flag.String("dbfilename", "rdb.snapshot", "RDB file name")
	portFlag      = flag.String("port", "6379", "port")
	replicaofFlag = flag.String("replicaof", "", "replica of")
)

var (
	MASTER = "master"
	SLAVE  = "slave"
)

func NewConfig() Config {
	flag.Parse()

	dir := *dirFlag
	dbfilename := *dbFileFlag
	return Config{
		port:       *portFlag,
		Dir:        dir,
		Dbfilename: dbfilename,
	}
}

func (config *Config) GetConfig(args []string) ([]string, error) {
	result := make([]string, 0)
	for _, arg := range args {
		var val string
		switch strings.ToLower(arg) {
		case "dir":
			val = config.Dir
		case "dbfilename":
			val = config.Dbfilename
		default:
			continue
		}
		result = append(result, arg, val)
	}
	return result, nil
}

func generateRandID() string {
	bytes := make([]byte, 20)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
