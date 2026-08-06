package server

import (
	"errors"
	"net"

	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
)

type Reader struct {
	buf []byte
	off int
}

func NewReader() Reader {
	return Reader{
		buf: make([]byte, 1024),
		off: 0,
	}
}

func (r *Reader) ReadRESP(conn net.Conn) (resp.Data, error) {
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

		n, err := conn.Read(b[r.off:])
		r.off += n

		if err != nil {
			return resp.Data{}, err
		}
	}
}
