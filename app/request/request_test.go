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

func resetStore() { store.DB = store.RedisStore{Store: make(map[string]*store.StoreMember)} }

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

// ---- List command tests (RPUSH / LRANGE) ----

func TestRequest_RPUSH_NewList(t *testing.T) {
	resetStore()
	// RPUSH mylist a b c
	payload := "*5\r\n$5\r\nRPUSH\r\n$6\r\nmylist\r\n$1\r\na\r\n$1\r\nb\r\n$1\r\nc\r\n"
	conn := newRW(payload)
	_, err := ReadAndHandleRequest(conn)
	require.NoError(t, err)
	// Should return integer length 3
	require.Equal(t, ":3\r\n", conn.Output())
}

func TestRequest_RPUSH_Append(t *testing.T) {
	resetStore()
	// RPUSH mylist a b, then RPUSH mylist c
	payload := "*4\r\n$5\r\nRPUSH\r\n$6\r\nmylist\r\n$1\r\na\r\n$1\r\nb\r\n*3\r\n$5\r\nRPUSH\r\n$6\r\nmylist\r\n$1\r\nc\r\n"
	conn := newRW(payload)
	_, err := ReadAndHandleRequest(conn)
	require.NoError(t, err)
	// First returns :2 then :3
	require.Equal(t, ":2\r\n:3\r\n", conn.Output())
}

func TestRequest_LRANGE_Full(t *testing.T) {
	resetStore()
	// RPUSH then LRANGE mylist 0 2
	payload := "*5\r\n$5\r\nRPUSH\r\n$6\r\nmylist\r\n$1\r\na\r\n$1\r\nb\r\n$1\r\nc\r\n*4\r\n$6\r\nLRANGE\r\n$6\r\nmylist\r\n$1\r\n0\r\n$1\r\n2\r\n"
	conn := newRW(payload)
	_, err := ReadAndHandleRequest(conn)
	require.NoError(t, err)
	// Expect array of 3 bulk strings: *3 $1 a $1 b $1 c (our RESP builder might format differently; adjust to produced output)
	out := conn.Output()
	require.Contains(t, out, "*3")
	require.Contains(t, out, "$1\r\na")
	require.Contains(t, out, "$1\r\nb")
	require.Contains(t, out, "$1\r\nc")
}

func TestRequest_LRANGE_EmptyRange(t *testing.T) {
	resetStore()
	// RPUSH then LRANGE mylist 2 1 (start >= end) -> empty array
	payload := "*4\r\n$5\r\nRPUSH\r\n$6\r\nmylist\r\n$1\r\na\r\n$1\r\nb\r\n*4\r\n$6\r\nLRANGE\r\n$6\r\nmylist\r\n$1\r\n2\r\n$1\r\n1\r\n"
	conn := newRW(payload)
	_, err := ReadAndHandleRequest(conn)
	require.NoError(t, err)
	require.Contains(t, conn.Output(), "*0\r\n")
}

func TestRequest_RPUSH_WrongArgTypes(t *testing.T) {
	resetStore()
	// Provide an empty bulk string? Here we simulate invalid by sending an integer style (but current parser may treat differently). Use missing args to trigger error.
	payload := "*2\r\n$5\r\nRPUSH\r\n$6\r\nmylist\r\n"
	conn := newRW(payload)
	_, err := ReadAndHandleRequest(conn)
	require.NoError(t, err)
	require.Contains(t, conn.Output(), "-ERROR")
}

func TestRequest_LPOP_Twice(t *testing.T) {
	resetStore()
	payload := "*10\r\n$5\r\nRPUSH\r\n$9\r\nblueberry\r\n$5\r\nmango\r\n$10\r\nstrawberry\r\n$5\r\napple\r\n$6\r\norange\r\n$9\r\npineapple\r\n$9\r\nblueberry\r\n$6\r\nbanana\r\n$4\r\npear\r\n*3\r\n$4\r\nLPOP\r\n$9\r\nblueberry\r\n$1\r\n2\r\n"
	conn := newRW(payload)
	_, err := ReadAndHandleRequest(conn)
	require.NoError(t, err)
	fmt.Println(conn.Output())
}

func TestRequest_LRANGE_WrongArity(t *testing.T) {
	resetStore()
	// LRANGE mylist 0 1 2 (too many)
	payload := "*5\r\n$6\r\nLRANGE\r\n$6\r\nmylist\r\n$1\r\n0\r\n$1\r\n1\r\n$1\r\n2\r\n"
	conn := newRW(payload)
	_, err := ReadAndHandleRequest(conn)
	require.NoError(t, err)
	require.Contains(t, conn.Output(), "-ERROR")
}

func TestRequest_LRANGE_NegativeEnd(t *testing.T) {
	resetStore()
	// RPUSH list_key a b c d e ; LRANGE list_key 2 -1  => c d e
	payload := "*7\r\n$5\r\nRPUSH\r\n$8\r\nlist_key\r\n$1\r\na\r\n$1\r\nb\r\n$1\r\nc\r\n$1\r\nd\r\n$1\r\ne\r\n" +
		"*4\r\n$6\r\nLRANGE\r\n$8\r\nlist_key\r\n$1\r\n2\r\n$2\r\n-1\r\n"
	conn := newRW(payload)
	_, err := ReadAndHandleRequest(conn)
	require.NoError(t, err)
	out := conn.Output()
	// Expect 3 elements c d e
	require.Contains(t, out, "*3")
	require.Contains(t, out, "$1\r\nc")
	require.Contains(t, out, "$1\r\nd")
	require.Contains(t, out, "$1\r\ne")
}

func TestRequest_LRANGE_NegativeEnd2(t *testing.T) {
	resetStore()
	payload := "*6\r\n$5\r\nRPUSH\r\n$9\r\nraspberry\r\n$6\r\nbanana\r\n$4\r\npear\r\n$9\r\nblueberry\r\n$9\r\nraspberry\r\n*4\r\n$6\r\nLRANGE\r\n$9\r\nraspberry\r\n$1\r\n0\r\n$2\r\n-3\r\n"
	conn := newRW(payload)
	_, err := ReadAndHandleRequest(conn)
	require.NoError(t, err)
	out := conn.Output()
	// Expect 3 elements c d e
	fmt.Println(out)
}

func TestRequest_Remove_multiple_elements(t *testing.T) {
	resetStore()
	payload := "*7\r\n$5\r\nRPUSH\r\n$10\r\nstrawberry\r\n$9\r\npineapple\r\n$5\r\ngrape\r\n$6\r\nbanana\r\n$5\r\napple\r\n$6\r\norange\r\n*3\r\n$4\r\nLPOP\r\n$10\r\nstrawberry\r\n$1\r\n4\r\n*4\r\n$6\r\nLRANGE\r\n$10\r\nstrawberry\r\n$1\r\n0\r\n$2\r\n-1\r\n"
	conn := newRW(payload)
	_, err := ReadAndHandleRequest(conn)
	require.NoError(t, err)
	out := conn.Output()
	fmt.Println(out)
}
