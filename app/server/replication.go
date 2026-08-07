package server

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

func (s *Server) NewReplica(c *client.Client) error {
	c.Role = client.SLAVE
	c.CloseMessageChan()

	str := []string{"FULLRESYNC", Global.ReplicationId, strconv.Itoa(Global.ReplicationOffset)}
	res := resp.NewData(resp.String, strings.Join(str, " "))

	_, err := c.Conn.Write(res.ToResponse())
	if err != nil {
		return fmt.Errorf("failed to send message to peer")
	}

	store.SendRDB(c)
	//TODO: handle error

	Global.replicas = append(Global.replicas, c)
	return nil
}

func (s *Server) Propagate(cmd client.Command) {
	for _, client := range s.replicas {
		client.QueueMessage(cmd.ToRESP())
	}
}

func (s *Server) HandshakeMaster() error {
	parts := strings.Fields(s.replicaof)
	host := parts[0]
	port := parts[1]

	address := net.JoinHostPort(host, port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("failed to connect to master at %s: %w", address, err)
	}

	r := NewReader()

	ping := resp.NewData(resp.Array, []string{"PING"})
	err = r.exchange(conn, ping, "PONG")
	if err != nil {
		return err
	}

	listening_port := resp.NewData(resp.Array, []string{"REPLCONF", "listening-port", s.Config.port})
	err = r.exchange(conn, listening_port, "OK")
	if err != nil {
		return err
	}

	capa := resp.NewData(resp.Array, []string{"REPLCONF", "capa", "psync2"})
	err = r.exchange(conn, capa, "OK")
	if err != nil {
		return err
	}

	psync := resp.NewData(resp.Array, []string{"PSYNC", "?", "-1"})
	_, err = conn.Write(psync.ToResponse())
	if err != nil {
		return fmt.Errorf("failed to send message to peer")
	}
	resync, err := r.ReadRESP(conn)
	if err != nil {
		return fmt.Errorf("failed to read message from peer")
	}
	fields := strings.Fields(resync.Str)
	repl_id := fields[1]
	repl_off := fields[2]
	fmt.Print(repl_id, repl_off)

	return nil
}
