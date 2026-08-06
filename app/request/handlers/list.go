package handler

import (
	"strconv"
	"time"

	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
	"github.com/codecrafters-io/redis-starter-go/app/server/store"
)

func (h Handler) HandleRpush(args []string) resp.Data {
	if len(args) < 2 {
		return resp.WrongArgs("rpush")
	}
	key := args[0]

	l, err := h.store.Rpush(key, args[1:])
	if err != nil {
		return resp.Err(err.Error())
	}
	return resp.NewData(resp.Integer, l)
}

func (h Handler) HandleLpush(args []string) resp.Data {
	if len(args) < 2 {
		return resp.WrongArgs("lpush")
	}
	key := args[0]

	l, err := h.store.Lpush(key, args[1:])
	if err != nil {
		return resp.Err(err.Error())
	}
	return resp.NewData(resp.Integer, l)
}

func (h Handler) HandleLpop(args []string) resp.Data {
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
	l, err := h.store.Lpop(key, pops)
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

func (h Handler) HandleLlen(args []string) resp.Data {
	if len(args) != 1 {
		return resp.WrongArgs("llen")
	}
	key := args[0]

	l, err := h.store.Llen(key)
	if err != nil {
		return resp.Err(err.Error())
	}
	return resp.NewData(resp.Integer, l)
}

func (h Handler) HandleLrange(args []string) resp.Data {
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
	elems, err := h.store.Lrange(key, startIdx, endIdx)
	if err != nil {
		return resp.Err(err.Error())
	}
	return resp.NewData(resp.Array, elems)
}

func (h Handler) HandleBlpop(args []string) resp.Data {
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
		popped, err := h.store.Lpop(key, 1)
		if err != nil {
			return resp.Err(err.Error())
		}
		if len(popped) == 1 {
			return resp.NewData(resp.Array, []string{key, popped[0]})
		}
	}

	result := make(chan store.BlockResult)
	for _, key := range keys {
		h.store.BlockedKeys[key] = append(h.store.BlockedKeys[key], result)
	}

	h.client.Block()

	go func() {
		var timeout <-chan time.Time
		if timeout_s > 0 {
			timeout = time.After(time.Duration(timeout_s * float64(time.Second)))
		}

		select {
		case msg := <-result:
			h.client.RespChan <- resp.NewData(resp.Array, []string{msg.Key, msg.Value})
		case <-timeout:
			h.client.RespChan <- resp.NewData(resp.Array)
		}

		h.client.UnBlock()
		for _, key := range keys {
			h.store.BlockedKeys[key] = filter(h.store.BlockedKeys[key], result)
		}
	}()

	return resp.Future()
}
