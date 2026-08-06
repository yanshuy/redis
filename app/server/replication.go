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
	ping := resp.NewData(resp.Array, []string{"PING"})
	_, err = conn.Write(ping.ToResponse())
	if err != nil {
		return fmt.Errorf("failed to send message to master at %s: %w", address, err)
	}
	return nil
}
