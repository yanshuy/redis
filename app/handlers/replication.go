package handler

import (
	"strconv"
	"strings"
	"time"

	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
	"github.com/codecrafters-io/redis-starter-go/app/client"
)

func HandleReplconf(c *client.Client) resp.Data {
	args := c.Command.Args
	if len(args) < 1 {
		return resp.WrongArgs("replconf")
	}
	sub := strings.ToLower(args[0])
	switch sub {
	case "getack":
		off := strconv.Itoa(s.ReplicationOffset)
		res := resp.NewData(resp.Array, []string{"REPLCONF", "ACK", off})
		c.QueueMessage(res)
		return resp.None()
	}
	return resp.NewData(resp.String, "OK")
}

func HandlePsync(c *client.Client) resp.Data {
	args := c.Command.Args
	if len(args) != 2 {
		return resp.WrongArgs("psync")
	}

	s.NewReplica(c)
	return resp.None()
}

func HandleWait(c *client.Client) resp.Data {
	args := c.Command.Args
	if len(args) != 2 {
		return resp.WrongArgs("wait")
	}

	if s.ReplicationOffset == 0 {
		return resp.NewData(resp.Integer, s.ReplicaCount())
	}

	n := len(args)
	timeout_s, err := strconv.ParseFloat(args[n-1], 64)
	if err != nil {
		return resp.Err("timeout is not an integer or out of range")
	}
	if timeout_s < 0 {
		return resp.Err("timeout is negative")
	}

	go func() {
		var timeout <-chan time.Time
		if timeout_s > 0 {
			timeout = time.After(time.Duration(timeout_s * float64(time.Second)))
		}
		_ = timeout

		c.QueueMessage(resp.NewData(resp.Integer, 0))

		// select {
		// case msg := <-result:
		// 	c.QueueMessage(resp.NewData(resp.Array, []string{msg.Key, msg.Value}))
		// case <-timeout:
		// 	c.QueueMessage(resp.NewData(resp.Integer, 0))
		// }
	}()
	return resp.None()
}
