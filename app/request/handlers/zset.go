package handler

import (
	"strconv"

	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
)

func (h Handler) HandleZadd(args []string) resp.Data {
	if len(args) <= 1 {
		return resp.WrongArgs("ZADD")
	}

	key := args[0]
	args = args[1:]

	n := len(args)
	scores := make([]float64, 0, n/2)
	members := make([]string, 0, n/2)

	for i := 0; i < n; i += 2 {
		score, err := strconv.ParseFloat(args[i], 64)
		if err != nil {
			return resp.Err("resulting score is not a number (NaN)")
		}
		scores = append(scores, score)
		members = append(members, args[i+1])
	}

	added, err := h.store.Zadd(key, scores, members)
	if err != nil {
		return resp.Err(err.Error())
	}
	return resp.NewData(resp.Integer, added)
}

func (h Handler) HandleZrank(args []string) resp.Data {
	if len(args) != 2 {
		return resp.WrongArgs("zrank")
	}

	rank, found, err := h.store.Zrank(args[0], args[1])
	if err != nil {
		return resp.Err(err.Error())
	}
	if !found {
		return resp.NewData(resp.NullBulkString)
	}
	return resp.NewData(resp.Integer, rank)
}

func (h Handler) HandleZrange(args []string) resp.Data {
	if len(args) != 3 {
		return resp.WrongArgs("zrange")
	}

	start, err := strconv.Atoi(args[1])
	if err != nil {
		return resp.Err("value is not an integer or out of range")
	}
	end, err := strconv.Atoi(args[2])
	if err != nil {
		return resp.Err("value is not an integer or out of range")
	}

	list, err := h.store.Zrange(args[0], start, end)
	if err != nil {
		return resp.Err(err.Error())
	}
	return resp.NewData(resp.Array, list)
}

func (h Handler) HandleZcard(args []string) resp.Data {
	if len(args) != 1 {
		return resp.WrongArgs("zcard")
	}

	card, err := h.store.Zcard(args[0])
	if err != nil {
		return resp.Err(err.Error())
	}
	return resp.NewData(resp.Integer, card)
}

func (h Handler) HandleZscore(args []string) resp.Data {
	if len(args) != 2 {
		return resp.WrongArgs("zscore")
	}

	score, found, err := h.store.Zscore(args[0], args[1])
	if err != nil {
		return resp.Err(err.Error())
	}
	if !found {
		return resp.NewData(resp.NullBulkString)
	}
	return resp.NewData(resp.BulkString, strconv.FormatFloat(score, 'g', -1, 64))
}

func (h Handler) HandleZrem(args []string) resp.Data {
	if len(args) != 2 {
		return resp.WrongArgs("zrem")
	}

	removed, err := h.store.Zrem(args[0], args[1])
	if err != nil {
		return resp.Err(err.Error())
	}
	return resp.NewData(resp.Integer, removed)
}
