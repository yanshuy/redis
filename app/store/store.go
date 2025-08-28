package store

import (
	"sync"
	"time"
)

type RedisStore struct {
	Store    map[string]*StoreMember
	Blockers map[string][]chan struct{}
	mu       sync.Mutex
}

func (rs *RedisStore) Look(key string) (*StoreMember, bool) {
	m, ok := rs.Store[key]
	return m, ok
}

var DB = RedisStore{
	Store:    make(map[string]*StoreMember),
	Blockers: make(map[string][]chan struct{}),
	mu:       sync.Mutex{},
}

// func (rs *RedisStore) subscribeKey(key string) chan struct{} {
// 	ch := make(chan struct{})
// 	rs.Listeners[key] = append(rs.Listeners[key], ch)
// 	return ch
// }

// func (rs *RedisStore) unsubscribeAll(key string) {
// 	delete(rs.Listeners, key)
// }

// func (rs *RedisStore) unsubscribe(key string, ch chan struct{}) (ok bool) {
// 	chans, ok := rs.Listeners[key]
// 	if !ok {
// 		return false
// 	}
// 	newChans := []chan struct{}{}
// 	for _, c := range chans {
// 		if c == ch {
// 			ok = true
// 			continue
// 		}
// 		newChans = append(newChans, c)
// 	}
// 	rs.Listeners[key] = newChans
// 	if len(newChans) == 0 {
// 		delete(rs.Listeners, key)
// 	}
// 	close(ch)
// 	return ok
// }

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

func (rs *RedisStore) RemoveMemberAfter(ttl_ms int64, key string) {
	timer := time.NewTimer(time.Duration(ttl_ms) * time.Millisecond)
	<-timer.C
	delete(rs.Store, key)
}

func (rs *RedisStore) Set(key string, val string, ttl_ms int64) {
	mem := NewStoreMember(String)
	mem.AssignValue(val)
	if ttl_ms > 0 {
		mem.ExpiryAt = time.Now().Add(time.Duration(ttl_ms) * time.Millisecond)
		go rs.RemoveMemberAfter(ttl_ms, key)
	}
	rs.Store[key] = mem
}

func (rs *RedisStore) Get(key string) (string, bool) {
	mem, ok := rs.Store[key]
	return mem.data.String, ok
}

func (rs *RedisStore) AddBlocker(key string, c chan struct{}) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.Blockers[key] = append(rs.Blockers[key], c)
}

func (rs *RedisStore) RemovBlocker(key string, ch chan struct{}) (ok bool) {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	chans, ok := rs.Blockers[key]
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
		delete(rs.Blockers, key)
	}
	close(ch)
	return ok
}

func (rs *RedisStore) NotifyBlockers(key string) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	chans, ok := rs.Blockers[key]
	if !ok {
		return
	}

	ch := chans[0]
	rs.Blockers[key] = chans[1:]

	ch <- struct{}{}
	close(ch)
}
