package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
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
		if n > 0 {
			start += n
			o, err2 := r.parseRequest(b[read:start])
			if err2 != nil {
				return nil, err2
			}
			fmt.Println("req parse done", r.commands, err, err2)
			read += o
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}

	fmt.Println("returing req", r.commands)
	return r, nil
}

func (r *Request) parseRequest(b []byte) (n int, err error) {
	if len(b) == 0 {
		return 0, nil
	}
	// i := bytes.IndexByte(b[n:], '\n')
	// if i == -1 {
	// 	return 0, nil
	// }
	err = r.HandleRequest(string(b[:]))
	if err != nil {
		log.Println("error handling request", err)
	}
	// n += i + 1 // including "\n"
	n += len(b)
	return n, nil
}

func (r *Request) HandleRequest(line string) (err error) {
	switch {
	case strings.Contains(line, "PING"):
		r.conn.Write([]byte("+PONG\r\n"))
		return nil
	default:
		return errors.New("unknown command")
	}
}
