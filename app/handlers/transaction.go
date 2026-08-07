package handler

import (
	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/server"
)

func HandleExec(c *client.Client) resp.Data {
	if !c.InMulti {
		return resp.Err("EXEC without MULTI")
	}
	c.InMulti = false

	dirty := c.CASDirty

	queued := c.QueuedCmds
	c.QueuedCmds = nil

	HandleUnWatch(c)

	if dirty {
		return resp.NewData(resp.Array)
	}

	results := make([]resp.Data, len(queued))
	for i, cmd := range queued {
		c.Command = cmd
		res := Chain(HandleCmd, Auth, SubscribeMode)(c)
		results[i] = res
	}
	return resp.NewData(resp.Array, results)
}

func HandleDiscard(c *client.Client) resp.Data {
	if c.InMulti {
		c.InMulti = false
	} else {
		return resp.Err("DISCARD without MULTI")
	}
	c.QueuedCmds = nil
	c.CASDirty = false
	return resp.NewData(resp.String, "OK")
}

func HandleWatch(c *client.Client) resp.Data {
	args := c.Command.Args
	for _, key := range args {
		c.WatchingKeys.Add(key)
		server.Global.Store.WatchedKeys[key] = append(server.Global.Store.WatchedKeys[key], c)
	}
	return resp.NewData(resp.String, "OK")
}

func HandleUnWatch(c *client.Client) resp.Data {
	for key := range c.WatchingKeys {
		clients := server.Global.Store.WatchedKeys[key]
		clients = filter(clients, c)

		if len(clients) == 0 {
			delete(server.Global.Store.WatchedKeys, key)
		} else {
			server.Global.Store.WatchedKeys[key] = clients
		}
	}
	clear(c.WatchingKeys)
	c.CASDirty = false
	return resp.NewData(resp.String, "OK")
}

func filter[T comparable](array []T, elem T) []T {
	for i, item := range array {
		if elem == item {
			return append(array[:i], array[i+1:]...)
		}
	}
	return array
}
