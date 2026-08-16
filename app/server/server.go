package server

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

var (
	MASTER = client.MASTER
	SLAVE  = client.SLAVE
)

type Server struct {
	Role              client.Role
	replicaof         string
	ReplicationId     string
	ReplicationOffset int

	Config    Config
	Store     *store.RedisStore
	Replicas  map[*client.Client]int
	BlClients []*client.Client

	ReqChan  chan client.Request
	BlopChan chan func()
}

func (s *Server) ReplicaCount() int {
	return len(s.Replicas)
}

var Svr *Server

func NewServer(role client.Role, config Config, store *store.RedisStore) *Server {
	s := &Server{
		Config:            config,
		Role:              role,
		Store:             store,
		replicaof:         *replicaofFlag,
		ReplicationId:     generateRandID(),
		ReplicationOffset: 0,
		Replicas:          make(map[*client.Client]int),
		ReqChan:           make(chan client.Request, 100),
		BlopChan:          make(chan func(), 100),
	}
	return s
}

var (
	portFlag      = flag.String("port", "6379", "port")
	replicaofFlag = flag.String("replicaof", "", "replica of")
)

func Init() *Server {
	config := NewConfig()
	store, err := store.InitializeStore(config.Dir, config.Dbfilename)
	if err != nil {
		log.Fatal(err)
	}

	if *replicaofFlag == "" {
		Svr = NewServer(MASTER, config, store)
	} else {
		Svr = NewServer(SLAVE, config, store)
	}
	return Svr
}

func (s *Server) Run() {
	port := s.Config.port

	l, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		log.Fatal("Failed to bind to port 6379")
	}

	if s.Role == client.SLAVE {
		master, err := s.HandshakeMaster()
		if err != nil {
			log.Fatal(err)
		}
		go s.HandleRequests(master)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}
		c := client.NewClient(conn, client.CLIENT)
		go s.HandleRequests(c)
	}
}

func (s *Server) HandleRequests(c *client.Client) {
	defer c.Close()

	for {
		req, err := c.ReadRequest()
		if err != nil {
			log.Println("error reading", err)
			return
		}
		s.ReqChan <- req
	}
}

func generateRandID() string {
	bytes := make([]byte, 20)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
