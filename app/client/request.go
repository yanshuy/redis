package client

import (
	"errors"
	"strings"

	resp "github.com/codecrafters-io/redis-starter-go/app/Resp"
)

type Request struct {
	Client *Client
	Cmd    Command
}

func (c *Client) ReadRequest() (Request, error) {
	if c.Blocked {
		<-c.Unblock
	}
	d, raw, err := c.Reader.Read_RESP(c.Conn)
	if err != nil {
		return Request{}, err
	}
	cmd, err := ValidateCommand(d)
	cmd.Raw = raw
	if err != nil {
		return Request{}, err
	}
	return Request{Client: c, Cmd: cmd}, nil
}

func ValidateCommand(req resp.Data) (Command, error) {
	cmd := Command{}
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
