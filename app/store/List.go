package store

import (
	"fmt"
	"math"
	"slices"
	"time"
)

func (rs *RedisStore) Rpush(key string, val []string) (int, error) {
	var mem *StoreMember
	if m, ok := rs.Look(key); ok {
		if m.data.Type != List {
			return 0, fmt.Errorf("provided key '%s' holds some other data", key)
		}
		mem = m
		m.AssignValue(val)
	} else {
		mem = NewStoreMember(List)
		mem.AssignValue(val)
		rs.Store[key] = mem
	}
	BlpopDeque(mem.data.List[0])
	return len(mem.data.List), nil
}

func (rs *RedisStore) Lpush(key string, val []string) (int, error) {
	var mem *StoreMember
	if m, ok := rs.Look(key); ok {
		if m.data.Type != List {
			return 0, fmt.Errorf("provided key '%s' does not hold a list", key)
		}
		mem = m
		slices.Reverse(val)
		m.data.List = append(val, m.data.List...)
	} else {
		mem = NewStoreMember(List)
		mem.AssignValue(val)
		rs.Store[key] = mem
	}
	BlpopDeque(mem.data.List[0])
	return len(mem.data.List), nil
}

func (rs *RedisStore) Lpop(key string, popCount int) ([]string, error) {
	if m, ok := rs.Look(key); ok {
		if m.data.Type != List {
			return nil, fmt.Errorf("provided key '%s' does not hold a list", key)
		}
		if popCount > len(m.data.List) {
			popCount = len(m.data.List)
		}
		popped := make([]string, 0, popCount)
		for _, item := range m.data.List[:popCount] {
			popped = append(popped, item)
		}
		m.data.List = m.data.List[popCount:]
		return popped, nil
	} else {
		return nil, nil
	}
}

func (rs *RedisStore) Llen(key string) (int, error) {
	if m, ok := rs.Look(key); ok {
		if m.data.Type != List {
			return 0, fmt.Errorf("provided key '%s' does not hold a list", key)
		}
		return len(m.data.List), nil
	} else {
		return 0, nil
	}
}

func (rs *RedisStore) Lrange(key string, startIdx int, endIdx int) ([]string, error) {
	if m, ok := rs.Look(key); ok {
		if m.data.Type != List {
			return nil, fmt.Errorf("provided key '%s' holds some other data", key)
		}
		if startIdx < 0 {
			startIdx = max(len(m.data.List)+startIdx, 0)
		}
		if endIdx < 0 {
			endIdx = max(len(m.data.List)+endIdx, 0)
		}
		if startIdx >= endIdx || startIdx > len(m.data.List) {
			return nil, nil
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
		return nil, nil
	}
}

var blpopQueue = []chan string{}

func BlpopAdd() chan string {
	c := make(chan string)
	blpopQueue = append(blpopQueue, c)
	return c
}

func BlpopDeque(item string) {
	if len(blpopQueue) > 0 {
		d := blpopQueue[0]
		blpopQueue = blpopQueue[1:]
		d <- item
	}
}

func (rs *RedisStore) Blpop(key string, timeout_s float64) (<-chan string, error) {
	item, err := rs.Lpop(key, 1)
	if err != nil {
		return nil, err
	}
	msgChan := make(chan string)
	if len(item) == 1 {
		msgChan <- item[0]
		close(msgChan)
		return msgChan, nil
	}

	if timeout_s == 0 {
		timeout_s = math.MaxInt32
	}

	timer := time.NewTimer(time.Duration(timeout_s * float64(time.Second)))
	ch := BlpopAdd()

	go func() {
		defer close(msgChan)
		defer timer.Stop()
		select {
		case item := <-ch:
			msgChan <- item
			rs.mu.Lock()
			rs.Lpop(key, 1)
			rs.mu.Unlock()
			return
		case <-timer.C:
			msgChan <- ""
			return
		}
	}()

	return msgChan, nil
}
