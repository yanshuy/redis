package store

import (
	"fmt"
	"time"
)

type RedisStore struct {
	Store map[string]*StoreMember
}

func (rs RedisStore) Look(key string) (*StoreMember, bool) {
	m, ok := rs.Store[key]
	return m, ok
}

var DB = RedisStore{
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
	return mem.data.String, ok
}

func (rs RedisStore) Rpush(key string, val []string) (int, error) {
	var mem *StoreMember
	if m, ok := rs.Look(key); ok {
		if m.data.Type != List {
			return 0, fmt.Errorf("provided key '%s' holds some other data", key)
		}
		mem = rs.Store[key]
		mem.AssignValue(val)
	} else {
		mem = NewStoreMember(List)
		mem.AssignValue(val)
		rs.Store[key] = mem
	}
	// fmt.Printf("%+v\n", mem)
	return len(mem.data.List), nil
}

func (rs RedisStore) Lrange(key string, startIdx int, endIdx int) ([]string, error) {
	if m, ok := rs.Look(key); ok {
		if m.data.Type != List {
			return nil, fmt.Errorf("provided key '%s' holds some other data", key)
		}
		if startIdx < 0 {
			startIdx = len(m.data.List) - startIdx
		}
		if endIdx < 0 {
			endIdx = len(m.data.List) - endIdx
		}
		if startIdx >= endIdx || startIdx > len(m.data.List) {
			return []string{}, nil
		}
		if endIdx >= len(m.data.List) {
			endIdx = len(m.data.List) - 1
		}
		items := make([]string, 0, endIdx-startIdx)
		for i := startIdx; i < endIdx+1; i++ {
			items = append(items, m.data.List[i])
		}
		return items, nil

	} else {
		return []string{}, nil
	}
}
