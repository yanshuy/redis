package handler

import (
	"strconv"
	"time"

	resp "github.com/codecrafters-io/redis-starter-go/app/Resp"
	"github.com/codecrafters-io/redis-starter-go/app/client"
)

func HandleRpush(c *client.Client) resp.Data {
	args := c.Command.Args
	if len(args) < 2 {
		return resp.WrongArgs("rpush")
	}
	key := args[0]

	l, err := s.Store.Rpush(key, args[1:])
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

	l, err := s.Store.Lpush(key, args[1:])
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
	l, err := s.Store.Lpop(key, pops)
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

	l, err := s.Store.Llen(key)
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
	elems, err := s.Store.Lrange(key, startIdx, endIdx)
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
		popped, err := s.Store.Lpop(key, 1)
		if err != nil {
			return resp.Err(err.Error())
		}
		if len(popped) == 1 {
			return resp.NewData(resp.Array, []string{key, popped[0]})
		}
	}

	for _, key := range keys {
		s.Store.BlockedKeys[key] = append(s.Store.BlockedKeys[key], c)
	}

	duration := time.Duration(timeout_s * float64(time.Second))
	c.Wait(duration, func() {
		c.QueueMessage(resp.NewData(resp.Array))
	}, func() {})

	return resp.None()
}
