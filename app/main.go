package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/request"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

var (
	dirFlag    = flag.String("dir", "tmp", "Directory for RDB persistence")
	dbFileFlag = flag.String("dbfilename", "rdb.snapshot", "RDB file name")
	portFlag   = flag.String("port", "6379", "port")
)

func main() {
	flag.Parse()
	config := store.NewConfig("dir", *dirFlag, "dbfilename", *dbFileFlag)

	store, err := store.InitializeStore(config)
	if err != nil {
		log.Fatal(err)
	}

	reqChan := make(chan request.Request, 100)
	go CommandLoop(store, reqChan)

	port := *portFlag
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
		go handleConnection(conn, reqChan)
	}
}

func CommandLoop(store *store.RedisStore, reqChan chan request.Request) {
	for req := range reqChan {
		request.HandleRequest(store, req)
	}
}

func handleConnection(conn net.Conn, reqChan chan<- request.Request) {
	defer conn.Close()

	c := client.NewClient(conn)
	go c.WriteLoop()
	err := request.ReadRequests(c, reqChan)
	if err != nil {
		log.Println("error reading", err)
	}
}
