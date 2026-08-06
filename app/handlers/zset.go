package handler

import (
	"strconv"

	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/server"
)

func HandleZadd(c *client.Client) {
	args := c.Command.Args
	if len(args) <= 1 {
		c.RespChan <- resp.WrongArgs("ZADD")
		return
	}

	key := args[0]
	args = args[1:]

	n := len(args)
	scores := make([]float64, 0, n/2)
	members := make([]string, 0, n/2)

	for i := 0; i < n; i += 2 {
		score, err := strconv.ParseFloat(args[i], 64)
		if err != nil {
			c.RespChan <- resp.Err("resulting score is not a number (NaN)")
			return
		}
		scores = append(scores, score)
		members = append(members, args[i+1])
	}

	added, err := server.Global.Store.Zadd(key, scores, members)
	if err != nil {
		c.RespChan <- resp.Err(err.Error())
		return
	}
	c.RespChan <- resp.NewData(resp.Integer, added)
}

func HandleZrank(c *client.Client) {
	args := c.Command.Args
	if len(args) != 2 {
		c.RespChan <- resp.WrongArgs("zrank")
		return
	}

	rank, found, err := server.Global.Store.Zrank(args[0], args[1])
	if err != nil {
		c.RespChan <- resp.Err(err.Error())
		return
	}
	if !found {
		c.RespChan <- resp.NewData(resp.NullBulkString)
		return
	}
	c.RespChan <- resp.NewData(resp.Integer, rank)
}

func HandleZrange(c *client.Client) {
	args := c.Command.Args
	if len(args) != 3 {
		c.RespChan <- resp.WrongArgs("zrange")
		return
	}

	start, err := strconv.Atoi(args[1])
	if err != nil {
		c.RespChan <- resp.Err("value is not an integer or out of range")
		return
	}
	end, err := strconv.Atoi(args[2])
	if err != nil {
		c.RespChan <- resp.Err("value is not an integer or out of range")
		return
	}

	list, err := server.Global.Store.Zrange(args[0], start, end)
	if err != nil {
		c.RespChan <- resp.Err(err.Error())
		return
	}
	c.RespChan <- resp.NewData(resp.Array, list)
}

func HandleZcard(c *client.Client) {
	args := c.Command.Args
	if len(args) != 1 {
		c.RespChan <- resp.WrongArgs("zcard")
		return
	}

	card, err := server.Global.Store.Zcard(args[0])
	if err != nil {
		c.RespChan <- resp.Err(err.Error())
		return
	}
	c.RespChan <- resp.NewData(resp.Integer, card)
}

func HandleZscore(c *client.Client) {
	args := c.Command.Args
	if len(args) != 2 {
		c.RespChan <- resp.WrongArgs("zscore")
		return
	}

	score, found, err := server.Global.Store.Zscore(args[0], args[1])
	if err != nil {
		c.RespChan <- resp.Err(err.Error())
		return
	}
	if !found {
		c.RespChan <- resp.NewData(resp.NullBulkString)
		return
	}
	c.RespChan <- resp.NewData(resp.BulkString, strconv.FormatFloat(score, 'g', -1, 64))
}

func HandleZrem(c *client.Client) {
	args := c.Command.Args
	if len(args) != 2 {
		c.RespChan <- resp.WrongArgs("zrem")
		return
	}

	removed, err := server.Global.Store.Zrem(args[0], args[1])
	if err != nil {
		c.RespChan <- resp.Err(err.Error())
		return
	}
	c.RespChan <- resp.NewData(resp.Integer, removed)
}
