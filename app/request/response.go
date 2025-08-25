package request

import (
	"fmt"
	"io"
	"strings"

	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
)

var Store = make(map[string]string)

func HandleRequest(w io.Writer, rs []resp.DataType) (err error) {
	for _, r := range rs {
		// fmt.Printf("%+v \n", r)
		var res resp.DataType
		switch r.Type {
		case resp.Array:
			args := make([]string, 0, len(r.Arr))
			for i := 1; i < len(r.Arr); i++ {
				args = append(args, r.Arr[i].Str)
			}
			res = HandleCmd(r.Arr[0].Str, args)

		default:
			res = HandleCmd(r.Str, nil)
		}

		_, err := w.Write(res.ToResponse())
		if err != nil {
			return err
		}
	}

	return err
}

func HandleCmd(cmd string, args []string) resp.DataType {
	switch strings.ToLower(cmd) {
	case "ping":
		return resp.NewData(resp.String, "PONG")

	case "echo":
		if len(args) != 1 {
			return resp.NewData(resp.Error, "wrong number of arguments for 'echo' command")
		}
		return resp.NewData(resp.String, args[0])

	case "get":
		if len(args) != 1 {
			return resp.NewData(resp.Error, "wrong number of arguments for 'get' command")
		}
		key := args[0]
		if val, ok := Store[key]; ok {
			return resp.NewData(resp.BulkString, val)
		} else {
			return resp.NewData(resp.BulkString, "")
		}

	case "set":
		if len(args) != 2 {
			return resp.NewData(resp.Error, "wrong number of arguments for 'set' command")
		}
		key := args[0]
		val := args[1]
		Store[key] = val

		return resp.NewData(resp.String, "OK")

	default:
		msg := fmt.Sprintf("unknown command `%s`", cmd)
		return resp.NewData(resp.Error, msg)
	}
}
