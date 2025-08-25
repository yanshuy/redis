package store

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStore_SetAndGet(t *testing.T) {
	DB = RedisStore{Store: make(map[string]*StoreMember)}
	DB.Set("foo", "bar", 0)
	v, ok := DB.Get("foo")
	require.True(t, ok)
	require.Equal(t, "bar", v)
}

func TestStore_GetMissing(t *testing.T) {
	DB = RedisStore{Store: make(map[string]*StoreMember)}
	_, ok := DB.Get("missing")
	require.False(t, ok)
}

func TestStore_SetExpiry(t *testing.T) {
	DB = RedisStore{Store: make(map[string]*StoreMember)}
	go DB.Set("temp", "val", 10)
	time.Sleep(11 * time.Millisecond)
	_, ok := DB.Get("temp")
	require.False(t, ok, "expected key to be expired and removed")
}

func TestStore_SetExpiry2(t *testing.T) {
	DB = RedisStore{Store: make(map[string]*StoreMember)}
	go DB.Set("temp", "val", 10)
	time.Sleep(8 * time.Millisecond)
	_, ok := DB.Get("temp")
	require.True(t, ok)
}
