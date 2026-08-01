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
	conn          io.Writer
	subscriptions map[string]chan string
	done          chan struct{}
}

func (c *Client) inSubscribeMode() bool {
	return len(c.subscriptions) > 0
}

func NewClient(conn io.Writer) *Client {
	return &Client{
		conn: conn,
		done: make(chan struct{}),
	}
}

func ReadAndHandleRequest(conn io.ReadWriter) (n int, err error) {
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

func (c *Client) HandleRequest(w io.Writer, rs []resp.Data) (err error) {
	for _, r := range rs {
		var res resp.Data
		switch r.Type {
		case resp.Array:
			if len(r.Arr) == 0 {
				res = Err("invalid command")
				break
			}
			f := r.Arr[0]
			if f.Str == "" {
				res = Err("invalid command")
				break
			}
			res = c.HandleCmd(f.Str, r.Arr[1:])
		case resp.String, resp.BulkString:
			res = c.HandleCmd(r.Str, nil)
		default:
			res = Err("invalid command")
		}

		resBytes := res.ToResponse()
		_, err := w.Write(resBytes)
		if err != nil {
			return err
		}
	}

	return err
}

func (c *Client) HandleCmd(cmd string, args []resp.Data) resp.Data {
	cmd = strings.ToLower(cmd)

	if c.inSubscribeMode() {
		switch cmd {
		case "ping":
			return resp.NewData(resp.Array, "pong", "")
		case "subscribe":
			return HandleSubscribe(c, args)
		case "unsubscribe":
			return HandleUnsubscribe(c, args)
		case "quit":
		default:
			return Err(fmt.Sprintf("Can't execute '%s': only (P|S)SUBSCRIBE / (P|S)UNSUBSCRIBE / PING / QUIT / RESET are allowed in this context", cmd))
		}
	}

	switch cmd {
	case "ping":
		return resp.NewData(resp.String, "PONG")

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
		return HandleCmdAuth(args)

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

	case "publish":
		return HandlePublish(args)

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

func ErrWrongArgs(cmd string) resp.Data {
	return WrongArgs(cmd)
}
