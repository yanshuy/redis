package server

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"

	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

func (s *Server) NewReplica(c *client.Client) error {
	c.MakeSlave()
	Global.replicas = append(Global.replicas, c)

	str := []string{"FULLRESYNC", Global.ReplicationId, strconv.Itoa(Global.ReplicationOffset)}
	res := resp.NewData(resp.String, strings.Join(str, " "))

	go func() {
		c.Block()

		c.Conn.Write(res.ToResponse())
		store.SendRDB(c)
		//TODO: handle error

		c.UnBlock()
	}()

	return nil
}

func (s *Server) Propagate(cmd client.Command) {
	if s.Role == client.MASTER {
		cmd := cmd.ToRESP()
		raw := cmd.ToResponse()
		s.ReplicationOffset += len(raw)
		go func() {
			for _, client := range s.replicas {
				client.Conn.Write(raw)
			}
		}()
	}
}

func (s *Server) RequestAcks() chan AckMessage {
	ch := make(chan AckMessage, s.ReplicaCount())

	getack := resp.NewData(resp.Array, []string{"REPLCONF", "GETACK", "*"})
	b := getack.ToResponse()

	for _, slave := range s.replicas {
		go func() {
			slave.Block()

			_, err := slave.Conn.Write(b)
			if err != nil {
				return
			}

		read:
			resp, _, err := slave.Reader.ReadRESP(slave.Conn)
			if err != nil {
				return
			}
			cmd, err := client.ValidateCommand(resp)
			if err != nil {
				return
			}
			if cmd.Name != "replconf" || len(cmd.Args) != 2 ||
				strings.ToLower(cmd.Args[0]) != "ack" {
				goto read
			}
			off, err := strconv.Atoi(cmd.Args[1])
			if err != nil || off < 0 {
				return
			}

			ch <- NewAck(slave, off)

			slave.UnBlock()
		}()
	}
	return ch
}

func (s *Server) HandshakeMaster() (*client.Client, error) {
	parts := strings.Fields(s.replicaof)
	host := parts[0]
	port := parts[1]

	address := net.JoinHostPort(host, port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to master at %s: %w", address, err)
	}
	c := client.NewClient(conn, client.MASTER)

	ping := resp.NewData(resp.Array, []string{"PING"})
	err = c.Reader.Exchange(conn, ping, "PONG")
	if err != nil {
		return nil, err
	}

	listening_port := resp.NewData(resp.Array, []string{"REPLCONF", "listening-port", s.Config.port})
	err = c.Reader.Exchange(conn, listening_port, "OK")
	if err != nil {
		return nil, err
	}

	capa := resp.NewData(resp.Array, []string{"REPLCONF", "capa", "psync2"})
	err = c.Reader.Exchange(conn, capa, "OK")
	if err != nil {
		return nil, err
	}

	psync := resp.NewData(resp.Array, []string{"PSYNC", "?", "-1"})
	_, err = conn.Write(psync.ToResponse())
	if err != nil {
		return nil, fmt.Errorf("failed to send message to peer")
	}

	resync, _, err := c.Reader.ReadRESP(conn)
	if err != nil {
		return nil, fmt.Errorf("failed to read message from peer")
	}
	fields := strings.Fields(resync.Str)
	repl_id := fields[1]
	repl_off := fields[2]
	fmt.Print(repl_id, " offset: ", repl_off, "\n")

	file, err := s.Store.GetRDBFile(os.O_CREATE | os.O_WRONLY | os.O_TRUNC)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	err = c.Reader.SaveRDB(conn, io.Discard)
	if err != nil {
		return nil, err
	}

	return c, nil
}
