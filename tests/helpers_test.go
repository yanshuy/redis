package tests

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/client"
	"github.com/codecrafters-io/redis-starter-go/app/request"
	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/server/store"
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
	return &server.Server{
		Config: server.Config{
			Dir:        tmpDir,
			Dbfilename: dbFile,
		},
		Store: st,
	}
}

func runTestPayload(s *server.Server, payload string) (string, error) {
	conn := newMockRW(payload)
	c := client.NewClient(conn)

	reqChan := make(chan request.Request, 50)
	done := make(chan struct{})

	go func() {
		for req := range reqChan {
			request.HandleRequest(s, req)
		}
		close(done)
	}()

	writerDone := make(chan struct{})
	go func() {
		for respData := range c.RespChan {
			resBytes := respData.ToResponse()
			conn.Write(resBytes)
		}
		close(writerDone)
	}()

	err := request.ReadRequests(c, reqChan)
	close(reqChan)
	if errors.Is(err, io.EOF) {
		err = nil
	}

	select {
	case <-done:
	case <-time.After(1 * time.Second):
	}

	c.Close()
	<-writerDone

	return conn.Output(), err
}

var _ io.ReadWriteCloser = (*mockRW)(nil)
