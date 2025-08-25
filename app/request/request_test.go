package request

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/codecrafters-io/redis-starter-go/app/store"
	"github.com/stretchr/testify/require"
)

type rw struct {
	r *bytes.Reader
	w bytes.Buffer
}

func newRW(in string) *rw                  { return &rw{r: bytes.NewReader([]byte(in))} }
func (rw *rw) Read(p []byte) (int, error)  { return rw.r.Read(p) }
func (rw *rw) Write(p []byte) (int, error) { return rw.w.Write(p) }
func (rw *rw) Output() string              { return rw.w.String() }

func TestRequest_PING(t *testing.T) {
	conn := newRW("+PING\r\n")
	_, err := ReadAndHandleRequest(conn)
	require.NoError(t, err)
	require.Equal(t, "+PONG\r\n", conn.Output())
}

func TestRequest_ECHO_NoArgs(t *testing.T) {
	conn := newRW("+ECHO\r\n")
	_, err := ReadAndHandleRequest(conn)
	require.NoError(t, err)
	require.Contains(t, conn.Output(), "-ERROR")
}

func TestRequest_Unknown(t *testing.T) {
	conn := newRW("+FOO\r\n")
	ReadAndHandleRequest(conn)
	require.Contains(t, conn.Output(), "-ERROR")
}

func TestRequest_ECHO_WithArgs(t *testing.T) {
	// Array form: *2 CRLF $4 CRLF ECHO CRLF $5 CRLF HELLO CRLF
	payload := "*2\r\n$4\r\nECHO\r\n$5\r\nHELLO\r\n"
	conn := newRW(payload)
	_, err := ReadAndHandleRequest(conn)
	fmt.Println(err)
	require.NoError(t, err)
	// Current implementation writes plain HELLO (no RESP wrapping)
	require.Equal(t, "+HELLO\r\n", conn.Output())
}

func TestRequest_PING_Array(t *testing.T) {
	payload := "*1\r\n$4\r\nPING\r\n"
	conn := newRW(payload)
	_, err := ReadAndHandleRequest(conn)
	require.NoError(t, err)
	require.Equal(t, "+PONG\r\n", conn.Output())
}

func TestRequest_MultipleInOneBuffer(t *testing.T) {
	// Two PING commands concatenated
	payload := "+PING\r\n+PING\r\n"
	conn := newRW(payload)
	_, err := ReadAndHandleRequest(conn)
	require.NoError(t, err)
	require.Equal(t, "+PONG\r\n+PONG\r\n", conn.Output())
}

func resetStore() { store.R = store.RedisStore{Store: make(map[string]store.StoreMember)} }

func TestRequest_GET_MissingKey(t *testing.T) {
	resetStore()
	payload := "*2\r\n$3\r\nGET\r\n$3\r\nFOO\r\n"
	conn := newRW(payload)
	_, err := ReadAndHandleRequest(conn)
	require.NoError(t, err)
	// Missing key should return null bulk string
	require.Equal(t, "$-1\r\n", conn.Output())
}

func TestRequest_SET_Then_GET(t *testing.T) {
	resetStore()
	// SET FOO BAR   then   GET FOO
	payload := "*3\r\n$3\r\nSET\r\n$3\r\nFOO\r\n$3\r\nBAR\r\n*2\r\n$3\r\nGET\r\n$3\r\nFOO\r\n"
	conn := newRW(payload)
	_, err := ReadAndHandleRequest(conn)
	require.NoError(t, err)
	require.Equal(t, "+OK\r\n$3\r\nBAR\r\n", conn.Output())
}

func TestRequest_SET_WithPX(t *testing.T) {
	resetStore()
	// SET FOO BAR PX 100
	payload := "*5\r\n$3\r\nSET\r\n$3\r\nFOO\r\n$3\r\nBAR\r\n$2\r\nPX\r\n$3\r\n100\r\n"
	conn := newRW(payload)
	_, err := ReadAndHandleRequest(conn)
	require.NoError(t, err)
	require.Equal(t, "+OK\r\n", conn.Output())
}

func TestRequest_SET_WrongArgs(t *testing.T) {
	resetStore()
	// Only SET FOO (missing value)
	payload := "*2\r\n$3\r\nSET\r\n$3\r\nFOO\r\n"
	conn := newRW(payload)
	_, err := ReadAndHandleRequest(conn)
	require.NoError(t, err)
	require.Contains(t, conn.Output(), "-ERROR")
}

func TestRequest_GET_WrongArgs(t *testing.T) {
	resetStore()
	// GET with two keys
	payload := "*3\r\n$3\r\nGET\r\n$3\r\nFOO\r\n$3\r\nBAR\r\n"
	conn := newRW(payload)
	_, err := ReadAndHandleRequest(conn)
	require.NoError(t, err)
	require.Contains(t, conn.Output(), "-ERROR")
}

func TestRequest_Unknown_Array(t *testing.T) {
	resetStore()
	payload := "*1\r\n$7\r\nUNKNOWN\r\n"
	conn := newRW(payload)
	_, err := ReadAndHandleRequest(conn)
	require.NoError(t, err)
	require.Contains(t, conn.Output(), "-ERROR")
}
