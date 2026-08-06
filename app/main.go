package main

import (
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/request"
	"github.com/codecrafters-io/redis-starter-go/app/server"
)

func main() {
	s := server.Start()

	reqChan := make(chan request.Request, 100)
	go CommandLoop(s, reqChan)

	s.Run(func(conn net.Conn) {
		request.HandleConnection(conn, reqChan)
	})
}

func CommandLoop(s *server.Server, reqChan <-chan request.Request) {
	for req := range reqChan {
		request.HandleRequest(s, req)
	}
}
