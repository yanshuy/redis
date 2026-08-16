package main

import (
	"flag"

	"github.com/codecrafters-io/redis-starter-go/app/client"
	handler "github.com/codecrafters-io/redis-starter-go/app/handlers"
	"github.com/codecrafters-io/redis-starter-go/app/server"
)

func main() {
	flag.Parse()
	s := server.Init()

	go CommandLoop(s, s.ReqChan, s.BlopChan)

	s.Run()
}

func CommandLoop(s *server.Server, reqChan <-chan client.Request, blopChan <-chan func()) {
	for {
		select {
		case req := <-reqChan:
			handler.HandleRequest(req)
		case t := <-blopChan:
			t()
		}
	}
}
