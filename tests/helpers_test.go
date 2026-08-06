package tests

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/codecrafters-io/redis-starter-go/app/client"
	handler "github.com/codecrafters-io/redis-starter-go/app/handlers"
	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

type mockRW struct {
	mu     sync.Mutex
	r      *bytes.Reader
	w      bytes.Buffer
	closed bool
}

func newMockRW(in string) *mockRW {
	return &mockRW{r: bytes.NewReader([]byte(in))}
}

func (rw *mockRW) Read(p []byte) (int, error) {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	return rw.r.Read(p)
}

func (rw *mockRW) Write(p []byte) (int, error) {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	return rw.w.Write(p)
}

func (rw *mockRW) Close() error {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	rw.closed = true
	return nil
}

func (rw *mockRW) Output() string {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	return rw.w.String()
}

func createTestStore() *server.Server {
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

func runTestPayload(s *server.Server, payload string) (string, error) {
	conn := newMockRW(payload)
	c := client.NewClient(conn)
	c.Server = s

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

	return conn.Output(), err
}

var _ io.ReadWriteCloser = (*mockRW)(nil)
