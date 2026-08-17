package handler

import (
	"fmt"
	"strings"

	resp "github.com/codecrafters-io/redis-starter-go/app/Resp"
	"github.com/codecrafters-io/redis-starter-go/app/client"
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

	if strings.ToLower(args[0]) != "streams" {
		return resp.WrongArgs("xread")
	}
	args = args[1:]

	mid := len(args) / 2
	streams := args[:mid]
	ids := args[mid:]

	if len(streams) != len(ids) {
		return resp.WrongArgs("xread")
	}

	answers := make([]resp.Data, 0, len(streams))
	for i, stream := range streams {
		id := ids[i]
		entries, err := s.Store.Xread(stream, id)
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
		pack := resp.NewData(resp.Array, answer)
		answers = append(answers, resp.NewData(resp.Array, stream, pack))
	}

	return resp.NewData(resp.Array, answers)
}
