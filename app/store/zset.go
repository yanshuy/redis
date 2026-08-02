package store

import (
	"fmt"
	"strconv"
)

type Z struct {
	member string
	score  float64
}

type znode struct {
	value Z
	next  *znode
}

type skiplist struct {
	head *znode
}

type Zset struct {
	dict map[string]Z
	list skiplist
}

func (s Zset) len() int {
	return len(s.dict)
}

func NewZset() Zset {
	return Zset{
		dict: make(map[string]Z),
		list: skiplist{},
	}
}

func insert(s Zset, z Z) Zset {
	s.dict[z.member] = z

	if s.list.head == nil {
		s.list.head = &znode{value: z}
		return s
	}
	cur := s.list.head
	for cur != nil && cur.value.score < z.score {
		cur = cur.next
	}

	cur.next = &znode{
		value: z,
		next:  cur.next,
	}
	return s
}

func (rs *RedisStore) Zadd(key string, score_member []string) (s int, err error) {
	m, ok := rs.Look(key)
	if !ok {
		m = rs.NewStoreMember(key, ZSET)
	}
	if m.data.Type != ZSET {
		return 0, fmt.Errorf("provided key '%s' holds some other data", key)
	}

	for i := 0; i+1 < len(score_member); i += 2 {
		score_str := score_member[i]
		member := score_member[i+1]

		score, err := strconv.ParseFloat(score_str, 64)
		if err != nil {
			return 0, err
		}
		m.data.Zset = insert(m.data.Zset, Z{member, score})
	}

	rs.TouchWatchedKey(key)
	return m.data.Zset.len(), nil
}
