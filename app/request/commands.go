package request

import (
	"fmt"
	"strings"

	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

func HandleCmd(cmd string, args []resp.DataType) resp.DataType {

	switch strings.ToLower(cmd) {
	case "ping":
		return resp.NewData(resp.String, "PONG")

	case "echo":
		if len(args) != 1 {
			return resp.NewData(resp.Error, "wrong number of arguments for 'echo' command")
		}
		return args[0]

	case "get":
		if len(args) != 1 {
			return resp.NewData(resp.Error, "wrong number of arguments for 'get' command")
		}
		key := args[0].Str
		if key == "" {
			return resp.NewData(resp.Error, "key must be a string length > 0")
		}
		return HandleCmdGet(key)

	case "set":
		if len(args) < 2 {
			return resp.NewData(resp.Error, "wrong number of arguments for 'set' command")
		}
		key := args[0].Str
		val := args[1].Str
		if key == "" || val == "" {
			return resp.NewData(resp.Error, "key, val must be a string length > 0")
		}
		return HandleCmdSet(key, val, args[2:])

	case "rpush":
		if len(args) < 2 {
			return resp.NewData(resp.Error, "wrong number of arguments for 'rpush' command")
		}
		key := args[0].Str
		if key == "" {
			return resp.NewData(resp.Error, "key must be a string length > 0")
		}
		return HandleRpush(key, args[1:])

	case "lpush":
		if len(args) < 2 {
			return resp.NewData(resp.Error, "wrong number of arguments for 'lpush' command")
		}
		key := args[0].Str
		if key == "" {
			return resp.NewData(resp.Error, "key must be a string length > 0")
		}
		return HandleLpush(key, args[1:])

	case "lrange":
		if len(args) != 3 {
			return resp.NewData(resp.Error, "wrong number of arguments for 'rpush' command")
		}
		key := args[0].Str
		if key == "" {
			return resp.NewData(resp.Error, "key, val must be a string length > 0")
		}
		return HandleLrange(key, args[1:])

	default:
		msg := fmt.Sprintf("unknown command `%s`", cmd)
		return resp.NewData(resp.Error, msg)
	}
}

func HandleCmdGet(key string) resp.DataType {
	if val, ok := store.DB.Get(key); ok {
		return resp.NewData(resp.BulkString, val)
	} else {
		return resp.NewData(resp.BulkString, "")
	}
}

func HandleCmdSet(key, val string, args []resp.DataType) resp.DataType {
	var expiry int64
	if len(args) >= 2 {
		arg := args[0].Str
		switch strings.ToLower(arg) {
		case "px", "ex":
			exp, err := args[1].Integer()
			if err != nil {
				return resp.NewData(resp.Error, "wrong expiry time expected a number")
			}
			if arg == "ex" {
				expiry = exp * 1000
			} else {
				expiry = exp
			}

		default:
			return resp.NewData(resp.Error, "unknown argument for 'set' command")
		}
	}

	store.DB.Set(key, val, expiry)
	return resp.NewData(resp.String, "OK")
}

func HandleRpush(key string, args []resp.DataType) resp.DataType {
	strArgs := make([]string, 0, len(args))
	for _, arg := range args {
		if arg.Is(resp.String) {
			strArgs = append(strArgs, arg.Str)
		} else {
			return resp.NewData(resp.Error, "invalid argument type for 'rpush' command expects only string")
		}
	}
	l, err := store.DB.Rpush(key, strArgs)
	if err != nil {
		return resp.NewData(resp.Error, err.Error())
	}
	return resp.NewData(resp.Integer, int64(l))
}

func HandleLpush(key string, args []resp.DataType) resp.DataType {
	strArgs := make([]string, 0, len(args))
	for _, arg := range args {
		if arg.Is(resp.String) {
			strArgs = append(strArgs, arg.Str)
		} else {
			return resp.NewData(resp.Error, "invalid argument type for 'Lpush' command expects only string")
		}
	}
	l, err := store.DB.Lpush(key, strArgs)
	if err != nil {
		return resp.NewData(resp.Error, err.Error())
	}
	return resp.NewData(resp.Integer, int64(l))
}

func HandleLrange(key string, args []resp.DataType) resp.DataType {
	startIdx, err := args[0].Integer()
	if err != nil {
		return resp.NewData(resp.Error, "expected start index to be an integer for 'rpush' command")
	}
	endIdx, err := args[1].Integer()
	if err != nil {
		return resp.NewData(resp.Error, "expected end index to be an integer for 'rpush' command")
	}
	elems, err := store.DB.Lrange(key, int(startIdx), int(endIdx))
	if err != nil {
		return resp.NewData(resp.Error, err.Error())
	}
	return resp.NewData(resp.Array, elems)
}
