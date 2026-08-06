package server

import (
	"flag"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/codecrafters-io/redis-starter-go/app/server/store"
)

type Config struct {
	port       string
	replicaof  *net.TCPAddr
	Role       string
	Dir        string
	Dbfilename string
}

type Server struct {
	Config Config
	Store  *store.RedisStore
}

func NewServer(config Config, store *store.RedisStore) *Server {
	return &Server{
		Config: config,
		Store:  store,
	}
}

func Start() *Server {
	config := NewConfig()
	store, err := store.InitializeStore(config.Dir, config.Dbfilename)
	if err != nil {
		log.Fatal(err)
	}

	s := NewServer(config, store)
	return s
}

func (s *Server) Run(handleConn func(net.Conn)) {
	port := s.Config.port
	l, err := net.Listen("tcp", "0.0.0.0:"+port)
	if err != nil {
		log.Fatal("Failed to bind to port 6379")
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

	var role string
	replicaof := *replicaofFlag
	if replicaof == "" {
		role = "master"
	} else {
		role = "slave"
	}
	dir := *dirFlag
	dbfilename := *dbFileFlag
	return Config{
		port:       *portFlag,
		Role:       role,
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
