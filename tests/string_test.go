package tests

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestString_PING(t *testing.T) {
	s := createTestStore()
	out, err := runTestPayload(s, "*1\r\n$4\r\nPING\r\n")
	require.NoError(t, err)
	require.Equal(t, "+PONG\r\n", out)
}

func TestString_ECHO(t *testing.T) {
	s := createTestStore()
	out, err := runTestPayload(s, "*2\r\n$4\r\nECHO\r\n$5\r\nHELLO\r\n")
	require.NoError(t, err)
	require.Equal(t, "$5\r\nHELLO\r\n", out)
}

func TestString_SET_GET(t *testing.T) {
	s := createTestStore()
	payload := "*3\r\n$3\r\nSET\r\n$3\r\nFOO\r\n$3\r\nBAR\r\n*2\r\n$3\r\nGET\r\n$3\r\nFOO\r\n"
	out, err := runTestPayload(s, payload)
	require.NoError(t, err)
	require.Contains(t, out, "+OK\r\n")
	require.Contains(t, out, "$3\r\nBAR\r\n")
}

func TestString_GET_MissingKey(t *testing.T) {
	s := createTestStore()
	payload := "*2\r\n$3\r\nGET\r\n$7\r\nMISSING\r\n"
	out, err := runTestPayload(s, payload)
	require.NoError(t, err)
	require.Equal(t, "$-1\r\n", out)
}

func TestString_INCR(t *testing.T) {
	s := createTestStore()
	payload := "*2\r\n$4\r\nINCR\r\n$7\r\ncounter\r\n*2\r\n$4\r\nINCR\r\n$7\r\ncounter\r\n"
	out, err := runTestPayload(s, payload)
	require.NoError(t, err)
	require.Contains(t, out, ":1\r\n")
	require.Contains(t, out, ":2\r\n")
}

func TestString_TYPE(t *testing.T) {
	s := createTestStore()
	payload := "*3\r\n$3\r\nSET\r\n$1\r\nk\r\n$1\r\nv\r\n*2\r\n$4\r\nTYPE\r\n$1\r\nk\r\n"
	out, err := runTestPayload(s, payload)
	require.NoError(t, err)
	require.Contains(t, out, "+OK\r\n")
	require.Contains(t, out, "+string\r\n")
}

func TestString_KEYS(t *testing.T) {
	s := createTestStore()
	payload := "*3\r\n$3\r\nSET\r\n$4\r\nfoo1\r\n$1\r\nv\r\n*2\r\n$4\r\nKEYS\r\n$1\r\n*\r\n"
	out, err := runTestPayload(s, payload)
	require.NoError(t, err)
	require.Contains(t, out, "foo1")
}
