package store

import (
	"strings"
	"sync"
	"time"
)

type StoreMember struct {
	ExpiryAt int64
	data     Data
}

func (rs *RedisStore) NewStoreMember(key string, t DataType) *StoreMember {
	m := &StoreMember{
		data: Data{Type: t},
	}
	rs.Store[key] = m
	return m
}

type RedisStore struct {
	Store     map[string]*StoreMember
	Config    map[string]string
	Listeners map[string][]chan struct{}
	mu        sync.Mutex
}

func (rs *RedisStore) Look(key string) (*StoreMember, bool) {
	m, ok := rs.Store[key]
	return m, ok
}

var RDB RedisStore = RedisStore{
	Store:     make(map[string]*StoreMember),
	Config:    make(map[string]string),
	Listeners: make(map[string][]chan struct{}),
	mu:        sync.Mutex{},
}

type DataType int

const (
	STRING DataType = iota
	LIST
	STREAM
)

type Data struct {
	Type   DataType
	String string
	Stream *StreamObj
	List   []string
}

func (rs *RedisStore) RemoveMemberAfter(ttl_ms int64, key string) {
	timer := time.NewTimer(time.Duration(ttl_ms) * time.Millisecond)
	<-timer.C
	delete(rs.Store, key)
}

func (rs *RedisStore) Set(key string, val string, ttl_ms int64) {
	mem := rs.NewStoreMember(key, STRING)
	mem.data.String = val
	if ttl_ms > 0 {
		mem.ExpiryAt = time.Now().UnixMilli() + ttl_ms
		go rs.RemoveMemberAfter(ttl_ms, key)
	}
}

func (rs *RedisStore) Get(key string) (string, bool) {
	mem, ok := rs.Store[key]
	if !ok {
		return "", false
	}
	if mem.ExpiryAt > 0 && mem.ExpiryAt <= time.Now().UnixMilli() {
		delete(rs.Store, key)
		return "", false
	}
	if mem.data.Type != STRING {
		return "", false
	}
	return mem.data.String, ok
}

func (rs *RedisStore) Keys(pattern string) []string {
	subStr := strings.ReplaceAll(pattern, "*", "")

	ans := make([]string, 0)
	for key := range rs.Store {
		if strings.Contains(key, subStr) {
			ans = append(ans, key)
		}
	}
	return ans
}

func (rs *RedisStore) Type(key string) string {
	if m, ok := rs.Look(key); ok {
		switch m.data.Type {
		case STRING:
			return "string"
		case LIST:
			return "list"
		case STREAM:
			return "stream"
		}
	}
	return "none"
}

func (rs *RedisStore) subscribe(key string) chan struct{} {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	ch := make(chan struct{})
	rs.Listeners[key] = append(rs.Listeners[key], ch)
	return ch
}

func (rs *RedisStore) unsubscribe(key string, ch chan struct{}) (ok bool) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	chans, ok := rs.Listeners[key]
	if !ok {
		return false
	}

	for i, c := range chans {
		if c == ch {
			chans = append(chans[:i], chans[i+1:]...)
			ok = true
		}
	}

	if len(chans) == 0 {
		delete(rs.Listeners, key)
	}
	close(ch)
	return ok
}

// func (rs *RedisStore) notifySubscribers(key string, events ...func()) {
// 	chans, ok := rs.Listeners[key]
// 	if !ok {
// 		return
// 	}
// 	for _, event := range events {
// 		event()
// 	}
// 	for _, c := range chans {
// 		c <- struct{}{}
// 	}
// }

func (rs *RedisStore) NotifyListener(key string) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	chans, ok := rs.Listeners[key]
	if !ok {
		return
	}
	ch := chans[0]
	rs.Listeners[key] = chans[1:]

	ch <- struct{}{}
}
