package handler

import (
	"strconv"
	"strings"

	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

func (h Handler) HandleGet(args []string) resp.Data {
	if len(args) != 1 {
		return resp.WrongArgs("get")
	}
	key := args[0]
	if key == "" {
		return resp.Err("key must be a string length > 0")
	}
	val := h.store.Get(key)
	if val == "" {
		return resp.NewData(resp.NullBulkString)
	}
	return resp.NewData(resp.BulkString, val)
}

func (h Handler) HandleSet(args []string) resp.Data {
	if len(args) < 2 {
		return resp.WrongArgs("set")
	}
	key := args[0]
	val := args[1]

	var expiry int64
	if len(args) >= 4 {
		arg := args[2]
		switch strings.ToLower(arg) {
		case "px", "ex":
			exp, err := strconv.ParseInt(args[3], 10, 64)
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

	h.store.Set(key, val, expiry)
	return resp.NewData(resp.String, "OK")
}

func (h Handler) HandleCmdIncr(args []string) resp.Data {
	if len(args) != 1 {
		return resp.WrongArgs("incr")
	}
	key := args[0]

	val, ok := h.store.Look(key)
	if !ok {
		h.store.Set(key, "1", 0)
		return resp.NewData(resp.Integer, 1)
	}

	str, err := store.As[string](val)
	if err != nil {
		return resp.Err(err.Error())
	}

	if i, err := strconv.Atoi(str); err == nil {
		i += 1
		h.store.Set(key, strconv.Itoa(i), val.ExpiryAt)
		return resp.NewData(resp.Integer, i)
	} else {
		return resp.Err("value is not an integer or out of range")
	}
}

func (h Handler) HandleSubscribe(args []string) resp.Data {
	if len(args) != 1 {
		return resp.WrongArgs("subscribe")
	}
	channel := args[0]

	client.Chans.Subscribe(h.client, channel)

	return resp.NewData(
		resp.Array,
		[]string{"subscribe", channel},
		resp.NewData(resp.Integer, h.client.SubscriptionCount()),
	)
}

func (h Handler) HandlePublish(args []string) resp.Data {
	if len(args) != 2 {
		return resp.WrongArgs("publish")
	}
	channel := args[0]
	message := args[1]

	n := client.Chans.Publish(channel, message)
	return resp.NewData(resp.Integer, n)
}

func (h Handler) HandleUnsubscribe(args []string) resp.Data {
	if len(args) != 1 {
		return resp.WrongArgs("unsubscribe")
	}
	channel := args[0]
	client.Chans.Unsubscribe(h.client, channel)
	return resp.NewData(resp.Array, "unsubscribe", channel, h.client.SubscriptionCount())
}

func (h Handler) HandleType(args []string) resp.Data {
	if len(args) != 1 {
		return resp.WrongArgs("type")
	}
	key := args[0]

	t := h.store.Type(key)
	return resp.NewData(resp.String, t)
}

func (h Handler) HandleConfig(args []string) resp.Data {
	if len(args) < 2 {
		return resp.WrongArgs("config")
	}

	sub := strings.ToLower(args[0])

	if sub == "get" {
		configs, err := h.store.ConfigGet(args[1:])
		if err != nil {
			return resp.Err(err.Error())
		}
		return resp.NewData(resp.Array, configs)
	}
	return resp.NewData(resp.Array)
}

func (h Handler) HandleKeys(args []string) resp.Data {
	if len(args) != 1 {
		return resp.WrongArgs("keys")
	}
	pattern := args[0]

	keys := h.store.Keys(pattern)
	return resp.NewData(resp.Array, keys)
}

func (h Handler) HandleAuth(args []string) resp.Data {
	if len(args) != 2 {
		return resp.WrongArgs("auth")
	}
	return client.Authenticate(h.client, args[0], args[1])
}

func (h Handler) HandleACL(args []string) resp.Data {
	if len(args) == 0 {
		return resp.WrongArgs("acl")
	}

	switch strings.ToLower(args[0]) {
	case "whoami":
		if len(args) != 1 {
			return resp.WrongArgs("acl|whoami")
		}
		return client.ACL_WHOAMI()

	case "getuser":
		if len(args) != 2 {
			return resp.WrongArgs("acl|getuser")
		}

		return client.ACL_GETUSER(args[1])

	case "setuser":
		if len(args) != 3 {
			return resp.WrongArgs("acl|setuser")
		}

		return client.ACL_SETUSER(args[1], args[2])

	default:
		return resp.Err("unknown subcommand '" + args[0] + "'")
	}
}
