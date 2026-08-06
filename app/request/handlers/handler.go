package handler

import (
	"fmt"
	"log"

	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/server"

	"github.com/codecrafters-io/redis-starter-go/app/server/store"
)

type Handler struct {
	s      *server.Server
	store  *store.RedisStore
	client *client.Client
}

func NewHandler(s *server.Server, client *client.Client) Handler {
	return Handler{
		s:      s,
		store:  s.Store,
		client: client,
	}
}

func (h Handler) HandleCmd(cmd client.Command) resp.Data {
	c := h.client
	args := cmd.Args

	switch cmd.Name {
	case "ping":
		return resp.NewData(resp.String, "PONG")

	case "echo":
		if len(args) != 1 {
			return resp.WrongArgs("echo")
		}
		return resp.NewData(resp.BulkString, args[0])

	case "type":
		return h.HandleType(args)

	case "keys":
		return h.HandleKeys(args)

	case "acl":
		return h.HandleACL(args)

	case "auth":
		return h.HandleAuth(args)

	case "get":
		return h.HandleGet(args)

	case "set":
		return h.HandleSet(args)

	case "subscribe":
		return h.HandleSubscribe(args)

	case "unsubscribe":
		return h.HandleUnsubscribe(args)

	case "publish":
		return h.HandlePublish(args)

	case "rpush":
		return h.HandleRpush(args)

	case "lpush":
		return h.HandleLpush(args)

	case "llen":
		return h.HandleLlen(args)

	case "lpop":
		return h.HandleLpop(args)

	case "lrange":
		return h.HandleLrange(args)

	case "blpop":
		return h.HandleBlpop(args)

	case "zadd":
		return h.HandleZadd(args)

	case "zrank":
		return h.HandleZrank(args)

	case "zrange":
		return h.HandleZrange(args)

	case "zcard":
		return h.HandleZcard(args)

	case "zscore":
		return h.HandleZscore(args)

	case "zrem":
		return h.HandleZrem(args)

	case "xadd":
		return h.HandleXadd(args)

	case "xrange":
		return h.HandleXrange(args)

	case "xread":
		return h.HandleXread(args)

	case "config":
		return h.HandleConfig(args)

	case "save":
		err := store.SaveRDBSnapshot(h.store)
		if err != nil {
			log.Println(err)
			return resp.Err("save failed")
		}
		return resp.NewData(resp.String, "OK")

	case "incr":
		return h.HandleCmdIncr(args)

	case "multi":
		c.InMulti = true
		return resp.NewData(resp.String, "OK")

	case "exec":
		return h.HandleExec()

	case "discard":
		return h.HandleDiscard()

	case "watch":
		return h.HandleWatch(args)

	case "unwatch":
		return h.HandleUnWatch()

	case "info":
		return h.HandleInfo(args)

	default:
		msg := fmt.Sprintf("unknown command `%s`", cmd.Name)
		return resp.Err(msg)
	}
}

func filter[T comparable](array []T, elem T) []T {
	for i, item := range array {
		if elem == item {
			return append(array[:i], array[i+1:]...)
		}
	}
	return array
}
