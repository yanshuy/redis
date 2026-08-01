package request

import (
	"fmt"
	"io"
	"log"
	"strings"

	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

type Client struct {
	conn          io.WriteCloser
	subscriptions map[string]chan string
	authAsUser    *User
	done          chan struct{}
	quit          bool
}

func (c *Client) inSubscribeMode() bool {
	return len(c.subscriptions) > 0
}

func (c *Client) Close() error {
	for channel := range c.subscriptions {
		Chans.unsubscribe(channel, c)
	}
	return c.conn.Close()
}

func (c *Client) isAuthenticated() bool {
	return c.authAsUser != nil
}

func NewClient(conn io.WriteCloser) *Client {
	var user *User
	if DefaultUser.flags.Contains("nopass") {
		user = DefaultUser
	}
	doneChan := make(chan struct{})

	return &Client{
		conn:       conn,
		authAsUser: user,
		done:       doneChan,
	}
}

func ReadAndHandleRequest(conn io.ReadWriteCloser) (n int, err error) {
	c := NewClient(conn)
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
				err := c.HandleRequest(conn, r)
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

func (c *Client) HandleRequest(w io.Writer, requests []resp.Data) (err error) {
	for _, req := range requests {
		res := c.execute(req)
		resBytes := res.ToResponse()
		if _, err := w.Write(resBytes); err != nil {
			return err
		}
		if c.quit {
			return c.Close()
		}
	}
	return nil
}

func (c *Client) execute(req resp.Data) resp.Data {
	var cmd string
	var args []resp.Data

	switch req.Type {
	case resp.Array:
		if len(req.Arr) == 0 {
			return Err("empty command")
		}
		first := req.Arr[0]
		if first.Type != resp.BulkString && first.Type != resp.String {
			return Err("invalid command")
		}
		cmd = first.Str
		args = req.Arr[1:]

	case resp.String, resp.BulkString:
		cmd = req.Str
	default:
		return Err("invalid command")
	}

	cmd = strings.ToLower(cmd)
	if !c.isAuthenticated() {
		switch cmd {
		case "auth", "hello", "ping":
		default:
			return NoAuth()
		}
	}

	return c.HandleCmd(cmd, args)
}

func (c *Client) HandleCmd(cmd string, args []resp.Data) resp.Data {
	if c.inSubscribeMode() {
		switch cmd {
		case "ping":
			return resp.NewData(resp.Array, "pong", "")
		case "subscribe", "unsubscribe", "quit":
		default:
			return Err(fmt.Sprintf("Can't execute '%s': only (P|S)SUBSCRIBE / (P|S)UNSUBSCRIBE / PING / QUIT / RESET are allowed in this context", cmd))
		}
	}

	switch cmd {
	case "ping":
		return resp.NewData(resp.String, "PONG")

	case "quit":
		c.quit = true
		return resp.NewData(resp.String, "OK")

	case "echo":
		if len(args) != 1 {
			return WrongArgs("echo")
		}
		return args[0]

	case "get":
		return HandleCmdGet(args)

	case "set":
		return HandleCmdSet(args)

	case "acl":
		return HandleCmdACL(args)

	case "auth":
		return HandleCmdAuth(c, args)

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
			return Err("save failed")
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

	default:
		msg := fmt.Sprintf("unknown command `%s`", cmd)
		return Err(msg)
	}
}

func Err(msg string) resp.Data {
	return resp.NewData(resp.Error, "ERR "+msg)
}

func WrongPass() resp.Data {
	return resp.NewData(resp.Error,
		"WRONGPASS invalid username-password pair or user is disabled.")
}

func NoAuth() resp.Data {
	return resp.NewData(resp.Error,
		"NOAUTH Authentication required.")
}

func WrongArgs(cmd string) resp.Data {
	return Err("wrong number of arguments for '" + cmd + "' command")
}
