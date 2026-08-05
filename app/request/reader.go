package request

import (
	"errors"
	"strings"

	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
	"github.com/codecrafters-io/redis-starter-go/app/client"
)

type RequestReader struct {
	buf []byte
	off int
}

func NewRequestReader() RequestReader {
	return RequestReader{
		buf: make([]byte, 1024),
		off: 0,
	}
}

func (r *RequestReader) Read(c *client.Client) (resp.Data, error) {
	b := r.buf

	for {
		if r.off > 0 {
			req, o, err := resp.Parse(b[:r.off])
			if err != nil {
				return req, err
			}
			if o > 0 {
				copy(b, b[o:r.off])
				r.off -= o
				return req, nil
			}
		}

		if r.off == len(b) {
			return resp.Data{}, errors.New("request too large")
		}

		n, err := c.Conn.Read(b[r.off:])
		r.off += n

		if err != nil {
			return resp.Data{}, err
		}
	}
}

func ValidateRequest(req resp.Data) (client.Command, error) {
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
