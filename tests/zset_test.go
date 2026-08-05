package tests

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestZSet_ZADD_ZRANGE(t *testing.T) {
	s := createTestStore()
	payload := "*6\r\n$4\r\nZADD\r\n$6\r\nmyzset\r\n$1\r\n1\r\n$4\r\none1\r\n$1\r\n2\r\n$4\r\ntwo2\r\n" +
		"*4\r\n$6\r\nZRANGE\r\n$6\r\nmyzset\r\n$1\r\n0\r\n$2\r\n-1\r\n"
	out, err := runTestPayload(s, payload)
	require.NoError(t, err)
	require.Contains(t, out, ":2\r\n")
	require.Contains(t, out, "*2\r\n")
	require.Contains(t, out, "$4\r\none1")
	require.Contains(t, out, "$4\r\ntwo2")
}

func TestZSet_ZCARD_ZSCORE(t *testing.T) {
	s := createTestStore()
	payload := "*4\r\n$4\r\nZADD\r\n$6\r\nmyzset\r\n$2\r\n10\r\n$1\r\na\r\n" +
		"*2\r\n$5\r\nZCARD\r\n$6\r\nmyzset\r\n" +
		"*3\r\n$6\r\nZSCORE\r\n$6\r\nmyzset\r\n$1\r\na\r\n"
	out, err := runTestPayload(s, payload)
	require.NoError(t, err)
	require.Contains(t, out, ":1\r\n")
	require.Contains(t, out, "$2\r\n10\r\n")
}

func TestZSet_ZRANK_ZREM(t *testing.T) {
	s := createTestStore()
	payload := "*6\r\n$4\r\nZADD\r\n$6\r\nmyzset\r\n$1\r\n1\r\n$1\r\na\r\n$1\r\n2\r\n$1\r\nb\r\n" +
		"*3\r\n$5\r\nZRANK\r\n$6\r\nmyzset\r\n$1\r\nb\r\n" +
		"*3\r\n$4\r\nZREM\r\n$6\r\nmyzset\r\n$1\r\na\r\n"
	out, err := runTestPayload(s, payload)
	require.NoError(t, err)
	require.Contains(t, out, ":1\r\n")
}
