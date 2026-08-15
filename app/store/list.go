package store

import (
	"slices"

	resp "github.com/codecrafters-io/redis-starter-go/app/Resp"
)

func (rs *RedisStore) NotifyBlockedClient(key string, values []string) []string {
	blocked := rs.BlockedKeys[key]
	if len(blocked) == 0 {
		return values
	}

	remainingValues := values

	for i, c := range blocked {
		if len(remainingValues) == 0 {
			rs.BlockedKeys[key] = blocked[i:]
			return remainingValues
		}
		c.QueueMessage(resp.NewData(resp.Array, []string{key, remainingValues[0]}))
		c.Blop.Cancel()
		remainingValues = remainingValues[1:]
	}

	delete(rs.BlockedKeys, key)
	return remainingValues
}

func (rs *RedisStore) Rpush(key string, elements []string) (int, error) {
	val, ok := rs.Look(key)
	if !ok {
		val = NewValue(LIST, 0)
		rs.Store[key] = val
	}
	list, err := As[List](val)
	if err != nil {
		return 0, err
	}

	totalLen := len(list) + len(elements)
	elements = rs.NotifyBlockedClient(key, elements)

	if len(elements) > 0 {
		list = append(list, elements...)
		val.Obj = list
	}

	rs.TouchWatchedKey(key)
	return totalLen, nil
}

func (rs *RedisStore) Lpush(key string, elements []string) (int, error) {
	val, ok := rs.Look(key)
	if !ok {
		val = NewValue(LIST, 0)
		rs.Store[key] = val
	}
	list, err := As[List](val)
	if err != nil {
		return 0, err
	}

	totalLen := len(list) + len(elements)
	elements = rs.NotifyBlockedClient(key, elements)

	if len(elements) > 0 {
		slices.Reverse(elements)
		list = append(elements, list...)
		val.Obj = list
	}

	rs.TouchWatchedKey(key)
	return totalLen, nil
}

func (rs *RedisStore) Lpop(key string, popCount int) ([]string, error) {
	val, ok := rs.Look(key)
	if !ok {
		return nil, nil
	}
	list, err := As[List](val)
	if err != nil {
		return nil, err
	}

	if popCount > len(list) {
		popCount = len(list)
	}
	popped := list[:popCount]

	list = list[popCount:]
	val.Obj = list

	return popped, nil
}

func (rs *RedisStore) Llen(key string) (int, error) {
	val, ok := rs.Look(key)
	if !ok {
		return 0, nil
	}
	list, err := As[List](val)
	if err != nil {
		return 0, err
	}

	return len(list), nil
}

func (rs *RedisStore) Lrange(key string, startIdx int, endIdx int) ([]string, error) {
	val, ok := rs.Look(key)
	if !ok {
		return []string{}, nil
	}
	list, err := As[List](val)
	if err != nil {
		return nil, err
	}

	n := len(list)
	if startIdx < 0 {
		startIdx = max(n+startIdx, 0)
	}
	if endIdx < 0 {
		endIdx = max(n+endIdx, 0)
	}
	if startIdx > endIdx || startIdx >= len(list) {
		return []string{}, nil
	}
	if endIdx >= len(list) {
		endIdx = len(list) - 1
	}

	return list[startIdx : endIdx+1], nil
}
