package tests

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSystem_ACL_GETUSER_Default(t *testing.T) {
	s := createTestStore()
	payload := "*3\r\n$3\r\nACL\r\n$7\r\nGETUSER\r\n$7\r\ndefault\r\n"
	out, err := runTestPayload(s, payload)
	require.NoError(t, err)
	require.Contains(t, out, "$9\r\npasswords\r\n*0\r\n")
}

func TestSystem_CONFIG_GET(t *testing.T) {
	s := createTestStore()
	payload := "*3\r\n$6\r\nCONFIG\r\n$3\r\nGET\r\n$3\r\ndir\r\n"
	out, err := runTestPayload(s, payload)
	require.NoError(t, err)
	require.Contains(t, out, "$3\r\ndir")
}

func TestSystem_MULTI_EXEC(t *testing.T) {
	s := createTestStore()
	payload := "*1\r\n$5\r\nMULTI\r\n" +
		"*3\r\n$3\r\nSET\r\n$1\r\nk\r\n$1\r\nv\r\n" +
		"*1\r\n$4\r\nEXEC\r\n"
	out, err := runTestPayload(s, payload)
	require.NoError(t, err)
	require.Contains(t, out, "+OK\r\n+QUEUED\r\n*1\r\n+OK\r\n")
}

func TestSystem_MULTI_DISCARD(t *testing.T) {
	s := createTestStore()
	payload := "*1\r\n$5\r\nMULTI\r\n" +
		"*3\r\n$3\r\nSET\r\n$1\r\nk\r\n$1\r\nv\r\n" +
		"*1\r\n$7\r\nDISCARD\r\n"
	out, err := runTestPayload(s, payload)
	require.NoError(t, err)
	require.Contains(t, out, "+OK\r\n+QUEUED\r\n+OK\r\n")
}
