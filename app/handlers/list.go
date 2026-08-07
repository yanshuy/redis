package handler

import (
	"strconv"
	"time"

	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

func HandleRpush(c *client.Client) resp.Data {
	args := c.Command.Args
	if len(args) < 2 {
		return resp.WrongArgs("rpush")
	}
	key := args[0]

	l, err := server.Global.Store.Rpush(key, args[1:])
	if err != nil {
		return resp.Err(err.Error())
	}
	return resp.NewData(resp.Integer, l)
}

func HandleLpush(c *client.Client) resp.Data {
	args := c.Command.Args
	if len(args) < 2 {
		return resp.WrongArgs("lpush")
	}
	key := args[0]

	l, err := server.Global.Store.Lpush(key, args[1:])
	if err != nil {
		return resp.Err(err.Error())
	}
	return resp.NewData(resp.Integer, l)
}

func HandleLpop(c *client.Client) resp.Data {
	args := c.Command.Args
	if len(args) < 1 {
		return resp.WrongArgs("lpop")
	}
	key := args[0]
	pops := 1
	if len(args) == 2 {
		p, err := strconv.Atoi(args[1])
		if err != nil {
			return resp.Err("2nd argument must be a integer")
		}
		if p < 0 {
			return resp.Err("count must be non-negative")
		}
		pops = p
	}
	l, err := server.Global.Store.Lpop(key, pops)
	if err != nil {
		return resp.Err(err.Error())
	}

	switch len(l) {
	case 0:
		return resp.NewData(resp.NullBulkString)
	case 1:
		return resp.NewData(resp.BulkString, l[0])
	default:
		return resp.NewData(resp.Array, l)
	}
}

func HandleLlen(c *client.Client) resp.Data {
	args := c.Command.Args
	if len(args) != 1 {
		return resp.WrongArgs("llen")
	}
	key := args[0]

	l, err := server.Global.Store.Llen(key)
	if err != nil {
		return resp.Err(err.Error())
	}
	return resp.NewData(resp.Integer, l)
}

func HandleLrange(c *client.Client) resp.Data {
	args := c.Command.Args
	if len(args) != 3 {
		return resp.WrongArgs("lrange")
	}
	key := args[0]
	startIdx, err := strconv.Atoi(args[1])
	if err != nil {
		return resp.Err("expected start index to be an integer for 'lrange' command")
	}
	endIdx, err := strconv.Atoi(args[2])
	if err != nil {
		return resp.Err("expected end index to be an integer for 'lrange' command")
	}
	elems, err := server.Global.Store.Lrange(key, startIdx, endIdx)
	if err != nil {
		return resp.Err(err.Error())
	}
	return resp.NewData(resp.Array, elems)
}

func HandleBlpop(c *client.Client) resp.Data {
	args := c.Command.Args
	if len(args) < 2 {
		return resp.WrongArgs("blpop")
	}
	n := len(args)
	keys := args[:n-1]

	timeout_s, err := strconv.ParseFloat(args[n-1], 64)
	if err != nil {
		return resp.Err("timeout is not an integer or out of range")
	}
	if timeout_s < 0 {
		return resp.Err("timeout is negative")
	}

	for _, key := range keys {
		popped, err := server.Global.Store.Lpop(key, 1)
		if err != nil {
			return resp.Err(err.Error())
		}
		if len(popped) == 1 {
			return resp.NewData(resp.Array, []string{key, popped[0]})
		}
	}

	result := make(chan store.BlockResult, 1)
	for _, key := range keys {
		server.Global.Store.BlockedKeys[key] = append(server.Global.Store.BlockedKeys[key], result)
	}

	c.Block()

	go func() {
		var timeout <-chan time.Time
		if timeout_s > 0 {
			timeout = time.After(time.Duration(timeout_s * float64(time.Second)))
		}

		select {
		case msg := <-result:
			c.QueueMessage(resp.NewData(resp.Array, []string{msg.Key, msg.Value}))
		case <-timeout:
			c.QueueMessage(resp.NewData(resp.Array))
		}

		c.UnBlock()
	}()

	return resp.None()
}
