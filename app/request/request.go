package request

import (
	"fmt"
	"io"
	"log"
	"strings"

	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

type Command struct {
	name string
	args []resp.Data
}

func ReadAndHandleRequest(conn io.ReadWriteCloser) (n int, err error) {
	c := client.NewClient(conn)
	// TODO: request > 1024
	b := make([]byte, 1024)
	bLen := 0
	for {
		n, err := conn.Read(b[bLen:])
		if n > 0 {
			bLen += n
			r, o, err := resp.Parse(b[:bLen])
			if err != nil {
				return bLen, err
			}
			if o > 0 {
				err := HandleRequests(c, r)
				if err != nil {
					return bLen, err
				}
				copy(b, b[o:bLen])
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

func HandleRequests(c *client.Client, requests []resp.Data) (err error) {
	for _, req := range requests {
		res := execute(c, req)

		resBytes := res.ToResponse()
		if _, err := c.Write(resBytes); err != nil {
			return err
		}
		if c.Quit {
			return c.Close()
		}
	}
	return nil
}

func execute(c *client.Client, req resp.Data) resp.Data {
	cmd := Command{}

	switch req.Type {
	case resp.Array:
		if len(req.Arr) == 0 {
			return resp.Err("empty command")
		}
		first := req.Arr[0]
		if first.Type != resp.BulkString && first.Type != resp.String {
			return resp.Err("invalid command")
		}
		cmd.name = strings.ToLower(first.Str)
		cmd.args = req.Arr[1:]

	case resp.String, resp.BulkString:
		cmd.name = req.Str
	default:
		return resp.Err("invalid command")
	}

	if cmd.name == "quit" {
		c.Quit = true
		return resp.NewData(resp.String, "OK")
	}

	if !c.IsAuthenticated() {
		switch cmd.name {
		case "auth", "hello", "ping":
		default:
			return resp.NoAuth()
		}
	}

	if c.InMulti {
		switch cmd.name {
		case "exec", "discard", "watch":
		default:
			c.QueuedCmds = append(c.QueuedCmds, client.Command{
				Name: cmd.name,
				Args: cmd.args,
			})
			return resp.NewData(resp.String, "QUEUED")
		}
	}

	return HandleCmd(c, cmd)
}

func HandleCmd(c *client.Client, cmd Command) resp.Data {
	args := cmd.args

	if c.InSubscribeMode() {
		switch cmd.name {
		case "ping":
			return resp.NewData(resp.Array, "pong", "")
		case "subscribe", "unsubscribe":
		default:
			return resp.Err(fmt.Sprintf("Can't execute '%s': only (P|S)SUBSCRIBE / (P|S)UNSUBSCRIBE / PING / QUIT / RESET are allowed in this context", cmd.name))
		}
	}

	switch cmd.name {
	case "ping":
		return resp.NewData(resp.String, "PONG")

	case "echo":
		if len(args) != 1 {
			return resp.WrongArgs("echo")
		}
		return args[0]

	case "get":
		return HandleGet(args)

	case "set":
		return HandleSet(args)

	case "acl":
		return HandleACL(args)

	case "auth":
		return HandleAuth(c, args)

	case "rpush":
		return HandleRpush(args)

	case "lpush":
		return HandleLpush(args)

	case "llen":
		return HandleLlen(args)

	case "lpop":
		return HandleLpop(args)

	case "lrange":
		return HandleLrange(args)

	case "blpop":
		return HandleBlpop(args)

	case "zadd":
		return HandleZadd(args)

	case "zrank":
		return HandleZrank(args)

	case "zrange":
		return HandleZrange(args)

	case "zcard":
		return HandleZcard(args)

	case "type":
		return HandleType(args)

	case "keys":
		return HandleKeys(args)

	case "xadd":
		return HandleXadd(args)

	case "xrange":
		return HandleXrange(args)

	case "xread":
		return HandleXread(args)

	case "config":
		return HandleConfig(args)

	case "save":
		err := store.RDB.SaveRDBSnapshot()
		if err != nil {
			log.Println(err)
			return resp.Err("save failed")
		}
		return resp.NewData(resp.String, "OK")

	case "subscribe":
		return HandleSubscribe(c, args)

	case "unsubscribe":
		return HandleUnsubscribe(c, args)

	case "publish":
		return HandlePublish(args)

	case "incr":
		return HandleCmdIncr(args)

	case "multi":
		c.InMulti = true
		return resp.NewData(resp.String, "OK")

	case "exec":
		return HandleExec(c)

	case "discard":
		return HandleDiscard(c)

	case "watch":
		return HandleWatch(c, args)

	case "unwatch":
		return HandleUnWatch(c)

	default:
		msg := fmt.Sprintf("unknown command `%s`", cmd.name)
		return resp.Err(msg)
	}
}
