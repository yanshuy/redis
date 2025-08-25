package store

import (
	"time"
)

type StoreMember struct {
	ExpiryAt time.Time
	data     string
}

type RedisStore struct {
	Store map[string]StoreMember
}

var R = RedisStore{
	Store: make(map[string]StoreMember),
}

func (rs RedisStore) RemoveExpiredMembers() {
	// rs.expired = make(chan StoreMember, 100)
	// for mem := range expiredMembers {
	// 	delete(rs.store[])
	// }
}

func (rs RedisStore) removeMemberAfter(ttl_ms int64, key string) {
	timer := time.NewTimer(time.Duration(ttl_ms) * time.Millisecond)
	<-timer.C
	delete(rs.Store, key)
}

func (rs RedisStore) Set(key string, val string, ttl_ms int64) {
	mem := StoreMember{
		data: val,
	}
	if ttl_ms > 0 {
		mem.ExpiryAt = time.Now().Add(time.Duration(ttl_ms) * time.Millisecond)
		go rs.removeMemberAfter(ttl_ms, key)
	}
	rs.Store[key] = mem
}

func (rs RedisStore) Get(key string) (string, bool) {
	mem, ok := rs.Store[key]
	return mem.data, ok
}
