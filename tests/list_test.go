package tests

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestList_RPUSH_LRANGE(t *testing.T) {
	s := createTestStore()
	payload := "*5\r\n$5\r\nRPUSH\r\n$6\r\nmylist\r\n$1\r\na\r\n$1\r\nb\r\n$1\r\nc\r\n" +
		"*4\r\n$6\r\nLRANGE\r\n$6\r\nmylist\r\n$1\r\n0\r\n$1\r\n2\r\n"
	out, err := runTestPayload(s, payload)
	require.NoError(t, err)
	require.Contains(t, out, ":3\r\n")
	require.Contains(t, out, "$1\r\na")
	require.Contains(t, out, "$1\r\nb")
	require.Contains(t, out, "$1\r\nc")
}

func TestList_LPUSH_LPOP(t *testing.T) {
	s := createTestStore()
	payload := "*4\r\n$5\r\nLPUSH\r\n$6\r\nmylist\r\n$1\r\na\r\n$1\r\nb\r\n" +
		"*3\r\n$4\r\nLPOP\r\n$6\r\nmylist\r\n$1\r\n1\r\n"
	out, err := runTestPayload(s, payload)
	require.NoError(t, err)
	require.Contains(t, out, ":2\r\n")
	require.Contains(t, out, "$1\r\nb\r\n")
}

func TestList_LLEN(t *testing.T) {
	s := createTestStore()
	payload := "*3\r\n$5\r\nRPUSH\r\n$6\r\nmylist\r\n$1\r\nx\r\n" +
		"*2\r\n$4\r\nLLEN\r\n$6\r\nmylist\r\n"
	out, err := runTestPayload(s, payload)
	require.NoError(t, err)
	require.Contains(t, out, ":1\r\n")
}

func TestList_LRANGE_EmptyRange(t *testing.T) {
	s := createTestStore()
	payload := "*4\r\n$5\r\nRPUSH\r\n$6\r\nmylist\r\n$1\r\na\r\n$1\r\nb\r\n" +
		"*4\r\n$6\r\nLRANGE\r\n$6\r\nmylist\r\n$1\r\n2\r\n$1\r\n1\r\n"
	out, err := runTestPayload(s, payload)
	require.NoError(t, err)
	require.Contains(t, out, "*0\r\n")
}

func TestList_LRANGE_NegativeIndex(t *testing.T) {
	s := createTestStore()
	payload := "*5\r\n$5\r\nRPUSH\r\n$6\r\nmylist\r\n$1\r\na\r\n$1\r\nb\r\n$1\r\nc\r\n" +
		"*4\r\n$6\r\nLRANGE\r\n$6\r\nmylist\r\n$1\r\n0\r\n$2\r\n-1\r\n"
	out, err := runTestPayload(s, payload)
	require.NoError(t, err)
	require.Contains(t, out, "*3\r\n")
}
