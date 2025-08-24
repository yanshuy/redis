package main

import (
	"fmt"
	"log"
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/request"
)

func main() {
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
