package store

import (
	"time"
)

type RedisStore struct {
	Store map[string]StoreMember
}

var R = RedisStore{
	Store: make(map[string]StoreMember),
}

func (rs RedisStore) removeMemberAfter(ttl_ms int64, key string) {
	timer := time.NewTimer(time.Duration(ttl_ms) * time.Millisecond)
	<-timer.C
	delete(rs.Store, key)
}

func (rs RedisStore) Set(key string, val string, ttl_ms int64) {
	mem, _ := NewStoreMember(String, val)
	if ttl_ms > 0 {
		mem.ExpiryAt = time.Now().Add(time.Duration(ttl_ms) * time.Millisecond)
		go rs.removeMemberAfter(ttl_ms, key)
	}
	rs.Store[key] = *mem
}

func (rs RedisStore) Get(key string) (string, bool) {
	mem, ok := rs.Store[key]
	return mem.data.String, ok
}

func (rs RedisStore) Rpush(key string, val []string) {
	mem, _ := NewStoreMember(List, val)
	rs.Store[key] = *mem
}
