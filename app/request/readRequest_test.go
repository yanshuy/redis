package request

import (
	"bytes"
	"fmt"
	"testing"

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

func TestReadAndHandleRequest_PING(t *testing.T) {
	conn := newRW("+PING\r\n")
	_, err := ReadAndHandleRequest(conn)
	require.NoError(t, err)
	require.Equal(t, "+PONG\r\n", conn.Output())
}

func TestReadAndHandleRequest_ECHO_NoArgs(t *testing.T) {
	conn := newRW("+ECHO\r\n")
	_, err := ReadAndHandleRequest(conn)
	require.NoError(t, err)
	require.Contains(t, conn.Output(), "-ERROR")
}

func TestReadAndHandleRequest_Unknown(t *testing.T) {
	conn := newRW("+FOO\r\n")
	ReadAndHandleRequest(conn)
	require.Contains(t, conn.Output(), "-ERROR")
}

func TestReadAndHandleRequest_ECHO_WithArgs(t *testing.T) {
	// Array form: *2 CRLF $4 CRLF ECHO CRLF $5 CRLF HELLO CRLF
	payload := "*2\r\n$4\r\nECHO\r\n$5\r\nHELLO\r\n"
	conn := newRW(payload)
	_, err := ReadAndHandleRequest(conn)
	fmt.Println(err)
	require.NoError(t, err)
	// Current implementation writes plain HELLO (no RESP wrapping)
	require.Equal(t, "+HELLO\r\n", conn.Output())
}

func TestReadAndHandleRequest_PING_Array(t *testing.T) {
	payload := "*1\r\n$4\r\nPING\r\n"
	conn := newRW(payload)
	_, err := ReadAndHandleRequest(conn)
	require.NoError(t, err)
	require.Equal(t, "+PONG\r\n", conn.Output())
}

func TestReadAndHandleRequest_MultipleInOneBuffer(t *testing.T) {
	// Two PING commands concatenated
	payload := "+PING\r\n+PING\r\n"
	conn := newRW(payload)
	_, err := ReadAndHandleRequest(conn)
	require.NoError(t, err)
	require.Equal(t, "+PONG\r\n+PONG\r\n", conn.Output())
}
