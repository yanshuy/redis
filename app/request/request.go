package request

import (
	"fmt"
	"log"
	"net"

	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
	"github.com/codecrafters-io/redis-starter-go/app/client"
	handler "github.com/codecrafters-io/redis-starter-go/app/request/handlers"
	"github.com/codecrafters-io/redis-starter-go/app/server"
)

type Request struct {
	client *client.Client
	cmd    client.Command
}

func NewRequest(c *client.Client, cmd client.Command) Request {
	return Request{
		client: c,
		cmd:    cmd,
	}
}

func ReadRequests(c *client.Client, reqChan chan<- Request) error {
	r := NewRequestReader()
	for {
		if c.Blocked {
			<-c.Unblock
		}
		d, err := r.Read(c)
		if err != nil {
			return err
		}
		cmd, err := ValidateCommand(d)
		if err != nil {
			return err
		}
		req := NewRequest(c, cmd)
		reqChan <- req
	}
}

func HandleConnection(conn net.Conn, reqChan chan<- Request) {
	defer conn.Close()

	c := client.NewClient(conn)
	go c.WriteLoop()

	err := ReadRequests(c, reqChan)
	if err != nil {
		log.Println("error reading", err)
	}
}

func HandleRequest(s *server.Server, req Request) {
	cmd := req.cmd
	c := req.client

	if cmd.Name == "quit" {
		c.RespChan <- resp.NewData(resp.String, "OK")
		c.Close()
		return
	}

	Chain(func() {
		h := handler.NewHandler(s, c)
		res := h.HandleCmd(cmd)
		if !res.Is(resp.Async) {
			c.RespChan <- res
		}
	}, Auth(req), SubscribeMode(req), Multi(req))()
}

type Handler func()

type Middleware func(Handler) Handler

func Chain(h Handler, ms ...Middleware) Handler {
	for i := len(ms) - 1; i >= 0; i-- {
		m := ms[i]
		h = m(h)
	}
	return h
}

func Auth(req Request) Middleware {
	return func(next Handler) Handler {
		return func() {
			c := req.client
			if c.IsAuthenticated() {
				next()
				return
			}
			switch req.cmd.Name {
			case "auth", "hello", "ping":
				next()
			default:
				c.RespChan <- resp.NoAuth()
			}
		}
	}
}

func SubscribeMode(req Request) Middleware {
	return func(next Handler) Handler {
		return func() {
			c := req.client
			if !c.InSubscribeMode() {
				next()
				return
			}
			switch req.cmd.Name {
			case "ping":
				c.RespChan <- resp.NewData(resp.Array, "pong", "")
			case "subscribe", "unsubscribe":
				next()
			default:
				msg := fmt.Sprintf("Can't execute '%s': only (P|S)SUBSCRIBE / (P|S)UNSUBSCRIBE / PING / QUIT / RESET are allowed in this context", req.cmd.Name)
				c.RespChan <- resp.Err(msg)
			}
		}
	}
}

func Multi(req Request) Middleware {
	return func(next Handler) Handler {
		return func() {
			c := req.client
			if !c.InMulti {
				next()
				return
			}
			switch req.cmd.Name {
			case "exec", "discard":
				next()
			case "watch":
				c.RespChan <- resp.Err("WATCH inside MULTI is not allowed")
			default:
				c.QueuedCmds = append(c.QueuedCmds, req.cmd)
				c.RespChan <- resp.NewData(resp.String, "QUEUED")
			}
		}
	}
}
