package request

import (
	"errors"
	"io"
	"strings"

	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
)

func HandleRequest(w io.Writer, rs []resp.DataType) (err error) {
	for _, r := range rs {
		// fmt.Printf("%+v \n", r)
		switch r.Type {
		case resp.Array:
			err = HandleCmd(w, r.Arr[0], r.Arr[1:])
		default:
			err = HandleCmd(w, r, nil)
		}
	}

	return err
}

func HandleCmd(w io.Writer, cmd resp.DataType, args []resp.DataType) error {
	switch strings.ToLower(cmd.String()) {
	case "ping":
		w.Write([]byte("+PONG\r\n"))

	case "echo":
		res := ""
		for _, d := range args {
			res += d.String()
		}
		// TODO: more than 1 args allowed?
		if res == "" {
			w.Write([]byte("-ERROR 'echo' command expects a single argumentd\r\n"))
		} else {
			w.Write([]byte("+" + res + "\r\n"))
		}

	default:
		return errors.New("unknown command: " + cmd.String())
	}
	return nil
}
