package request

import (
	"fmt"
	"io"
	"log"
	"strings"

	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

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

type Client struct {
	conn          io.Writer
	subscriptions map[string]struct{}
	subscribeMode bool
}

func NewClient(conn io.Writer) *Client {
	return &Client{
		conn:          conn,
		subscriptions: make(map[string]struct{}),
		subscribeMode: false,
	}
}

func (c *Client) HandleRequest(w io.Writer, rs []resp.DataType) (err error) {
	for _, r := range rs {
		var res resp.DataType
		switch r.Type {
		case resp.Array:
			f := r.Arr[0]
			if f.Str == "" {
				res = resp.NewData(resp.Error, "invalid command")
				break
			}
			if strings.ToLower(f.Str) == "ping" {
				res = resp.NewData(resp.String, "PONG")
				break
			}
			res = c.HandleCmd(f.Str, r.Arr[1:])
		default:
			if strings.ToLower(r.Str) == "ping" {
				res = resp.NewData(resp.String, "PONG")
				break
			}
			res = resp.NewData(resp.Error, "invalid command")
		}

		resBytes := res.ToResponse()
		_, err := w.Write(resBytes)
		if err != nil {
			return err
		}
	}

	return err
}

func (c *Client) HandleCmd(cmd string, args []resp.DataType) resp.DataType {
	cmd = strings.ToLower(cmd)

	if c.subscribeMode {
		switch cmd {
		case "ping":
			return resp.NewData(resp.Array, "PONG", "")
		case "subscribe":
			return HandleSubscribe(c, args)
		case "unsubscribe":
		case "quit":
		default:
			return resp.NewData(resp.Error, fmt.Sprintf("Can't execute '%s': only (P|S)SUBSCRIBE / (P|S)UNSUBSCRIBE / PING / QUIT / RESET are allowed in this context", cmd))
		}
	}

	switch cmd {
	case "echo":
		if len(args) != 1 {
			return resp.NewData(resp.Error, "wrong number of arguments for 'echo' command")
		}
		return args[0]

	case "get":
		return HandleCmdGet(args)

	case "set":
		return HandleCmdSet(args)

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
			return resp.NewData(resp.Error, "save failed")
		}
		return resp.NewData(resp.String, "OK")

	case "subscribe":
		c.subscribeMode = true
		return HandleSubscribe(c, args)

	default:
		msg := fmt.Sprintf("unknown command `%s`", cmd)
		return resp.NewData(resp.Error, msg)
	}
}
