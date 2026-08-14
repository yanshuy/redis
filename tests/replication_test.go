package tests

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/client"
	handler "github.com/codecrafters-io/redis-starter-go/app/handlers"
	"github.com/stretchr/testify/require"
)

func TestReplication_WAIT_NoReplicas_ZeroOffset(t *testing.T) {
	s := createTestStore()
	out, err := runTestPayload(s, "*3\r\n$4\r\nWAIT\r\n$1\r\n1\r\n$3\r\n100\r\n")
	require.NoError(t, err)
	require.Equal(t, ":0\r\n", out)
}

func TestReplication_WAIT_ZeroReplicasRequested(t *testing.T) {
	s := createTestStore()
	r1Pipe, _ := net.Pipe()
	r2Pipe, _ := net.Pipe()
	defer r1Pipe.Close()
	defer r2Pipe.Close()

	r1 := client.NewClient(r1Pipe, client.SLAVE)
	r2 := client.NewClient(r2Pipe, client.SLAVE)

	s.Replicas[r1] = 0
	s.Replicas[r2] = 0

	out, err := runTestPayload(s, "*3\r\n$4\r\nWAIT\r\n$1\r\n0\r\n$3\r\n500\r\n")
	require.NoError(t, err)
	require.Equal(t, ":2\r\n", out)
}

func TestReplication_WAIT_AlreadySynced(t *testing.T) {
	s := createTestStore()
	r1Pipe, _ := net.Pipe()
	r2Pipe, _ := net.Pipe()
	defer r1Pipe.Close()
	defer r2Pipe.Close()

	r1 := client.NewClient(r1Pipe, client.SLAVE)
	r2 := client.NewClient(r2Pipe, client.SLAVE)

	s.ReplicationOffset = 50
	s.Replicas[r1] = 50
	s.Replicas[r2] = 50

	start := time.Now()
	out, err := runTestPayload(s, "*3\r\n$4\r\nWAIT\r\n$1\r\n2\r\n$4\r\n2000\r\n")
	require.NoError(t, err)
	require.True(t, time.Since(start) < 200*time.Millisecond, "WAIT should return immediately when all replicas are synced")
	require.Equal(t, ":2\r\n", out)
}

func TestReplication_WAIT_TimeoutWithPartialACKs(t *testing.T) {
	s := createTestStore()
	r1Pipe, _ := net.Pipe()
	r2Pipe, _ := net.Pipe()
	r3Pipe, _ := net.Pipe()
	defer r1Pipe.Close()
	defer r2Pipe.Close()
	defer r3Pipe.Close()

	r1 := client.NewClient(r1Pipe, client.SLAVE)
	r2 := client.NewClient(r2Pipe, client.SLAVE)
	r3 := client.NewClient(r3Pipe, client.SLAVE)

	s.ReplicationOffset = 100
	s.Replicas[r1] = 100
	s.Replicas[r2] = 100
	s.Replicas[r3] = 40

	out, err := runTestPayload(s, "*3\r\n$4\r\nWAIT\r\n$1\r\n3\r\n$2\r\n50\r\n")
	require.NoError(t, err)
	require.Equal(t, ":2\r\n", out)
}

func TestReplication_WAIT_UnblocksEarlyOnACK(t *testing.T) {
	s := createTestStore()
	s.Role = client.MASTER

	r1Pipe, _ := net.Pipe()
	r2Pipe, _ := net.Pipe()
	defer r1Pipe.Close()
	defer r2Pipe.Close()

	r1 := client.NewClient(r1Pipe, client.SLAVE)
	r2 := client.NewClient(r2Pipe, client.SLAVE)

	s.Replicas[r1] = 0
	s.Replicas[r2] = 0
	s.ReplicationOffset = 50

	clientPipe, testPipe := net.Pipe()
	c := client.NewClient(testPipe, client.CLIENT)

	go func() {
		_ = c.WriteLoop()
	}()

	reqChan := make(chan client.Request, 10)

	stopChan := make(chan struct{})
	go func() {
		for {
			select {
			case req := <-reqChan:
				handler.HandleRequest(req)
			case t := <-client.BlopChan:
				t()
			case <-stopChan:
				return
			}
		}
	}()

	waitCmd := client.Command{
		Name: "wait",
		Args: []string{"2", "3000"},
	}
	reqChan <- client.Request{Client: c, Cmd: waitCmd}

	time.Sleep(50 * time.Millisecond)

	ack1Cmd := client.Command{
		Name: "replconf",
		Args: []string{"ack", "50"},
	}
	reqChan <- client.Request{Client: r1, Cmd: ack1Cmd}

	time.Sleep(50 * time.Millisecond)

	ack2Cmd := client.Command{
		Name: "replconf",
		Args: []string{"ack", "50"},
	}
	reqChan <- client.Request{Client: r2, Cmd: ack2Cmd}

	buf := make([]byte, 100)
	clientPipe.SetReadDeadline(time.Now().Add(1 * time.Second))
	n, err := clientPipe.Read(buf)
	require.NoError(t, err)
	require.Equal(t, ":2\r\n", string(buf[:n]))

	close(stopChan)
	c.Close()
	clientPipe.Close()
	testPipe.Close()
}

func TestReplication_REPLCONF_GETACK_Slave(t *testing.T) {
	s := createTestStore()
	s.Role = client.SLAVE
	s.ReplicationOffset = 31

	payload := "*3\r\n$8\r\nREPLCONF\r\n$6\r\nGETACK\r\n$1\r\n*\r\n"
	out, err := runTestPayload(s, payload)
	require.NoError(t, err)
	require.Equal(t, "*3\r\n$8\r\nREPLCONF\r\n$3\r\nACK\r\n$2\r\n31\r\n", out)
}

func TestReplication_INFO_Replication(t *testing.T) {
	s := createTestStore()
	s.Role = client.MASTER
	s.ReplicationId = "8371b7b16d329739a4c8d5123456789012345678"
	s.ReplicationOffset = 0

	payload := "*2\r\n$4\r\nINFO\r\n$11\r\nreplication\r\n"
	out, err := runTestPayload(s, payload)
	require.NoError(t, err)
	require.Contains(t, out, "role:master")
	require.Contains(t, out, fmt.Sprintf("master_replid:%s", s.ReplicationId))
	require.Contains(t, out, "master_repl_offset:0")
}
