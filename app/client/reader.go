package client

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"

	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
)

type Reader struct {
	buf []byte
	off int
}

func NewReader() *Reader {
	return &Reader{
		buf: make([]byte, 1024),
		off: 0,
	}
}

func (r *Reader) ReadRESP(conn io.Reader) (resp.Data, error) {
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

func (r *Reader) Exchange(conn net.Conn, out resp.Data, expected string) error {
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

func (r *Reader) SaveRDB(conn io.Reader, file io.Writer) error {
	var (
		rdbLen int
		used   int
		err    error
	)

	for {
		rdbLen, used, err = resp.ReadBulkLength(r.buf[:r.off])
		if err == nil {
			break
		}

		if r.off == len(r.buf) {
			return errors.New("RDB header too large")
		}

		n, err := conn.Read(r.buf[r.off:])
		if err != nil {
			return err
		}
		r.off += n
	}

	buffered := min(r.off-used, rdbLen)

	if buffered > 0 {
		_, err := file.Write(r.buf[used : used+buffered])
		if err != nil {
			return err
		}
	}

	rem := rdbLen - buffered
	if rem > 0 {
		_, err := io.CopyN(file, conn, int64(rem))
		if err != nil {
			return err
		}
	}

	left := used + buffered
	copy(r.buf, r.buf[left:r.off])
	r.off -= left
	return nil
}
