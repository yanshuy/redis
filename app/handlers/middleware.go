package handler

import (
	"fmt"

	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
	"github.com/codecrafters-io/redis-starter-go/app/client"
)

type MiddlewareFunc func()

type Middleware func(MiddlewareFunc) MiddlewareFunc

func Chain(h MiddlewareFunc, ms ...Middleware) MiddlewareFunc {
	for i := len(ms) - 1; i >= 0; i-- {
		m := ms[i]
		h = m(h)
	}
	return h
}

func Auth(c *client.Client) Middleware {
	return func(next MiddlewareFunc) MiddlewareFunc {
		return func() {
			if c.IsAuthenticated() {
				next()
				return
			}
			switch c.Command.Name {
			case "auth", "hello", "ping":
				next()
			default:
				c.RespChan <- resp.NoAuth()
			}
		}
	}
}

func SubscribeMode(c *client.Client) Middleware {
	return func(next MiddlewareFunc) MiddlewareFunc {
		return func() {
			if !c.InSubscribeMode() {
				next()
				return
			}
			switch c.Command.Name {
			case "ping":
				c.RespChan <- resp.NewData(resp.Array, "pong", "")
			case "subscribe", "unsubscribe":
				next()
			default:
				msg := fmt.Sprintf("Can't execute '%s': only (P|S)SUBSCRIBE / (P|S)UNSUBSCRIBE / PING / QUIT / RESET are allowed in this context", c.Command.Name)
				c.RespChan <- resp.Err(msg)
			}
		}
	}
}

func Multi(c *client.Client) Middleware {
	return func(next MiddlewareFunc) MiddlewareFunc {
		return func() {
			if !c.InMulti {
				next()
				return
			}
			switch c.Command.Name {
			case "exec", "discard":
				next()
			case "watch":
				c.RespChan <- resp.Err("WATCH inside MULTI is not allowed")
			default:
				c.QueuedCmds = append(c.QueuedCmds, c.Command)
				c.RespChan <- resp.NewData(resp.String, "QUEUED")
			}
		}
	}
}
