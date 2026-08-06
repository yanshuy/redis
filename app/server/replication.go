package server

import (
	"fmt"
	"net"
	"strings"

	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
)

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

func (r *Reader) exchange(conn net.Conn, out resp.Data, expected string) error {
	_, err := conn.Write(out.ToResponse())
	if err != nil {
		return fmt.Errorf("failed to send message to peer")
	}
	resp, err := r.ReadRESP(conn)
	if err != nil {
		return fmt.Errorf("failed to read message from peer")
	}
	if !strings.EqualFold(resp.Str, expected) {
		return fmt.Errorf("unexpected response from peer")
	}
	return nil
}
