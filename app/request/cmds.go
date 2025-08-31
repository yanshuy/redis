package request

import (
	"fmt"
	"strconv"
	"strings"

	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

func HandleCmdGet(args []resp.DataType) resp.DataType {
	if len(args) != 1 {
		return resp.NewData(resp.Error, "wrong number of arguments for 'get' command")
	}
	key := args[0].Str
	if key == "" {
		return resp.NewData(resp.Error, "key must be a string length > 0")
	}
	if val, ok := store.RDB.Get(key); ok {
		return resp.NewData(resp.BulkString, val)
	} else {
		return resp.NewData(resp.BulkString, "-1")
	}
}

func HandleCmdSet(args []resp.DataType) resp.DataType {
	if len(args) < 2 {
		return resp.NewData(resp.Error, "wrong number of arguments for 'set' command")
	}
	key := args[0].Str
	val := args[1].Str
	if key == "" || val == "" {
		return resp.NewData(resp.Error, "key, val must be a string length > 0")
	}
	var expiry int64
	if len(args) >= 4 {
		arg := args[2].Str
		switch strings.ToLower(arg) {
		case "px", "ex":
			exp, err := args[3].Integer()
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

	store.RDB.Set(key, val, expiry)
	return resp.NewData(resp.String, "OK")
}

func HandleRpush(args []resp.DataType) resp.DataType {
	if len(args) < 2 {
		return resp.NewData(resp.Error, "wrong number of arguments for 'rpush' command")
	}
	key := args[0].Str
	if key == "" {
		return resp.NewData(resp.Error, "key must be a string length > 0")
	}
	strArgs := make([]string, 0, len(args)-1)
	for _, arg := range args[1:] {
		if arg.Is(resp.BulkString) {
			strArgs = append(strArgs, arg.Str)
		} else {
			return resp.NewData(resp.Error, "invalid argument type for 'rpush' command expects only string")
		}
	}
	l, err := store.RDB.Rpush(key, strArgs)
	if err != nil {
		return resp.NewData(resp.Error, err.Error())
	}
	return resp.NewData(resp.Integer, int64(l))
}

func HandleLpush(args []resp.DataType) resp.DataType {
	if len(args) < 2 {
		return resp.NewData(resp.Error, "wrong number of arguments for 'lpush' command")
	}
	key := args[0].Str
	if key == "" {
		return resp.NewData(resp.Error, "key must be a string length > 0")
	}
	strArgs := make([]string, 0, len(args)-1)
	for _, arg := range args[1:] {
		if arg.Is(resp.BulkString) {
			strArgs = append(strArgs, arg.Str)
		} else {
			return resp.NewData(resp.Error, "invalid argument type for 'lpush' command expects only string")
		}
	}
	l, err := store.RDB.Lpush(key, strArgs)
	if err != nil {
		return resp.NewData(resp.Error, err.Error())
	}
	return resp.NewData(resp.Integer, int64(l))
}

func HandleLpop(args []resp.DataType) resp.DataType {
	if len(args) < 1 || len(args) > 2 {
		return resp.NewData(resp.Error, "wrong number of arguments for 'lpop' command")
	}
	key := args[0].Str
	if key == "" {
		return resp.NewData(resp.Error, "key must be a string length > 0")
	}
	pops := 1
	if len(args) == 2 {
		p, err := args[1].Integer()
		if err != nil {
			return resp.NewData(resp.Error, "2nd argument must be a integer")
		}
		pops = int(p)
	}
	l, err := store.RDB.Lpop(key, pops)
	if err != nil {
		return resp.NewData(resp.Error, err.Error())
	}
	switch len(l) {
	case 0:
		return resp.NewData(resp.BulkString, "-1")
	case 1:
		return resp.NewData(resp.BulkString, l[0])
	default:
		return resp.NewData(resp.Array, l)
	}
}

func HandleLlen(args []resp.DataType) resp.DataType {
	if len(args) != 1 {
		return resp.NewData(resp.Error, "wrong number of arguments for 'llen' command")
	}
	key := args[0].Str
	if key == "" {
		return resp.NewData(resp.Error, "key must be a string length > 0")
	}
	l, err := store.RDB.Llen(key)
	if err != nil {
		return resp.NewData(resp.Error, err.Error())
	}
	return resp.NewData(resp.Integer, int64(l))
}

func HandleLrange(args []resp.DataType) resp.DataType {
	if len(args) != 3 {
		return resp.NewData(resp.Error, "wrong number of arguments for 'lrange' command")
	}
	key := args[0].Str
	if key == "" {
		return resp.NewData(resp.Error, "key, val must be a string length > 0")
	}
	startIdx, err := args[1].Integer()
	if err != nil {
		return resp.NewData(resp.Error, "expected start index to be an integer for 'lrange' command")
	}
	endIdx, err := args[2].Integer()
	if err != nil {
		return resp.NewData(resp.Error, "expected end index to be an integer for 'lrange' command")
	}
	elems, err := store.RDB.Lrange(key, int(startIdx), int(endIdx))
	if err != nil {
		return resp.NewData(resp.Error, err.Error())
	}
	return resp.NewData(resp.Array, elems)
}

func HandleBlpop(args []resp.DataType) resp.DataType {
	if len(args) != 2 {
		return resp.NewData(resp.Error, "wrong number of arguments for 'blpop' command")
	}
	key := args[0].Str
	if key == "" {
		return resp.NewData(resp.Error, "key, val must be a string length > 0")
	}
	timeout_s, err := strconv.ParseFloat(args[1].Str, 10)
	if err != nil {
		return resp.NewData(resp.Error, "expected 2 argument to be an number for 'blpop' command")
	}

	msgChan, err := store.RDB.Blpop(key, timeout_s)
	if err != nil {
		return resp.NewData(resp.Error, err.Error())
	}
	s := <-msgChan
	if s == "" {
		return resp.NewData(resp.Array, nil)
	}
	return resp.NewData(resp.Array, []string{key, s})
}

func HandleType(args []resp.DataType) resp.DataType {
	if len(args) != 1 {
		return resp.NewData(resp.Error, "wrong number of arguments for 'blpop' command")
	}
	key := args[0].Str
	if key == "" {
		return resp.NewData(resp.Error, "key, val must be a string length > 0")
	}
	t := store.RDB.Type(key)
	return resp.NewData(resp.String, t)
}

func HandleXadd(args []resp.DataType) resp.DataType {
	if len(args) < 2 {
		return resp.NewData(resp.Error, "wrong number of arguments for 'xadd' command")
	}
	key := args[0].Str
	stream_key := args[1].Str
	if key == "" || stream_key == "" {
		return resp.NewData(resp.Error, "key, val must be a string length > 0")
	}
	rest := args[2:]
	key_vals := make([]string, 0, len(args[2:]))
	for i := 0; i < len(rest); i += 2 {
		key := rest[i]
		if key.Str == "" {
			return resp.NewData(resp.Error, "key, val must be a string length > 0")
		}
		if i+1 > len(rest) {
			return resp.NewData(resp.Error, fmt.Sprintf("no value for the key %s specified", key.Str))
		}
		val := rest[i+1]
		if val.Str == "" {
			return resp.NewData(resp.Error, "key, val must be a string length > 0")
		}
		key_vals = append(key_vals, key.Str, val.Str)
	}
	s, err := store.RDB.Xadd(key, stream_key, key_vals)
	if err != nil {
		return resp.NewData(resp.Error, err.Error())
	}
	return resp.NewData(resp.BulkString, s)
}

func HandleXrange(args []resp.DataType) resp.DataType {
	if len(args) != 3 {
		return resp.NewData(resp.Error, "wrong number of arguments for 'xadd' command")
	}
	key := args[0].Str
	if key == "" {
		return resp.NewData(resp.Error, "key, val must be a string length > 0")
	}
	startStr := args[1].Str
	endStr := args[2].Str

	entries, err := store.RDB.XRange(key, startStr, endStr)
	if err != nil {
		return resp.NewData(resp.Error, err.Error())
	}

	res := resp.NewData(resp.Array)
	for _, entry := range entries {
		id := resp.NewData(resp.BulkString, fmt.Sprintf("%d-%d", entry.Id.MS, entry.Id.Seq))
		fields := resp.NewData(resp.Array, entry.Fields)

		entryArr := resp.NewData(resp.Array)
		entryArr.Arr = append(entryArr.Arr, id, fields)
		res.Arr = append(res.Arr, entryArr)
	}
	return res
}

func HandleXread(args []resp.DataType) resp.DataType {
	return resp.NewData(resp.Array)
}

func HandleConfig(args []resp.DataType) resp.DataType {
	if len(args) < 2 {
		return resp.NewData(resp.Error, "wrong number of arguments for 'config' command")
	}
	sub := strings.ToLower(args[0].Str)
	if sub == "get" {
		strArgs := make([]string, 0, len(args))
		for _, arg := range args[1:] {
			if arg.Str == "" {
				return resp.NewData(resp.Error, "argument must be a string length > 0")
			}
			strArgs = append(strArgs, arg.Str)
		}
		configs, err := store.RDB.ConfigGet(strArgs)
		if err != nil {
			return resp.NewData(resp.Error, err.Error())
		}
		return resp.NewData(resp.Array, configs)
	}
	return resp.NewData(resp.Array, nil)
}

func HandleKeys(args []resp.DataType) resp.DataType {
	if len(args) != 1 {
		return resp.NewData(resp.Error, "wrong number of arguments for 'config' command")
	}
	pattern := args[0].Str
	if pattern == "" {
		return resp.NewData(resp.Error, "key, val must be a string length > 0")
	}
	keys := store.RDB.Keys(pattern)
	return resp.NewData(resp.Array, keys)
}

func HandleSubscribe(args []resp.DataType) resp.DataType {
	if len(args) != 1 {
		return resp.NewData(resp.Error, "wrong number of arguments for 'config' command")
	}
	channel := args[0].Str
	if channel == "" {
		return resp.NewData(resp.Error, "channel name must be a string length > 0")
	}

	Chans.subscribe(channel)
	return resp.NewData(resp.Array, []string{"subscribe", channel}, resp.NewData(resp.Integer, int64(Chans.subscribers(channel))))
}

func HandlePublish(args []resp.DataType) resp.DataType {
	if len(args) != 2 {
		return resp.NewData(resp.Error, "wrong number of arguments for 'config' command")
	}
	channel := args[0].Str
	if channel == "" {
		return resp.NewData(resp.Error, "channel name must be a string length > 0")
	}

	return resp.NewData(resp.Integer, int64(Chans.subscribers(channel)))
}
