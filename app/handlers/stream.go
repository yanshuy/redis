package handler

import (
	"fmt"

	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/server"
)

func HandleXadd(c *client.Client) resp.Data {
	args := c.Command.Args
	if len(args) < 4 || (len(args)-2)%2 != 0 {
		return resp.WrongArgs("xadd")
	}
	key := args[0]
	stream_key := args[1]
	key_vals := args[2:]

	s, err := server.Global.Store.Xadd(key, stream_key, key_vals)
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

	entries, err := server.Global.Store.XRange(key, startStr, endStr)
	if err != nil {
		return resp.Err(err.Error())
	}

	res := resp.NewData(resp.Array)
	for _, entry := range entries {
		id := resp.NewData(resp.BulkString, fmt.Sprintf("%d-%d", entry.Id.MS, entry.Id.Seq))
		fields := resp.NewData(resp.Array, entry.Fields)

		entryArr := resp.NewData(resp.Array)
		entryArr.Arr = append(entryArr.Arr, id, fields)
		res.Arr = append(res.Arr, entryArr)
	}
	return res
}

func HandleXread(c *client.Client) resp.Data {
	return resp.NewData(resp.Array)
}
