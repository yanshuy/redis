package request

import (
	"fmt"
	"strconv"
	"strings"

	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

func HandleGet(args []resp.Data) resp.Data {
	if len(args) != 1 {
		return resp.WrongArgs("get")
	}
	key := args[0].Str
	if key == "" {
		return resp.Err("key must be a string length > 0")
	}
	if entry := store.RDB.Get(key); entry != nil {
		return resp.NewData(resp.BulkString, entry.Value)
	} else {
		return resp.NewData(resp.BulkString, "-1")
	}
}

func HandleSet(args []resp.Data) resp.Data {
	if len(args) < 2 {
		return resp.WrongArgs("set")
	}
	key := args[0].Str
	val := args[1].Str
	if key == "" || val == "" {
		return resp.Err("key, val must be a string length > 0")
	}
	var expiry int64
	if len(args) >= 4 {
		arg := args[2].Str
		switch strings.ToLower(arg) {
		case "px", "ex":
			exp, err := args[3].Integer()
			if err != nil {
				return resp.Err("wrong expiry time expected a number")
			}
			if arg == "ex" {
				expiry = exp * 1000
			} else {
				expiry = exp
			}

		default:
			return resp.Err("unknown argument for 'set' command")
		}
	}

	store.RDB.Set(key, val, expiry)
	return resp.NewData(resp.String, "OK")
}

func HandleRpush(args []resp.Data) resp.Data {
	if len(args) < 2 {
		return resp.WrongArgs("rpush")
	}
	key := args[0].Str
	if key == "" {
		return resp.Err("key must be a string length > 0")
	}
	strArgs := make([]string, 0, len(args)-1)
	for _, arg := range args[1:] {
		if arg.Is(resp.BulkString) {
			strArgs = append(strArgs, arg.Str)
		} else {
			return resp.Err("invalid argument type for 'rpush' command expects only string")
		}
	}
	l, err := store.RDB.Rpush(key, strArgs)
	if err != nil {
		return resp.Err(err.Error())
	}
	return resp.NewData(resp.Integer, int64(l))
}

func HandleLpush(args []resp.Data) resp.Data {
	if len(args) < 2 {
		return resp.WrongArgs("lpush")
	}
	key := args[0].Str
	if key == "" {
		return resp.Err("key must be a string length > 0")
	}
	strArgs := make([]string, 0, len(args)-1)
	for _, arg := range args[1:] {
		if arg.Is(resp.BulkString) {
			strArgs = append(strArgs, arg.Str)
		} else {
			return resp.Err("invalid argument type for 'lpush' command expects only string")
		}
	}
	l, err := store.RDB.Lpush(key, strArgs)
	if err != nil {
		return resp.Err(err.Error())
	}
	return resp.NewData(resp.Integer, int64(l))
}

func HandleLpop(args []resp.Data) resp.Data {
	if len(args) < 1 || len(args) > 2 {
		return resp.WrongArgs("lpop")
	}
	key := args[0].Str
	if key == "" {
		return resp.Err("key must be a string length > 0")
	}
	pops := 1
	if len(args) == 2 {
		p, err := args[1].Integer()
		if err != nil {
			return resp.Err("2nd argument must be a integer")
		}
		pops = int(p)
	}
	l, err := store.RDB.Lpop(key, pops)
	if err != nil {
		return resp.Err(err.Error())
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

func HandleLlen(args []resp.Data) resp.Data {
	if len(args) != 1 {
		return resp.WrongArgs("llen")
	}
	key := args[0].Str
	if key == "" {
		return resp.Err("key must be a string length > 0")
	}
	l, err := store.RDB.Llen(key)
	if err != nil {
		return resp.Err(err.Error())
	}
	return resp.NewData(resp.Integer, int64(l))
}

func HandleLrange(args []resp.Data) resp.Data {
	if len(args) != 3 {
		return resp.WrongArgs("lrange")
	}
	key := args[0].Str
	if key == "" {
		return resp.Err("key, val must be a string length > 0")
	}
	startIdx, err := args[1].Integer()
	if err != nil {
		return resp.Err("expected start index to be an integer for 'lrange' command")
	}
	endIdx, err := args[2].Integer()
	if err != nil {
		return resp.Err("expected end index to be an integer for 'lrange' command")
	}
	elems, err := store.RDB.Lrange(key, int(startIdx), int(endIdx))
	if err != nil {
		return resp.Err(err.Error())
	}
	return resp.NewData(resp.Array, elems)
}

func HandleBlpop(args []resp.Data) resp.Data {
	if len(args) != 2 {
		return resp.WrongArgs("blpop")
	}
	key := args[0].Str
	if key == "" {
		return resp.Err("key, val must be a string length > 0")
	}
	timeout_s, err := strconv.ParseFloat(args[1].Str, 10)
	if err != nil {
		return resp.Err("expected 2 argument to be an number for 'blpop' command")
	}

	msgChan, err := store.RDB.Blpop(key, timeout_s)
	if err != nil {
		return resp.Err(err.Error())
	}
	s := <-msgChan
	if s == "" {
		return resp.NewData(resp.Array)
	}
	return resp.NewData(resp.Array, []string{key, s})
}

func HandleType(args []resp.Data) resp.Data {
	if len(args) != 1 {
		return resp.WrongArgs("type")
	}
	key := args[0].Str
	if key == "" {
		return resp.Err("key, val must be a string length > 0")
	}
	t := store.RDB.Type(key)
	return resp.NewData(resp.String, t)
}

func HandleXadd(args []resp.Data) resp.Data {
	if len(args) < 2 {
		return resp.WrongArgs("xadd")
	}
	key := args[0].Str
	stream_key := args[1].Str
	if key == "" || stream_key == "" {
		return resp.Err("key, val must be a string length > 0")
	}
	rest := args[2:]
	key_vals := make([]string, 0, len(args[2:]))
	for i := 0; i < len(rest); i += 2 {
		key := rest[i]
		if key.Str == "" {
			return resp.Err("key, val must be a string length > 0")
		}
		if i+1 > len(rest) {
			return resp.Err(fmt.Sprintf("no value for the key %s specified", key.Str))
		}
		val := rest[i+1]
		if val.Str == "" {
			return resp.Err("key, val must be a string length > 0")
		}
		key_vals = append(key_vals, key.Str, val.Str)
	}
	s, err := store.RDB.Xadd(key, stream_key, key_vals)
	if err != nil {
		return resp.Err(err.Error())
	}
	return resp.NewData(resp.BulkString, s)
}

func HandleXrange(args []resp.Data) resp.Data {
	if len(args) != 3 {
		return resp.WrongArgs("xrange")
	}
	key := args[0].Str
	if key == "" {
		return resp.Err("key, val must be a string length > 0")
	}
	startStr := args[1].Str
	endStr := args[2].Str

	entries, err := store.RDB.XRange(key, startStr, endStr)
	if err != nil {
		return resp.Err(err.Error())
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

func HandleXread(args []resp.Data) resp.Data {
	return resp.NewData(resp.Array)
}

func HandleConfig(args []resp.Data) resp.Data {
	if len(args) < 2 {
		return resp.WrongArgs("config")
	}

	sub := strings.ToLower(args[0].Str)

	if sub == "get" {
		strArgs := make([]string, 0, len(args))
		for _, arg := range args[1:] {
			if arg.Str == "" {
				return resp.Err("argument must be a string length > 0")
			}
			strArgs = append(strArgs, arg.Str)
		}

		configs, err := store.RDB.ConfigGet(strArgs)
		if err != nil {
			return resp.Err(err.Error())
		}
		return resp.NewData(resp.Array, configs)
	}
	return resp.NewData(resp.Array)
}

func HandleKeys(args []resp.Data) resp.Data {
	if len(args) != 1 {
		return resp.WrongArgs("keys")
	}
	pattern := args[0].Str
	if pattern == "" {
		return resp.Err("key, val must be a string length > 0")
	}
	keys := store.RDB.Keys(pattern)
	return resp.NewData(resp.Array, keys)
}

func HandleSubscribe(c *client.Client, args []resp.Data) resp.Data {
	if len(args) != 1 {
		return resp.WrongArgs("subscribe")
	}
	channel := args[0].Str
	if channel == "" {
		return resp.Err("channel name must be a string length > 0")
	}

	client.Chans.Subscribe(channel, c)

	return resp.NewData(
		resp.Array,
		[]string{"subscribe", channel},
		resp.NewData(resp.Integer, int64(c.SubscriptionCount())),
	)
}

func HandlePublish(args []resp.Data) resp.Data {
	if len(args) != 2 {
		return resp.WrongArgs("publish")
	}
	channel := args[0].Str
	if channel == "" {
		return resp.Err("channel name must be a string length > 0")
	}
	message := args[1].Str
	if channel == "" {
		return resp.Err("message name must be a string length > 0")
	}

	n := client.Chans.Publish(channel, message)
	return resp.NewData(resp.Integer, int64(n))
}

func HandleUnsubscribe(c *client.Client, args []resp.Data) resp.Data {
	if len(args) != 1 {
		return resp.WrongArgs("unsubscribe")
	}
	channel := args[0].Str
	if channel == "" {
		return resp.Err("channel name must be a string length > 0")
	}
	client.Chans.Unsubscribe(channel, c)
	return resp.NewData(resp.Array, "unsubscribe", channel, int64(c.SubscriptionCount()))
}

func HandleCmdIncr(args []resp.Data) resp.Data {
	if len(args) != 1 {
		return resp.WrongArgs("incr")
	}
	key := args[0].Str
	if key == "" {
		return resp.Err("key must be a string length > 0")
	}
	if entry := store.RDB.Get(key); entry != nil {
		if i, err := strconv.Atoi(entry.Value); err == nil {
			i += 1
			store.RDB.Set(key, strconv.Itoa(i), entry.ExpiryAt)
			return resp.NewData(resp.Integer, int64(i))
		} else {
			return resp.Err("value is not an integer or out of range")
		}
	} else {
		store.RDB.Set(key, "1", 0)
		return resp.NewData(resp.Integer, int64(1))
	}
}

func HandleACL(args []resp.Data) resp.Data {
	if len(args) == 0 {
		return resp.WrongArgs("acl")
	}

	switch strings.ToLower(args[0].Str) {
	case "whoami":
		if len(args) != 1 {
			return resp.WrongArgs("acl|whoami")
		}
		return client.ACL_WHOAMI()

	case "getuser":
		if len(args) != 2 {
			return resp.WrongArgs("acl|getuser")
		}

		if args[1].Type != resp.BulkString {
			return resp.Err("username must be a bulk string")
		}

		return client.ACL_GETUSER(args[1].Str)

	case "setuser":
		if len(args) != 3 {
			return resp.WrongArgs("acl|setuser")
		}

		if args[1].Type != resp.BulkString {
			return resp.Err("username must be a bulk string")
		}

		if args[2].Type != resp.BulkString {
			return resp.Err("rule must be a bulk string")
		}

		return client.ACL_SETUSER(args[1].Str, args[2].Str)

	default:
		return resp.Err("unknown subcommand '" + args[0].Str + "'")
	}
}

func HandleAuth(c *client.Client, args []resp.Data) resp.Data {
	if len(args) != 2 {
		return resp.WrongArgs("auth")
	}

	if args[0].Type != resp.BulkString {
		return resp.Err("username must be a bulk string")
	}

	if args[1].Type != resp.BulkString {
		return resp.Err("password must be a bulk string")
	}

	return client.Authenticate(c, args[0].Str, args[1].Str)
}

func HandleExec(c *client.Client) resp.Data {
	if c.InMulti {
		c.InMulti = false
	} else {
		return resp.Err("EXEC without MULTI")
	}
	if c.CASDirty {
		c.CASDirty = false
		return resp.NewData(resp.Array)
	}

	queued := c.QueuedCmds
	c.QueuedCmds = nil
	respArr := make([]resp.Data, 0, len(queued))
	for _, cmd := range queued {
		res := HandleCmd(c, Command{cmd.Name, cmd.Args})
		respArr = append(respArr, res)
	}
	c.CASDirty = false
	return resp.NewData(resp.Array, respArr)
}

func HandleDiscard(c *client.Client) resp.Data {
	if c.InMulti {
		c.InMulti = false
	} else {
		return resp.Err("DISCARD without MULTI")
	}
	c.QueuedCmds = nil
	c.CASDirty = false
	return resp.NewData(resp.String, "OK")
}

func HandleWatch(c *client.Client, args []resp.Data) resp.Data {
	if c.InMulti {
		return resp.Err("WATCH inside MULTI is not allowed")
	}
	for _, keyData := range args {
		if keyData.Type != resp.BulkString {
			return resp.Err("key must be a bulkstring")
		}
		key := keyData.Str
		c.WatchKeys = append(c.WatchKeys, key)
		store.RDB.WatchedKeys[key] = append(store.RDB.WatchedKeys[key], c)
	}
	return resp.NewData(resp.String, "OK")
}

func HandleUnWatch(c *client.Client) resp.Data {
	for _, key := range c.WatchKeys {
		clients := store.RDB.WatchedKeys[key]
		clients = filter(clients, c)

		if len(clients) == 0 {
			delete(store.RDB.WatchedKeys, key)
		} else {
			store.RDB.WatchedKeys[key] = clients
		}
	}
	c.WatchKeys = nil
	c.CASDirty = false
	return resp.NewData(resp.String, "OK")
}

func filter[T comparable](array []T, elem T) []T {
	for i, item := range array {
		if elem == item {
			return append(array[:i], array[i+1:]...)
		}
	}
	return array
}
