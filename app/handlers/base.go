package handler

import (
	"strconv"
	"strings"

	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

func HandleGet(c *client.Client) {
	args := c.Command.Args
	if len(args) != 1 {
		c.RespChan <- resp.WrongArgs("get")
		return
	}
	key := args[0]
	val := server.Global.Store.Get(key)
	if val == "" {
		c.RespChan <- resp.NewData(resp.NullBulkString)
		return
	}
	c.RespChan <- resp.NewData(resp.BulkString, val)
}

func HandleSet(c *client.Client) {
	args := c.Command.Args
	if len(args) < 2 {
		c.RespChan <- resp.WrongArgs("set")
		return
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
				c.RespChan <- resp.Err("wrong expiry time expected a number")
				return
			}
			if arg == "ex" {
				expiry = exp * 1000
			} else {
				expiry = exp
			}
		default:
			c.RespChan <- resp.Err("unknown argument for 'set' command")
			return
		}
	}

	server.Global.Store.Set(key, val, expiry)
	c.RespChan <- resp.NewData(resp.String, "OK")
}

func HandleCmdIncr(c *client.Client) {
	args := c.Command.Args
	if len(args) != 1 {
		c.RespChan <- resp.WrongArgs("incr")
		return
	}
	key := args[0]

	val, ok := server.Global.Store.Look(key)
	if !ok {
		server.Global.Store.Set(key, "1", 0)
		c.RespChan <- resp.NewData(resp.Integer, 1)
		return
	}

	str, err := store.As[string](val)
	if err != nil {
		c.RespChan <- resp.Err(err.Error())
		return
	}

	if i, err := strconv.Atoi(str); err == nil {
		i += 1
		server.Global.Store.Set(key, strconv.Itoa(i), val.ExpiryAt)
		c.RespChan <- resp.NewData(resp.Integer, i)
	} else {
		c.RespChan <- resp.Err("value is not an integer or out of range")
	}
}

func HandleSubscribe(c *client.Client) {
	args := c.Command.Args
	if len(args) != 1 {
		c.RespChan <- resp.WrongArgs("subscribe")
		return
	}
	channel := args[0]

	client.Chans.Subscribe(c, channel)

	c.RespChan <- resp.NewData(
		resp.Array,
		[]string{"subscribe", channel},
		resp.NewData(resp.Integer, c.SubscriptionCount()),
	)
}

func HandlePublish(c *client.Client) {
	args := c.Command.Args
	if len(args) != 2 {
		c.RespChan <- resp.WrongArgs("publish")
		return
	}
	channel := args[0]
	message := args[1]

	n := client.Chans.Publish(channel, message)
	c.RespChan <- resp.NewData(resp.Integer, n)
}

func HandleUnsubscribe(c *client.Client) {
	args := c.Command.Args
	if len(args) != 1 {
		c.RespChan <- resp.WrongArgs("unsubscribe")
		return
	}
	channel := args[0]
	client.Chans.Unsubscribe(c, channel)
	c.RespChan <- resp.NewData(resp.Array, "unsubscribe", channel, c.SubscriptionCount())
}

func HandleType(c *client.Client) {
	args := c.Command.Args
	if len(args) != 1 {
		c.RespChan <- resp.WrongArgs("type")
		return
	}
	key := args[0]

	t := server.Global.Store.Type(key)
	c.RespChan <- resp.NewData(resp.String, t)
}

func HandleConfig(c *client.Client) {
	args := c.Command.Args
	if len(args) < 2 {
		c.RespChan <- resp.WrongArgs("config")
		return
	}

	sub := strings.ToLower(args[0])

	if sub == "get" {
		configs, err := server.Global.Config.GetConfig(args[1:])
		if err != nil {
			c.RespChan <- resp.Err(err.Error())
			return
		}
		c.RespChan <- resp.NewData(resp.Array, configs)
		return
	}
	c.RespChan <- resp.NewData(resp.Array)
}

func HandleKeys(c *client.Client) {
	args := c.Command.Args
	if len(args) != 1 {
		c.RespChan <- resp.WrongArgs("keys")
		return
	}
	pattern := args[0]

	keys := server.Global.Store.Keys(pattern)
	c.RespChan <- resp.NewData(resp.Array, keys)
}

func HandleAuth(c *client.Client) {
	args := c.Command.Args
	if len(args) != 2 {
		c.RespChan <- resp.WrongArgs("auth")
		return
	}
	res := client.Authenticate(c, args[0], args[1])
	c.RespChan <- res
}

func HandleACL(c *client.Client) {
	args := c.Command.Args
	if len(args) == 0 {
		c.RespChan <- resp.WrongArgs("acl")
		return
	}

	switch strings.ToLower(args[0]) {
	case "whoami":
		if len(args) != 1 {
			c.RespChan <- resp.WrongArgs("acl|whoami")
			return
		}
		c.RespChan <- client.ACL_WHOAMI()

	case "getuser":
		if len(args) != 2 {
			c.RespChan <- resp.WrongArgs("acl|getuser")
			return
		}
		c.RespChan <- client.ACL_GETUSER(args[1])

	case "setuser":
		if len(args) != 3 {
			c.RespChan <- resp.WrongArgs("acl|setuser")
			return
		}
		c.RespChan <- client.ACL_SETUSER(args[1], args[2])

	default:
		c.RespChan <- resp.Err("unknown subcommand '" + args[0] + "'")
	}
}

func HandleInfo(c *client.Client) {
	args := c.Command.Args
	if len(args) == 0 {
		c.RespChan <- resp.WrongArgs("info")
		return
	}

	switch strings.ToLower(args[0]) {
	case "replication":
		if len(args) != 1 {
			c.RespChan <- resp.WrongArgs("info|replication")
			return
		}
		reply := strings.Join([]string{
			"role:" + server.Global.Role,
			"master_replid:" + server.Global.ReplicationId,
			"master_repl_offset:" + strconv.Itoa(server.Global.ReplicationOffset),
		}, "\n")
		c.RespChan <- resp.NewData(resp.BulkString, reply)

	default:
		c.RespChan <- resp.Err("unsupported/unknown subcommand '" + args[0] + "'")
	}
}

func HandleReplconf(c *client.Client) {
	// args := c.Command.Args
	c.RespChan <- resp.NewData(resp.String, "OK")
}

func HandlePsync(c *client.Client) {
	args := c.Command.Args
	if len(args) != 2 {
		c.RespChan <- resp.WrongArgs("psync")
		return
	}

	var sendRDB bool
	str := []string{"FULLRESYNC"}
	if args[0] == "?" {
		str = append(str, server.Global.ReplicationId)
		sendRDB = true
	}
	if args[1] == "-1" {
		str = append(str, strconv.Itoa(server.Global.ReplicationOffset))
		sendRDB = true
	}

	c.RespChan <- resp.NewData(resp.String, strings.Join(str, " "))
	if sendRDB {
		c.SendRDB()
	}
}
