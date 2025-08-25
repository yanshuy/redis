package store

import (
	"fmt"
	"time"
)

type RedisStore struct {
	Store map[string]*StoreMember
}

func (rs RedisStore) KeyExists(key string) (DataStructType, bool) {
	m, ok := rs.Store[key]
	return m.data.Type, ok
}

var R = RedisStore{
	Store: make(map[string]*StoreMember),
}

func (rs RedisStore) removeMemberAfter(ttl_ms int64, key string) {
	timer := time.NewTimer(time.Duration(ttl_ms) * time.Millisecond)
	<-timer.C
	delete(rs.Store, key)
}

func (rs RedisStore) Set(key string, val string, ttl_ms int64) {
	mem := NewStoreMember(String)
	mem.AssignValue(val)
	if ttl_ms > 0 {
		mem.ExpiryAt = time.Now().Add(time.Duration(ttl_ms) * time.Millisecond)
		go rs.removeMemberAfter(ttl_ms, key)
	}
	rs.Store[key] = mem
}

func (rs RedisStore) Get(key string) (string, bool) {
	mem, ok := rs.Store[key]
	return *mem.data.String, ok
}

func (rs RedisStore) Rpush(key string, val []string) (int, error) {
	var mem *StoreMember
	if t, ok := rs.KeyExists(key); ok {
		if t != List {
			return 0, fmt.Errorf("provided key '%s' does not hold a List", key)
		}
		mem = rs.Store[key]
		mem.AssignValue(val)
	} else {
		mem = NewStoreMember(List)
		mem.AssignValue(val)
		rs.Store[key] = mem
	}
	fmt.Printf("%+v\n", mem)
	return len(mem.data.List), nil
}
