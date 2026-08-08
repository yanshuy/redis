package main

import (
	"github.com/codecrafters-io/redis-starter-go/app/client"
	handler "github.com/codecrafters-io/redis-starter-go/app/handlers"
	"github.com/codecrafters-io/redis-starter-go/app/server"
)

func main() {
	s := server.Init()

	reqChan := make(chan client.Request, 100)
	go CommandLoop(s, reqChan)

	s.Run(func(c *client.Client) {
		server.HandleClient(c, reqChan)
	})
}

func CommandLoop(s *server.Server, reqChan <-chan client.Request) {
	for req := range reqChan {
		handler.HandleRequest(req)
	}
}
