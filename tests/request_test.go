package tests

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sync"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/client"
	handler "github.com/codecrafters-io/redis-starter-go/app/handlers"
	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/store"
	"github.com/stretchr/testify/require"
)

type rw struct {
	mu     sync.Mutex
	r      *bytes.Reader
	w      bytes.Buffer
	closed bool
}

func newRW(in string) *rw { return &rw{r: bytes.NewReader([]byte(in))} }
func (rw *rw) Read(p []byte) (int, error) {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	return rw.r.Read(p)
}
func (rw *rw) Write(p []byte) (int, error) {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	return rw.w.Write(p)
}
func (rw *rw) Close() error {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	rw.closed = true
	return nil
}
func (rw *rw) Output() string {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	return rw.w.String()
}

func setupTestStore() *server.Server {
	tmpDir := os.TempDir()
	dbFile := filepath.Join(tmpDir, "rdb.test")
	st, err := store.InitializeStore(tmpDir, dbFile)
	if err != nil {
		panic(err)
	}
	s := &server.Server{
		Config: server.Config{
			Dir:        tmpDir,
			Dbfilename: dbFile,
		},
		Store: st,
	}
	server.Global = s
	return s
}

func testReadAndHandleRequest(rs *server.Server, conn *rw) error {
	c := client.NewClient(conn)
	c.Server = rs

	writeDone := make(chan struct{})
	go func() {
		_ = c.WriteLoop()
		close(writeDone)
	}()

	cmdChan := make(chan *client.Client, 50)

	errChan := make(chan error, 1)
	go func() {
		errChan <- server.ReadRequests(c, cmdChan)
		close(cmdChan)
	}()

	for range cmdChan {
		handler.HandleCommand(c)
	}

	err := <-errChan

	c.Close()
	<-writeDone

	if errors.Is(err, io.EOF) {
		err = nil
	}

	return err
}

func TestRequest_PING(t *testing.T) {
	s := setupTestStore()
	conn := newRW("+PING\r\n")
	err := testReadAndHandleRequest(s, conn)
	require.NoError(t, err)
	require.Equal(t, "+PONG\r\n", conn.Output())
}

func TestRequest_ECHO_NoArgs(t *testing.T) {
	s := setupTestStore()
	conn := newRW("+ECHO\r\n")
	err := testReadAndHandleRequest(s, conn)
	require.NoError(t, err)
	require.Contains(t, conn.Output(), "-ERR")
}

func TestRequest_Unknown(t *testing.T) {
	s := setupTestStore()
	conn := newRW("+FOO\r\n")
	testReadAndHandleRequest(s, conn)
	require.Contains(t, conn.Output(), "-ERR")
}

func TestRequest_ECHO_WithArgs(t *testing.T) {
	s := setupTestStore()
	payload := "*2\r\n$4\r\nECHO\r\n$5\r\nHELLO\r\n"
	conn := newRW(payload)
	err := testReadAndHandleRequest(s, conn)
	require.NoError(t, err)
	require.Equal(t, "$5\r\nHELLO\r\n", conn.Output())
}

func TestRequest_PING_Array(t *testing.T) {
	s := setupTestStore()
	payload := "*1\r\n$4\r\nPING\r\n"
	conn := newRW(payload)
	err := testReadAndHandleRequest(s, conn)
	require.NoError(t, err)
	require.Equal(t, "+PONG\r\n", conn.Output())
}

func TestRequest_MultipleInOneBuffer(t *testing.T) {
	s := setupTestStore()
	payload := "+PING\r\n+PING\r\n"
	conn := newRW(payload)
	err := testReadAndHandleRequest(s, conn)
	require.NoError(t, err)
	require.Equal(t, "+PONG\r\n+PONG\r\n", conn.Output())
}

func TestRequest_GET_MissingKey(t *testing.T) {
	s := setupTestStore()
	payload := "*2\r\n$3\r\nGET\r\n$3\r\nFOO\r\n"
	conn := newRW(payload)
	err := testReadAndHandleRequest(s, conn)
	require.NoError(t, err)
	require.Equal(t, "$-1\r\n", conn.Output())
}

func TestRequest_SET_Then_GET(t *testing.T) {
	s := setupTestStore()
	payload := "*3\r\n$3\r\nSET\r\n$3\r\nFOO\r\n$3\r\nBAR\r\n*2\r\n$3\r\nGET\r\n$3\r\nFOO\r\n"
	conn := newRW(payload)
	err := testReadAndHandleRequest(s, conn)
	require.NoError(t, err)
	require.Equal(t, "+OK\r\n$3\r\nBAR\r\n", conn.Output())
}

func TestRequest_RPUSH_NewList(t *testing.T) {
	s := setupTestStore()
	payload := "*5\r\n$5\r\nRPUSH\r\n$6\r\nmylist\r\n$1\r\na\r\n$1\r\nb\r\n$1\r\nc\r\n"
	conn := newRW(payload)
	err := testReadAndHandleRequest(s, conn)
	require.NoError(t, err)
	require.Equal(t, ":3\r\n", conn.Output())
}

func TestRequest_LRANGE_Full(t *testing.T) {
	s := setupTestStore()
	payload := "*5\r\n$5\r\nRPUSH\r\n$6\r\nmylist\r\n$1\r\na\r\n$1\r\nb\r\n$1\r\nc\r\n*4\r\n$6\r\nLRANGE\r\n$6\r\nmylist\r\n$1\r\n0\r\n$1\r\n2\r\n"
	conn := newRW(payload)
	err := testReadAndHandleRequest(s, conn)
	require.NoError(t, err)
	out := conn.Output()
	require.Contains(t, out, "*3")
	require.Contains(t, out, "$1\r\na")
	require.Contains(t, out, "$1\r\nb")
	require.Contains(t, out, "$1\r\nc")
}

func Test_func(t *testing.T) {
	s := setupTestStore()
	cwd, _ := os.Getwd()
	dir := s.Config.Dir
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(cwd, dir)
	}
	j := path.Join("/root", s.Config.Dir, s.Config.Dbfilename)
	fmt.Println(j)
}
