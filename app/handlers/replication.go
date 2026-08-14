package handler

import (
	"strconv"
	"strings"
	"time"

	resp "github.com/codecrafters-io/redis-starter-go/app/Resp"
	"github.com/codecrafters-io/redis-starter-go/app/client"
)

func HandleReplconf(c *client.Client) resp.Data {
	args := c.Command.Args
	if len(args) < 1 {
		return resp.WrongArgs("replconf")
	}

	sub := strings.ToLower(args[0])
	switch s.Role {
	case client.SLAVE:
		switch sub {
		case "getack":
			res := resp.NewData(resp.Array, []string{"REPLCONF", "ACK", strconv.Itoa(s.ReplicationOffset)})
			c.QueueMessage(res) // because returned response are not sent if client is master
			return resp.None()
		}
	case client.MASTER:
		switch sub {
		case "ack":
			if len(args) != 2 {
				return resp.WrongArgs("replconf|ack")
			}
			off, err := strconv.Atoi(args[1])
			if err != nil || off < 0 {
				return resp.Err("value is not an integer or out of range")
			}
			s.Replicas[c] = off
			for _, blClient := range s.BlClients {
				if s.CountSyncedReplicas(blClient.Blop.Reploffset) >= blClient.Blop.Numreplicas {
					if blClient.Blop.Cancel != nil {
						blClient.Blop.Cancel()
					}
				}
			}
			return resp.None()
		}
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

	numreplicas, err := strconv.Atoi(args[0])
	if err != nil || numreplicas < 0 {
		return resp.Err("numreplicas is not an integer or out of range")
	}

	timeout_s, err := strconv.ParseFloat(args[1], 64)
	if err != nil || timeout_s < 0 {
		return resp.Err("timeout is not an integer or out of range")
	}

	targetOffset := s.ReplicationOffset
	synced := s.CountSyncedReplicas(targetOffset)

	if synced >= numreplicas || targetOffset == 0 {
		return resp.NewData(resp.Integer, synced)
	}

	c.Blop.Reploffset = targetOffset
	c.Blop.Numreplicas = numreplicas
	s.BlClients = append(s.BlClients, c)

	f := func() {
		s.BlClients = filter(s.BlClients, c)
		c.QueueMessage(resp.NewData(resp.Integer, s.CountSyncedReplicas(targetOffset)))
	}
	duration := time.Duration(timeout_s * float64(time.Millisecond))
	c.Wait(duration, f, f)

	getack := resp.NewData(resp.Array, []string{"REPLCONF", "GETACK", "*"})
	for slave := range s.Replicas {
		slave.QueueMessage(getack)
	}

	return resp.None()
}
