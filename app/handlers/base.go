package handler

import (
	"strconv"
	"strings"

	resp "github.com/codecrafters-io/redis-starter-go/app/Resp"
	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

func HandleGet(c *client.Client) resp.Data {
	args := c.Command.Args
	if len(args) != 1 {
		return resp.WrongArgs("get")
	}
	key := args[0]
	val := s.Store.Get(key)
	if val == "" {
		return resp.NewData(resp.NullBulkString)
	}
	return resp.NewData(resp.BulkString, val)
}

func HandleSet(c *client.Client) resp.Data {
	args := c.Command.Args
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

	s.Store.Set(key, val, expiry)
	s.Propagate(c.Command)
	return resp.NewData(resp.String, "OK")
}

func HandleCmdIncr(c *client.Client) resp.Data {
	args := c.Command.Args
	if len(args) != 1 {
		return resp.WrongArgs("incr")
	}
	key := args[0]

	val, ok := s.Store.Look(key)
	if !ok {
		s.Store.Set(key, "1", 0)
		return resp.NewData(resp.Integer, 1)
	}

	str, err := store.As[string](val)
	if err != nil {
		return resp.Err(err.Error())
	}

	if i, err := strconv.Atoi(str); err == nil {
		i += 1
		s.Store.Set(key, strconv.Itoa(i), val.ExpiryAt)
		return resp.NewData(resp.Integer, i)
	} else {
		return resp.Err("value is not an integer or out of range")
	}
}

func HandleSubscribe(c *client.Client) resp.Data {
	args := c.Command.Args
	if len(args) != 1 {
		return resp.WrongArgs("subscribe")
	}
	channel := args[0]

	client.Chans.Subscribe(c, channel)

	return resp.NewData(
		resp.Array,
		[]string{"subscribe", channel},
		resp.NewData(resp.Integer, c.SubscriptionCount()),
	)
}

func HandlePublish(c *client.Client) resp.Data {
	args := c.Command.Args
	if len(args) != 2 {
		return resp.WrongArgs("publish")
	}
	channel := args[0]
	message := args[1]

	n := client.Chans.Publish(channel, message)
	return resp.NewData(resp.Integer, n)
}

func HandleUnsubscribe(c *client.Client) resp.Data {
	args := c.Command.Args
	if len(args) != 1 {
		return resp.WrongArgs("unsubscribe")
	}
	channel := args[0]
	client.Chans.Unsubscribe(c, channel)
	return resp.NewData(resp.Array, "unsubscribe", channel, c.SubscriptionCount())
}

func HandleType(c *client.Client) resp.Data {
	args := c.Command.Args
	if len(args) != 1 {
		return resp.WrongArgs("type")
	}
	key := args[0]

	t := s.Store.Type(key)
	return resp.NewData(resp.String, t)
}

func HandleConfig(c *client.Client) resp.Data {
	args := c.Command.Args
	if len(args) < 2 {
		return resp.WrongArgs("config")
	}

	sub := strings.ToLower(args[0])

	if sub == "get" {
		configs, err := s.Config.GetConfig(args[1:])
		if err != nil {
			return resp.Err(err.Error())
		}
		return resp.NewData(resp.Array, configs)
	}
	return resp.NewData(resp.Array)
}

func HandleKeys(c *client.Client) resp.Data {
	args := c.Command.Args
	if len(args) != 1 {
		return resp.WrongArgs("keys")
	}
	pattern := args[0]

	keys := s.Store.Keys(pattern)
	return resp.NewData(resp.Array, keys)
}

func HandleAuth(c *client.Client) resp.Data {
	args := c.Command.Args
	if len(args) != 2 {
		return resp.WrongArgs("auth")
	}
	res := client.Authenticate(c, args[0], args[1])
	return res
}

func HandleACL(c *client.Client) resp.Data {
	args := c.Command.Args
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

func HandleInfo(c *client.Client) resp.Data {
	args := c.Command.Args
	if len(args) == 0 {
		return resp.WrongArgs("info")
	}

	switch strings.ToLower(args[0]) {
	case "replication":
		if len(args) != 1 {
			return resp.WrongArgs("info|replication")
		}
		reply := strings.Join([]string{
			"role:" + s.Role.String(),
			"master_replid:" + s.ReplicationId,
			"master_repl_offset:" + strconv.Itoa(s.ReplicationOffset),
		}, "\n")
		return resp.NewData(resp.BulkString, reply)

	default:
		return resp.Err("unsupported/unknown subcommand '" + args[0] + "'")
	}
}
