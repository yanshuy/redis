package main

import (
	"flag"
	"log"

	handler "github.com/codecrafters-io/redis-starter-go/app/handlers"
	"github.com/codecrafters-io/redis-starter-go/app/server"
)

func main() {
	flag.Parse()

	s := server.Init()
	s.ReqHandler = handler.HandleRequest
	s.CmdHandler = handler.HandleCmd

	if err := s.InitAOF(); err != nil {
		log.Fatal(err)
	}

	s.Run()
}
