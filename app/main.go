package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/request"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

var (
	dirFlag    = flag.String("dir", "tmp", "Directory for RDB persistence")
	dbFileFlag = flag.String("dbfilename", "rdb.snapshot", "RDB file name")
)

func main() {
	flag.Parse()
	store.RDB.InitConfig("dir", *dirFlag, "dbfilename", *dbFileFlag)

	err := store.RDB.RestoreRDBSnapshot()
	if err != nil {
		log.Fatal(err)
	}

	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		log.Fatal("Failed to bind to port 6379")
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	n, err := request.ReadAndHandleRequest(conn)
	if err != nil {
		log.Println("error reading", err, "\nbytes read", n)
	}
}
