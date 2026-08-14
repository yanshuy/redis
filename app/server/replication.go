package server

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"

	resp "github.com/codecrafters-io/redis-starter-go/app/Resp"
	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

func (s *Server) NewReplica(c *client.Client) error {
	c.MakeSlave()

	str := []string{"FULLRESYNC", Global.ReplicationId, strconv.Itoa(Global.ReplicationOffset)}
	res := resp.NewData(resp.String, strings.Join(str, " "))

	_, err := c.Conn.Write(res.ToResponse())
	if err != nil {
		return err
	}
	err = store.SendRDB(c)
	if err != nil {
		return err
	}

	Global.Replicas[c] = 0
	return nil
}

func (s *Server) Propagate(cmd client.Command) {
	if s.Role == client.MASTER {
		cmd := cmd.ToRESP()
		raw := cmd.ToResponse()
		s.ReplicationOffset += len(raw)
		go func() {
			for slave := range s.Replicas {
				slave.Conn.Write(raw)
			}
		}()
	}
}

func (s *Server) CountSyncedReplicas(targetOffset int) int {
	count := 0
	for _, off := range s.Replicas {
		if off >= targetOffset {
			count++
		}
	}
	return count
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
	err = Exchange(c, ping, "PONG")
	if err != nil {
		return nil, err
	}

	listening_port := resp.NewData(resp.Array, []string{"REPLCONF", "listening-port", s.Config.port})
	err = Exchange(c, listening_port, "OK")
	if err != nil {
		return nil, err
	}

	capa := resp.NewData(resp.Array, []string{"REPLCONF", "capa", "psync2"})
	err = Exchange(c, capa, "OK")
	if err != nil {
		return nil, err
	}

	psync := resp.NewData(resp.Array, []string{"PSYNC", "?", "-1"})
	_, err = conn.Write(psync.ToResponse())
	if err != nil {
		return nil, fmt.Errorf("failed to send message to peer")
	}

	resync, _, err := c.Reader.Read_RESP(conn)
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

	err = c.Reader.Read_RDB(conn, io.Discard)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func Exchange(c *client.Client, out resp.Data, expected string) error {
	_, err := c.Conn.Write(out.ToResponse())
	if err != nil {
		return fmt.Errorf("failed to send message to peer")
	}
	resp, _, err := c.Reader.Read_RESP(c.Conn)
	if err != nil {
		return fmt.Errorf("failed to read message from peer")
	}
	if !strings.EqualFold(resp.Str, expected) {
		return fmt.Errorf("unexpected response from peer")
	}
	return nil
}
