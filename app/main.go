package main

import (
	"net"

	"github.com/codecrafters-io/redis-starter-go/app/client"
	handler "github.com/codecrafters-io/redis-starter-go/app/handlers"
	"github.com/codecrafters-io/redis-starter-go/app/server"
)

func main() {
	s := server.Init()

	cmdChan := make(chan *client.Client, 100)
	go CommandLoop(s, cmdChan)

	s.Run(func(c net.Conn) {
		server.HandleConnection(c, cmdChan)
	})
}

func CommandLoop(s *server.Server, cmdChan <-chan *client.Client) {
	for c := range cmdChan {
		handler.HandleRequest(c)
	}
}
