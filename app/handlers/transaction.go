package handler

import (
	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/server"
)

func HandleExec(c *client.Client) {
	if c.InMulti {
		c.InMulti = false
	} else {
		c.RespChan <- resp.Err("EXEC without MULTI")
		return
	}

	dirty := c.CASDirty
	c.CASDirty = false

	queued := c.QueuedCmds
	c.QueuedCmds = nil

	if dirty {
		c.RespChan <- resp.NewData(resp.Array)
		return
	}

	respChanBak := c.RespChan
	results := make([]resp.Data, 0, len(queued))
	c.RespChan = make(chan resp.Data, len(queued))

	for _, cmd := range queued {
		c.Command = cmd
		HandleCmd(c)
		results = append(results, <-c.RespChan)
	}

	c.RespChan = respChanBak
	c.CASDirty = false
	c.RespChan <- resp.NewData(resp.Array, results)
}

func HandleDiscard(c *client.Client) {
	if c.InMulti {
		c.InMulti = false
	} else {
		c.RespChan <- resp.Err("DISCARD without MULTI")
		return
	}
	c.QueuedCmds = nil
	c.CASDirty = false
	c.RespChan <- resp.NewData(resp.String, "OK")
}

func HandleWatch(c *client.Client) {
	args := c.Command.Args
	for _, key := range args {
		c.WatchingKeys.Add(key)
		server.Global.Store.WatchedKeys[key] = append(server.Global.Store.WatchedKeys[key], c)
	}
	c.RespChan <- resp.NewData(resp.String, "OK")
}

func HandleUnWatch(c *client.Client) {
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
	c.RespChan <- resp.NewData(resp.String, "OK")
}
