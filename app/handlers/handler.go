package handler

import (
	"fmt"
	"log"

	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

func HandleCommand(c *client.Client) {
	if c.Command.Name == "quit" {
		c.RespChan <- resp.NewData(resp.String, "OK")
		c.Close()
		return
	}

	Chain(func() {
		HandleCmd(c)
	}, Auth(c), SubscribeMode(c), Multi(c))()
}

func HandleCmd(c *client.Client) {
	cmd := c.Command
	args := cmd.Args

	switch cmd.Name {
	case "ping":
		c.RespChan <- resp.NewData(resp.String, "PONG")

	case "echo":
		if len(args) != 1 {
			c.RespChan <- resp.WrongArgs("echo")
			return
		}
		c.RespChan <- resp.NewData(resp.BulkString, args[0])

	case "type":
		HandleType(c)

	case "keys":
		HandleKeys(c)

	case "acl":
		HandleACL(c)

	case "auth":
		HandleAuth(c)

	case "get":
		HandleGet(c)

	case "set":
		HandleSet(c)

	case "subscribe":
		HandleSubscribe(c)

	case "unsubscribe":
		HandleUnsubscribe(c)

	case "publish":
		HandlePublish(c)

	case "rpush":
		HandleRpush(c)

	case "lpush":
		HandleLpush(c)

	case "llen":
		HandleLlen(c)

	case "lpop":
		HandleLpop(c)

	case "lrange":
		HandleLrange(c)

	case "blpop":
		HandleBlpop(c)

	case "zadd":
		HandleZadd(c)

	case "zrank":
		HandleZrank(c)

	case "zrange":
		HandleZrange(c)

	case "zcard":
		HandleZcard(c)

	case "zscore":
		HandleZscore(c)

	case "zrem":
		HandleZrem(c)

	case "xadd":
		HandleXadd(c)

	case "xrange":
		HandleXrange(c)

	case "xread":
		HandleXread(c)

	case "config":
		HandleConfig(c)

	case "save":
		err := store.SaveRDBSnapshot(server.Global.Store)
		if err != nil {
			log.Println(err)
			c.RespChan <- resp.Err("save failed")
			return
		}
		c.RespChan <- resp.NewData(resp.String, "OK")

	case "incr":
		HandleCmdIncr(c)

	case "multi":
		c.InMulti = true
		c.RespChan <- resp.NewData(resp.String, "OK")

	case "exec":
		HandleExec(c)

	case "discard":
		HandleDiscard(c)

	case "watch":
		HandleWatch(c)

	case "unwatch":
		HandleUnWatch(c)

	case "info":
		HandleInfo(c)

	case "replconf":
		HandleReplconf(c)

	case "psync":
		HandlePsync(c)

	default:
		msg := fmt.Sprintf("unknown command `%s`", cmd.Name)
		c.RespChan <- resp.Err(msg)
	}
}

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
