package handler

import (
	"strconv"
	"time"

	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

func HandleRpush(c *client.Client) {
	args := c.Command.Args
	if len(args) < 2 {
		c.RespChan <- resp.WrongArgs("rpush")
		return
	}
	key := args[0]

	l, err := server.Global.Store.Rpush(key, args[1:])
	if err != nil {
		c.RespChan <- resp.Err(err.Error())
		return
	}
	c.RespChan <- resp.NewData(resp.Integer, l)
}

func HandleLpush(c *client.Client) {
	args := c.Command.Args
	if len(args) < 2 {
		c.RespChan <- resp.WrongArgs("lpush")
		return
	}
	key := args[0]

	l, err := server.Global.Store.Lpush(key, args[1:])
	if err != nil {
		c.RespChan <- resp.Err(err.Error())
		return
	}
	c.RespChan <- resp.NewData(resp.Integer, l)
}

func HandleLpop(c *client.Client) {
	args := c.Command.Args
	if len(args) < 1 {
		c.RespChan <- resp.WrongArgs("lpop")
		return
	}
	key := args[0]
	pops := 1
	if len(args) == 2 {
		p, err := strconv.Atoi(args[1])
		if err != nil {
			c.RespChan <- resp.Err("2nd argument must be a integer")
			return
		}
		if p < 0 {
			c.RespChan <- resp.Err("count must be non-negative")
			return
		}
		pops = p
	}
	l, err := server.Global.Store.Lpop(key, pops)
	if err != nil {
		c.RespChan <- resp.Err(err.Error())
		return
	}

	switch len(l) {
	case 0:
		c.RespChan <- resp.NewData(resp.NullBulkString)
	case 1:
		c.RespChan <- resp.NewData(resp.BulkString, l[0])
	default:
		c.RespChan <- resp.NewData(resp.Array, l)
	}
}

func HandleLlen(c *client.Client) {
	args := c.Command.Args
	if len(args) != 1 {
		c.RespChan <- resp.WrongArgs("llen")
		return
	}
	key := args[0]

	l, err := server.Global.Store.Llen(key)
	if err != nil {
		c.RespChan <- resp.Err(err.Error())
		return
	}
	c.RespChan <- resp.NewData(resp.Integer, l)
}

func HandleLrange(c *client.Client) {
	args := c.Command.Args
	if len(args) != 3 {
		c.RespChan <- resp.WrongArgs("lrange")
		return
	}
	key := args[0]
	startIdx, err := strconv.Atoi(args[1])
	if err != nil {
		c.RespChan <- resp.Err("expected start index to be an integer for 'lrange' command")
		return
	}
	endIdx, err := strconv.Atoi(args[2])
	if err != nil {
		c.RespChan <- resp.Err("expected end index to be an integer for 'lrange' command")
		return
	}
	elems, err := server.Global.Store.Lrange(key, startIdx, endIdx)
	if err != nil {
		c.RespChan <- resp.Err(err.Error())
		return
	}
	c.RespChan <- resp.NewData(resp.Array, elems)
}

func HandleBlpop(c *client.Client) {
	args := c.Command.Args
	if len(args) < 2 {
		c.RespChan <- resp.WrongArgs("blpop")
		return
	}
	n := len(args)
	keys := args[:n-1]

	timeout_s, err := strconv.ParseFloat(args[n-1], 64)
	if err != nil {
		c.RespChan <- resp.Err("timeout is not an integer or out of range")
		return
	}
	if timeout_s < 0 {
		c.RespChan <- resp.Err("timeout is negative")
		return
	}

	for _, key := range keys {
		popped, err := server.Global.Store.Lpop(key, 1)
		if err != nil {
			c.RespChan <- resp.Err(err.Error())
			return
		}
		if len(popped) == 1 {
			c.RespChan <- resp.NewData(resp.Array, []string{key, popped[0]})
			return
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
			c.RespChan <- resp.NewData(resp.Array, []string{msg.Key, msg.Value})
		case <-timeout:
			c.RespChan <- resp.NewData(resp.Array)
		}

		c.UnBlock()
	}()
}
