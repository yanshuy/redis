package handler

import (
	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
)

func (h Handler) HandleExec() resp.Data {
	c := h.client
	if c.InMulti {
		c.InMulti = false
	} else {
		return resp.Err("EXEC without MULTI")
	}

	dirty := c.CASDirty
	c.CASDirty = false

	queued := c.QueuedCmds
	c.QueuedCmds = nil

	if dirty {
		return resp.NewData(resp.Array)
	}

	respArr := make([]resp.Data, 0, len(queued))
	for _, cmd := range queued {
		res := h.HandleCmd(cmd)
		respArr = append(respArr, res)
	}
	h.client.CASDirty = false
	return resp.NewData(resp.Array, respArr)
}

func (h Handler) HandleDiscard() resp.Data {
	c := h.client
	if c.InMulti {
		c.InMulti = false
	} else {
		return resp.Err("DISCARD without MULTI")
	}
	c.QueuedCmds = nil
	c.CASDirty = false
	return resp.NewData(resp.String, "OK")
}

func (h Handler) HandleWatch(args []string) resp.Data {
	c := h.client
	for _, key := range args {
		c.WatchingKeys.Add(key)
		h.store.WatchedKeys[key] = append(h.store.WatchedKeys[key], c)
	}
	return resp.NewData(resp.String, "OK")
}

func (h Handler) HandleUnWatch() resp.Data {
	c := h.client
	for key, _ := range c.WatchingKeys {
		clients := h.store.WatchedKeys[key]
		clients = filter(clients, c)

		if len(clients) == 0 {
			delete(h.store.WatchedKeys, key)
		} else {
			h.store.WatchedKeys[key] = clients
		}
	}
	clear(c.WatchingKeys)
	c.CASDirty = false
	return resp.NewData(resp.String, "OK")
}
