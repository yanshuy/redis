package main

import (
	"bytes"
	"errors"
	"io"
	"log"
	"net"
)

type Request struct {
	conn     net.Conn
	commands []string
}

func NewRequest(conn net.Conn) *Request {
	return &Request{
		conn: conn,
	}
}

func ReadRequest(conn net.Conn) (*Request, error) {
	// TODO: request > 1024
	r := NewRequest(conn)
	b := make([]byte, 1024)
	start := 0
	read := 0
	for {
		n, err := conn.Read(b[start:])
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		o, err := r.parseRequest(b[read : start+n])
		if err != nil {
			return nil, err
		}
		read += o
		start += n
	}

	return r, nil
}

func (r *Request) parseRequest(b []byte) (n int, err error) {
	for n < len(b) {
		i := bytes.IndexByte(b[n:], '\n')
		if i == -1 {
			return 0, nil
		}
		r.commands = append(r.commands, string(b[:i]))
		n += i + 1 // including "\n"
	}
	return n, nil
}

func HandleRequest(r *Request) (err error) {
	for _, cmd := range r.commands {
		switch cmd {
		case "PING":
			r.conn.Write([]byte("+PONG\r\n"))
			return nil
		default:
			return errors.New("unknown command")
		}
	}
	log.Panicln("should not reach")
	return nil
}
