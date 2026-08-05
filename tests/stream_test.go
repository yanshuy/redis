package tests

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStream_XADD_ExplicitID(t *testing.T) {
	s := createTestStore()
	payload := "*5\r\n$4\r\nXADD\r\n$8\r\nmystream\r\n$6\r\n1000-1\r\n$1\r\nk\r\n$1\r\nv\r\n"
	out, err := runTestPayload(s, payload)
	require.NoError(t, err)
	require.Equal(t, "$6\r\n1000-1\r\n", out)
}

func TestStream_XADD_InvalidIDZero(t *testing.T) {
	s := createTestStore()
	payload := "*5\r\n$4\r\nXADD\r\n$8\r\nmystream\r\n$3\r\n0-0\r\n$1\r\nk\r\n$1\r\nv\r\n"
	out, err := runTestPayload(s, payload)
	require.NoError(t, err)
	require.Contains(t, out, "-ERR")
	require.Contains(t, out, "must be greater than 0-0")
}

func TestStream_XADD_EqualOrSmallerID(t *testing.T) {
	s := createTestStore()
	payload := "*5\r\n$4\r\nXADD\r\n$8\r\nmystream\r\n$6\r\n1000-1\r\n$1\r\nk\r\n$1\r\nv\r\n" +
		"*5\r\n$4\r\nXADD\r\n$8\r\nmystream\r\n$6\r\n1000-1\r\n$1\r\nk\r\n$1\r\nv\r\n"
	out, err := runTestPayload(s, payload)
	require.NoError(t, err)
	require.Contains(t, out, "-ERR")
	require.Contains(t, out, "equal or smaller")
}

func TestStream_XRANGE_Full(t *testing.T) {
	s := createTestStore()
	payload := "*5\r\n$4\r\nXADD\r\n$8\r\nmystream\r\n$6\r\n1000-1\r\n$1\r\na\r\n$1\r\nb\r\n" +
		"*4\r\n$6\r\nXRANGE\r\n$8\r\nmystream\r\n$1\r\n-\r\n$1\r\n+\r\n"
	out, err := runTestPayload(s, payload)
	require.NoError(t, err)
	require.Contains(t, out, "*1\r\n")
	require.Contains(t, out, "$6\r\n1000-1")
}
