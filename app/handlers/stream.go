package handler

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	resp "github.com/codecrafters-io/redis-starter-go/app/Resp"
	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

func HandleXadd(c *client.Client) resp.Data {
	args := c.Command.Args
	if len(args) < 4 || (len(args)-2)%2 != 0 {
		return resp.WrongArgs("xadd")
	}
	key := args[0]
	stream_key := args[1]
	key_vals := args[2:]

	s, err := s.Store.Xadd(key, stream_key, key_vals)
	if err != nil {
		return resp.Err(err.Error())
	}

	return resp.NewData(resp.BulkString, s)
}

func HandleXrange(c *client.Client) resp.Data {
	args := c.Command.Args
	if len(args) != 3 {
		return resp.WrongArgs("xrange")
	}
	key := args[0]
	if key == "" {
		return resp.Err("key, val must be a string length > 0")
	}
	startStr := args[1]
	endStr := args[2]

	entries, err := s.Store.Xrange(key, startStr, endStr)
	if err != nil {
		return resp.Err(err.Error())
	}

	answer := make([]resp.Data, 0)
	for _, entry := range entries {
		id := fmt.Sprintf("%d-%d", entry.Id.MS, entry.Id.Seq)
		fields := resp.NewData(resp.Array, entry.Fields)

		entryArr := resp.NewData(resp.Array, id, fields)
		answer = append(answer, entryArr)
	}
	return resp.NewData(resp.Array, answer)
}

func HandleXread(c *client.Client) resp.Data {
	args := c.Command.Args
	var duration time.Duration
	blockMode := false

	if len(args) > 0 && strings.EqualFold(args[0], "block") {
		if len(args) < 2 {
			return resp.WrongArgs("xread")
		}
		timeout_ms, err := strconv.ParseUint(args[1], 10, 64)
		if err != nil {
			return resp.Err("timeout is not an integer or out of range")
		}
		duration = time.Duration(timeout_ms) * time.Millisecond
		blockMode = true
		args = args[2:]
	}

	if len(args) == 0 || !strings.EqualFold(args[0], "streams") {
		return resp.WrongArgs("xread")
	}
	args = args[1:]
	if len(args) == 0 || len(args)%2 != 0 {
		return resp.WrongArgs("xread")
	}

	mid := len(args) / 2
	streams := args[:mid]
	ids := args[mid:]

	answers := make([]resp.Data, 0, len(streams))
	for i, stream := range streams {
		id := ids[i]
		var entries []store.StreamEntry
		var err error

		if id != "$" {
			entries, err = s.Store.Xread(stream, id)
			if err != nil {
				return resp.Err(err.Error())
			}
		}

		answer := make([]resp.Data, 0, len(entries))
		for _, entry := range entries {
			id := fmt.Sprintf("%d-%d", entry.Id.MS, entry.Id.Seq)
			fields := resp.NewData(resp.Array, entry.Fields)
			entryArr := resp.NewData(resp.Array, id, fields)
			answer = append(answer, entryArr)
		}
		pack := resp.NewData(resp.Array, answer)
		answers = append(answers, resp.NewData(resp.Array, stream, pack))
	}

	if !blockMode {
		if len(answers) == 0 {
			return resp.NewData(resp.Array)
		}
		return resp.NewData(resp.Array, answers)
	}

	for _, key := range streams {
		s.Store.BlockedOnKeys[key] = append(s.Store.BlockedOnKeys[key], c)
	}
	cleanUp := func() {
		for _, key := range streams {
			s.Store.BlockedOnKeys[key] = filter(s.Store.BlockedOnKeys[key], c)
		}
	}
	onTimeout := func() {
		c.QueueMessage(resp.NewData(resp.Array))
		cleanUp()
	}
	Wait(c, duration, cleanUp, onTimeout)

	return resp.None()
}
