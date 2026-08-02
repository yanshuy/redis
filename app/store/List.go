package store

import (
	"fmt"
	"math"
	"slices"
	"time"
)

func (rs *RedisStore) Rpush(key string, val []string) (int, error) {
	var mem *Value
	if m, ok := rs.Look(key); ok {
		if m.Data.Type != LIST {
			return 0, fmt.Errorf("provided key '%s' holds some other data", key)
		}
		mem = m
		mem.Data.List = append(mem.Data.List, val...)
	} else {
		mem = rs.NewStoreMember(key, LIST)
		mem.Data.List = append(mem.Data.List, val...)
	}
	go rs.NotifyListener(key)
	rs.TouchWatchedKey(key)
	return len(mem.Data.List), nil
}

func (rs *RedisStore) Lpush(key string, val []string) (int, error) {
	var mem *Value
	if m, ok := rs.Look(key); ok {
		if m.Data.Type != LIST {
			return 0, fmt.Errorf("provided key '%s' does not hold a list", key)
		}
		mem = m
		slices.Reverse(val)
		m.Data.List = append(val, m.Data.List...)
	} else {
		mem = rs.NewStoreMember(key, LIST)
		mem.Data.List = append(mem.Data.List, val...)
	}
	go rs.NotifyListener(key)
	rs.TouchWatchedKey(key)
	return len(mem.Data.List), nil
}

func (rs *RedisStore) Lpop(key string, popCount int) ([]string, error) {
	if m, ok := rs.Look(key); ok {
		if m.Data.Type != LIST {
			return nil, fmt.Errorf("provided key '%s' does not hold a list", key)
		}
		if popCount > len(m.Data.List) {
			popCount = len(m.Data.List)
		}
		popped := make([]string, 0, popCount)
		for _, item := range m.Data.List[:popCount] {
			popped = append(popped, item)
		}
		m.Data.List = m.Data.List[popCount:]
		rs.TouchWatchedKey(key)
		return popped, nil
	} else {
		return nil, nil
	}
}

func (rs *RedisStore) Llen(key string) (int, error) {
	if m, ok := rs.Look(key); ok {
		if m.Data.Type != LIST {
			return 0, fmt.Errorf("provided key '%s' does not hold a LIST", key)
		}
		return len(m.Data.List), nil
	} else {
		return 0, nil
	}
}

func (rs *RedisStore) Lrange(key string, startIdx int, endIdx int) ([]string, error) {
	if m, ok := rs.Look(key); ok {
		if m.Data.Type != LIST {
			return nil, fmt.Errorf("provided key '%s' holds some other data", key)
		}
		if startIdx < 0 {
			startIdx = max(len(m.Data.List)+startIdx, 0)
		}
		if endIdx < 0 {
			endIdx = max(len(m.Data.List)+endIdx, 0)
		}
		if startIdx > endIdx || startIdx > len(m.Data.List) {
			return []string{}, nil
		}
		if endIdx >= len(m.Data.List) {
			endIdx = len(m.Data.List) - 1
		}
		items := make([]string, 0, endIdx-startIdx)
		for i := startIdx; i < endIdx+1; i++ {
			items = append(items, m.Data.List[i])
		}
		return items, nil

	} else {
		return []string{}, nil
	}
}

func (rs *RedisStore) Blpop(key string, timeout_s float64) (<-chan string, error) {
	item, err := rs.Lpop(key, 1)
	if err != nil {
		return nil, err
	}

	msgChan := make(chan string, 1)

	if len(item) == 1 {
		msgChan <- item[0]
		close(msgChan)
		return msgChan, nil
	}

	if timeout_s <= 0 {
		timeout_s = math.MaxInt32
	}

	timer := time.NewTimer(time.Duration(timeout_s * float64(time.Second)))
	ch := rs.subscribe(key)

	go func() {
		defer timer.Stop()
		select {
		case <-ch:
			item, _ := rs.Lpop(key, 1)
			msgChan <- item[0]
		case <-timer.C:
			msgChan <- ""
		}
		close(msgChan)
		rs.unsubscribe(key, ch)
	}()
	return msgChan, nil
}
