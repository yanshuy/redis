package tests

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRequest_PING(t *testing.T) {
	s := createTestStore()
	out, err := runTestPayload(s, "*1\r\n$4\r\nPING\r\n")
	require.NoError(t, err)
	require.Equal(t, "+PONG\r\n", out)
}

func TestRequest_ECHO_NoArgs(t *testing.T) {
	s := createTestStore()
	out, err := runTestPayload(s, "*1\r\n$4\r\nECHO\r\n")
	require.NoError(t, err)
	require.Contains(t, out, "-ERR")
}

func TestRequest_Unknown(t *testing.T) {
	s := createTestStore()
	out, err := runTestPayload(s, "*1\r\n$3\r\nFOO\r\n")
	require.NoError(t, err)
	require.Contains(t, out, "-ERR")
}

func TestRequest_ECHO_WithArgs(t *testing.T) {
	s := createTestStore()
	payload := "*2\r\n$4\r\nECHO\r\n$5\r\nHELLO\r\n"
	out, err := runTestPayload(s, payload)
	require.NoError(t, err)
	require.Equal(t, "$5\r\nHELLO\r\n", out)
}

func TestRequest_PING_Array(t *testing.T) {
	s := createTestStore()
	payload := "*1\r\n$4\r\nPING\r\n"
	out, err := runTestPayload(s, payload)
	require.NoError(t, err)
	require.Equal(t, "+PONG\r\n", out)
}

func TestRequest_GET_MissingKey(t *testing.T) {
	s := createTestStore()
	payload := "*2\r\n$3\r\nGET\r\n$3\r\nFOO\r\n"
	out, err := runTestPayload(s, payload)
	require.NoError(t, err)
	require.Equal(t, "$-1\r\n", out)
}

func TestRequest_SET_Then_GET(t *testing.T) {
	s := createTestStore()
	payload := "*3\r\n$3\r\nSET\r\n$3\r\nFOO\r\n$3\r\nBAR\r\n*2\r\n$3\r\nGET\r\n$3\r\nFOO\r\n"
	out, err := runTestPayload(s, payload)
	require.NoError(t, err)
	require.Contains(t, out, "+OK\r\n")
	require.Contains(t, out, "$3\r\nBAR\r\n")
}

func TestRequest_RPUSH_NewList(t *testing.T) {
	s := createTestStore()
	payload := "*5\r\n$5\r\nRPUSH\r\n$6\r\nmylist\r\n$1\r\na\r\n$1\r\nb\r\n$1\r\nc\r\n"
	out, err := runTestPayload(s, payload)
	require.NoError(t, err)
	require.Equal(t, ":3\r\n", out)
}

func TestRequest_LRANGE_Full(t *testing.T) {
	s := createTestStore()
	payload := "*5\r\n$5\r\nRPUSH\r\n$6\r\nmylist\r\n$1\r\na\r\n$1\r\nb\r\n$1\r\nc\r\n*4\r\n$6\r\nLRANGE\r\n$6\r\nmylist\r\n$1\r\n0\r\n$1\r\n2\r\n"
	out, err := runTestPayload(s, payload)
	require.NoError(t, err)
	require.Contains(t, out, "*3")
	require.Contains(t, out, "$1\r\na")
	require.Contains(t, out, "$1\r\nb")
	require.Contains(t, out, "$1\r\nc")
}

func TestConfig_Path(t *testing.T) {
	s := createTestStore()
	cwd, _ := os.Getwd()
	dir := s.Config.Dir
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(cwd, dir)
	}
	j := path.Join("/root", s.Config.Dir, s.Config.Dbfilename)
	fmt.Println(j)
}
