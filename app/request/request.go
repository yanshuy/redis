package request

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

func ReadAndHandleRequest(conn io.ReadWriter) (n int, err error) {
	// TODO: request > 1024
	b := make([]byte, 1024)
	bLen := 0
	for {
		n, err := conn.Read(b[bLen:])
		if n > 0 {
			bLen += n
			r, o, err := Parse(b[:bLen])
			if err != nil {
				return bLen, err
			}
			if o > 0 {
				err := HandleRequest(conn, r)
				if err != nil {
					return bLen, err
				}
				copy(b, b[o:n])
				bLen -= o
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return bLen, err
		}
	}

	return bLen, nil
}

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
		return HandleCmdGet(cmd, args)
	case "set":
		return HandleCmdSet(cmd, args)

	case "rpush":
		if len(args) < 2 {
			return resp.NewData(resp.Error, "wrong number of arguments for 'rpush' command")
		}
		l, err := store.R.Rpush(args[0], args[1:])
		if err != nil {
			return resp.NewData(resp.Error, err.Error())
		}
		return resp.NewData(resp.Integer, int64(l))

	default:
		msg := fmt.Sprintf("unknown command `%s`", cmd)
		return resp.NewData(resp.Error, msg)
	}
}

func HandleCmdGet(cmd string, args []string) resp.DataType {
	if len(args) != 1 {
		return resp.NewData(resp.Error, "wrong number of arguments for 'get' command")
	}
	key := args[0]
	if val, ok := store.R.Get(key); ok {
		return resp.NewData(resp.BulkString, val)
	} else {
		return resp.NewData(resp.BulkString, "")
	}
}

func HandleCmdSet(cmd string, args []string) resp.DataType {
	if len(args) < 2 {
		return resp.NewData(resp.Error, "wrong number of arguments for 'set' command")
	}
	key := args[0]
	val := args[1]
	expiry := 0

	if len(args) >= 4 {
		switch args[2] {
		case "px", "ex":
			exp, err := strconv.Atoi(args[3])
			if err != nil {
				return resp.NewData(resp.Error, "wrong expiry time expected a number")
			}
			if args[2] == "ex" {
				expiry = exp * 1000
			} else {
				expiry = exp
			}
		default:
			return resp.NewData(resp.Error, "unknown argument")
		}
	}

	store.R.Set(key, val, int64(expiry))
	return resp.NewData(resp.String, "OK")
}
