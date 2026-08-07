package server

import (
	"errors"
	"fmt"
	"net"
	"strings"

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

func (r *Reader) exchange(conn net.Conn, out resp.Data, expected string) error {
	_, err := conn.Write(out.ToResponse())
	if err != nil {
		return fmt.Errorf("failed to send message to peer")
	}
	resp, err := r.ReadRESP(conn)
	if err != nil {
		return fmt.Errorf("failed to read message from peer")
	}
	if !strings.EqualFold(resp.Str, expected) {
		return fmt.Errorf("unexpected response from peer")
	}
	return nil
}
