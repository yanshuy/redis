package tests

import (
	"bytes"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/client"
	handler "github.com/codecrafters-io/redis-starter-go/app/handlers"
	"github.com/codecrafters-io/redis-starter-go/app/server"
	"github.com/codecrafters-io/redis-starter-go/app/store"
)

type mockAddr struct{}

func (m mockAddr) Network() string { return "tcp" }
func (m mockAddr) String() string  { return "127.0.0.1:6379" }

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

func (rw *mockRW) LocalAddr() net.Addr                { return mockAddr{} }
func (rw *mockRW) RemoteAddr() net.Addr               { return mockAddr{} }
func (rw *mockRW) SetDeadline(t time.Time) error      { return nil }
func (rw *mockRW) SetReadDeadline(t time.Time) error  { return nil }
func (rw *mockRW) SetWriteDeadline(t time.Time) error { return nil }

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
		Role: client.MASTER,
		Config: server.Config{
			Dir:        tmpDir,
			Dbfilename: dbFile,
		},
		Store:    st,
		Replicas: make(map[*client.Client]int),
	}
	server.Global = s
	return s
}

func runTestPayload(s *server.Server, payload string) (string, error) {
	conn := newMockRW(payload)
	c := client.NewClient(conn, client.CLIENT)

	writeDone := make(chan struct{})
	go func() {
		_ = c.WriteLoop()
		close(writeDone)
	}()

	reqChan := make(chan client.Request, 50)
	readErrChan := make(chan error, 1)

	go func() {
		for {
			req, err := c.ReadRequest()
			if err != nil {
				readErrChan <- err
				close(reqChan)
				return
			}
			reqChan <- req
		}
	}()

	var activeReqChan <-chan client.Request = reqChan
	done := false

	for !done {
		select {
		case req, ok := <-activeReqChan:
			if !ok {
				activeReqChan = nil
			} else {
				handler.HandleRequest(req)
			}
		case t := <-client.BlopChan:
			t()
		}

		if activeReqChan == nil && !c.Blocked && len(client.BlopChan) == 0 {
			done = true
		}
	}

	time.Sleep(5 * time.Millisecond)

	c.Close()
	<-writeDone

	err := <-readErrChan
	if errors.Is(err, io.EOF) {
		err = nil
	}

	return conn.Output(), err
}

var _ net.Conn = (*mockRW)(nil)
