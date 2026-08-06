package server

import (
	"errors"

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
