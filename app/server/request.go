package server

import (
	"errors"
	"strings"

	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
	"github.com/codecrafters-io/redis-starter-go/app/client"
)

func ReadRequests(c *client.Client, cmdChan chan<- *client.Client) error {
	r := NewRequestReader()
	for {
		if c.Blocked {
			<-c.Unblock
		}
		d, err := r.Read(c)
		if err != nil {
			return err
		}
		cmd, err := ValidateCommand(d)
		if err != nil {
			return err
		}
		c.Command = cmd
		cmdChan <- c
	}
}

func ValidateCommand(req resp.Data) (client.Command, error) {
	cmd := client.Command{}
	switch req.Type {
	case resp.Array:
		if len(req.Arr) == 0 {
			return cmd, errors.New("empty command")
		}
		first := req.Arr[0]
		args := make([]string, 0, len(req.Arr[1:]))
		for _, arg := range req.Arr[1:] {
			if arg.Type != resp.BulkString && arg.Type != resp.String {
				return cmd, errors.New("invalid command")
			}
			args = append(args, arg.Str)
		}
		cmd.Name = strings.ToLower(first.Str)
		cmd.Args = args

	case resp.String, resp.BulkString:
		cmd.Name = strings.ToLower(req.Str)
	default:
		return cmd, errors.New("invalid command")
	}
	return cmd, nil
}
