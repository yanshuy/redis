package handler

import (
	"fmt"
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
		if s.Role == client.SLAVE {
			fmt.Println("got getack as a slave sendin", c.Role)
			res := resp.NewData(resp.Array, []string{"REPLCONF", "ACK", strconv.Itoa(s.ReplicationOffset)})
			c.QueueMessage(res) // because returned response are not sent if client is master
		}
		return resp.None()

	case "ack":
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

	numreplicas, err := strconv.Atoi(args[0])
	if err != nil || numreplicas < 0 {
		return resp.Err("numreplicas is not an integer or out of range")
	}
	if numreplicas == 0 {
		return resp.NewData(resp.Integer, 0)
	}
	timeout_s, err := strconv.ParseFloat(args[1], 64)
	if err != nil || timeout_s < 0 {
		return resp.Err("timeout is not an integer or out of range")
	}

	if s.ReplicationOffset == 0 {
		return resp.NewData(resp.Integer, s.ReplicaCount())
	}

	targetOffset := s.ReplicationOffset

	go func() {
		var timeout <-chan time.Time
		if timeout_s > 0 {
			timeout = time.After(time.Duration(timeout_s * float64(time.Millisecond)))
		}

		ackChan := s.RequestAcks()
		acksRev := 0
	outer:
		for acksRev < numreplicas {
			select {
			case ack := <-ackChan:
				if ack.Offset >= targetOffset {
					acksRev++
				}
			case <-timeout:
				break outer
			}
		}
		c.QueueMessage(resp.NewData(resp.Integer, acksRev))
	}()

	return resp.None()
}
