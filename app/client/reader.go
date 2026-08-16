package client

import (
	"errors"
	"io"

	resp "github.com/codecrafters-io/redis-starter-go/app/Resp"
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

func (r *Reader) Read_RESP(conn io.Reader) (resp.Data, []byte, error) {
	b := r.buf

	for {
		if r.off > 0 {
			buf := b[:r.off]
			req, o, err := resp.Parse(buf)
			if err != nil {
				return req, nil, err
			}
			if o > 0 {
				raw := make([]byte, o)
				copy(raw, b[:o])
				copy(b, b[o:r.off])
				r.off -= o
				return req, raw, nil
			}
		}

		if r.off == len(b) {
			return resp.Data{}, nil, errors.New("request too large")
		}

		n, err := conn.Read(b[r.off:])
		r.off += n
		if err != nil {
			return resp.Data{}, nil, err
		}
	}
}

func (r *Reader) Read_RDB(conn io.Reader, file io.Writer) error {
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
