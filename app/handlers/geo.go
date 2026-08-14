package handler

import (
	"strconv"

	resp "github.com/codecrafters-io/redis-starter-go/app/Resp"
	"github.com/codecrafters-io/redis-starter-go/app/client"
)

func HandleGeoAdd(c *client.Client) resp.Data {
	args := c.Command.Args
	if len(args) != 4 {
		return resp.WrongArgs("geoadd")
	}
	key := args[0]

	longitude, err := strconv.ParseFloat(args[1], 64)
	if err != nil {
		return resp.Err("longitude is not an integer or out of range")
	}
	latitude, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return resp.Err("latitude is not an integer or out of range")
	}

	member := args[3]

	n, err := s.Store.Geoadd(key, longitude, latitude, member)
	if err != nil {
		return resp.Err(err.Error())
	}
	return resp.NewData(resp.Integer, n)
}
