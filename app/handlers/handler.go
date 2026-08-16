package handler

import (
	"context"
	"fmt"
	"log"
	"time"

	resp "github.com/codecrafters-io/redis-starter-go/app/Resp"
	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

var s *server.Server

var handler = Chain(HandleCmd, Auth, SubscribeMode, Multi)

func HandleRequest(req client.Request) {
	s = server.Svr // crime

	c := req.Client
	c.Command = req.Cmd

	res := handler(c)
	if c.Role == client.MASTER {
		s.ReplicationOffset += len(req.Cmd.Raw)
	} else {
		c.QueueMessage(res)
	}
}

func HandleCmd(c *client.Client) resp.Data {
	s = server.Svr
	cmd := c.Command
	args := cmd.Args

	switch cmd.Name {
	case "ping":
		return resp.NewData(resp.String, "PONG")

	case "echo":
		if len(args) != 1 {
			return resp.WrongArgs("echo")
		}
		return resp.NewData(resp.BulkString, args[0])

	case "quit":
		res := resp.NewData(resp.String, "OK")
		c.Conn.Write(res.ToResponse())
		c.Close()
		return resp.None()

	case "type":
		return HandleType(c)

	case "keys":
		return HandleKeys(c)

	case "acl":
		return HandleACL(c)

	case "auth":
		return HandleAuth(c)

	case "get":
		return HandleGet(c)

	case "set":
		return HandleSet(c)

	case "subscribe":
		return HandleSubscribe(c)

	case "unsubscribe":
		return HandleUnsubscribe(c)

	case "publish":
		return HandlePublish(c)

	case "rpush":
		return HandleRpush(c)

	case "lpush":
		return HandleLpush(c)

	case "llen":
		return HandleLlen(c)

	case "lpop":
		return HandleLpop(c)

	case "lrange":
		return HandleLrange(c)

	case "blpop":
		return HandleBlpop(c)

	case "zadd":
		return HandleZadd(c)

	case "zrank":
		return HandleZrank(c)

	case "zrange":
		return HandleZrange(c)

	case "zcard":
		return HandleZcard(c)

	case "zscore":
		return HandleZscore(c)

	case "zrem":
		return HandleZrem(c)

	case "xadd":
		return HandleXadd(c)

	case "xrange":
		return HandleXrange(c)

	case "xread":
		return HandleXread(c)

	case "config":
		return HandleConfig(c)

	case "save":
		err := store.SaveRDBSnapshot(s.Store)
		if err != nil {
			log.Println(err)
			return resp.Err("save failed")
		}
		return resp.NewData(resp.String, "OK")

	case "incr":
		return HandleIncr(c)

	case "multi":
		c.InMulti = true
		return resp.NewData(resp.String, "OK")

	case "exec":
		return HandleExec(c)

	case "discard":
		return HandleDiscard(c)

	case "watch":
		return HandleWatch(c)

	case "unwatch":
		return HandleUnWatch(c)

	case "info":
		return HandleInfo(c)

	case "replconf":
		return HandleReplconf(c)

	case "psync":
		return HandlePsync(c)

	case "wait":
		return HandleWait(c)

	case "geoadd":
		return HandleGeoAdd(c)

	case "geopos":
		return HandleGeoPos(c)

	case "geodist":
		return HandleGeodist(c)

	case "geosearch":
		return HandleGeosearch(c)

	default:
		msg := fmt.Sprintf("unknown command `%s`", cmd.Name)
		return resp.Err(msg)
	}
}

type MiddlewareFunc func(*client.Client) resp.Data

type Middleware func(MiddlewareFunc) MiddlewareFunc

func Chain(h MiddlewareFunc, ms ...Middleware) MiddlewareFunc {
	for i := len(ms) - 1; i >= 0; i-- {
		m := ms[i]
		h = m(h)
	}
	return h
}

func Auth(next MiddlewareFunc) MiddlewareFunc {
	return func(c *client.Client) resp.Data {
		if c.IsAuthenticated() {
			return next(c)
		}
		switch c.Command.Name {
		case "auth", "hello", "ping":
			return next(c)
		default:
			return resp.NoAuth()
		}
	}
}

func SubscribeMode(next MiddlewareFunc) MiddlewareFunc {
	return func(c *client.Client) resp.Data {
		if !c.InSubscribeMode() {
			return next(c)
		}
		switch c.Command.Name {
		case "ping":
			return resp.NewData(resp.Array, "pong", "")
		case "subscribe", "unsubscribe":
			return next(c)
		default:
			msg := fmt.Sprintf("Can't execute '%s': only (P|S)SUBSCRIBE / (P|S)UNSUBSCRIBE / PING / QUIT / RESET are allowed in this context", c.Command.Name)
			return resp.Err(msg)
		}
	}
}

func Multi(next MiddlewareFunc) MiddlewareFunc {
	return func(c *client.Client) resp.Data {
		if !c.InMulti {
			return next(c)
		}
		switch c.Command.Name {
		case "exec", "discard":
			return next(c)
		case "watch":
			return resp.Err("WATCH inside MULTI is not allowed")
		default:
			c.QueuedCmds = append(c.QueuedCmds, c.Command)
			return resp.NewData(resp.String, "QUEUED")
		}
	}
}

func Wait(c *client.Client, timeout time.Duration, onTimeout func(), onSuccess func()) {
	var ctx context.Context
	var cancel context.CancelFunc
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
	} else {
		ctx, cancel = context.WithCancel(context.Background())
	}
	c.Blop.Cancel = cancel

	c.Block()
	go func() {
		<-ctx.Done()
		c.Blocked = false
		c.UnBlock()

		if ctx.Err() == context.Canceled {
			s.BlopChan <- onSuccess
		}

		if ctx.Err() == context.DeadlineExceeded {
			s.BlopChan <- onTimeout
			cancel()
		}
	}()
}
