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

	default:
		msg := fmt.Sprintf("unknown command `%s`", cmd.Name)
		c.RespChan <- resp.Err(msg)
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
