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
	Z
	next *znode
}

type list struct {
	head *znode
}

type Zset struct {
	dict map[string]Z
	list list
}

func (s Zset) len() int {
	return len(s.dict)
}

func NewZset() Zset {
	return Zset{
		dict: make(map[string]Z),
		list: list{head: &znode{}},
	}
}

func (s *Zset) remove(z Z) bool {
	prev := s.list.head
	for prev.next != nil {
		if prev.next.member == z.member && prev.next.score == z.score {
			prev.next = prev.next.next
			return true
		}
		prev = prev.next
	}

	return false
}

func (s *Zset) insert(z Z) bool {
	old, ok := s.dict[z.member]
	if ok {
		s.remove(old)
	}

	s.dict[z.member] = z
	node := &znode{Z: z}

	cur := s.list.head
	for cur.next != nil {
		if cur.next.score > z.score ||
			(cur.next.score == z.score && cur.next.member > z.member) {
			break
		}
		cur = cur.next
	}

	node.next = cur.next
	cur.next = node

	return !ok
}

func (rs *RedisStore) Zadd(key string, score_member []string) (s int, err error) {
	m, ok := rs.Look(key)
	if !ok {
		m = rs.NewStoreMember(key, ZSET)
	}
	if m.data.Type != ZSET {
		return 0, fmt.Errorf("provided key '%s' holds some other data", key)
	}

	inserts := 0
	for i := 0; i+1 < len(score_member); i += 2 {
		score_str := score_member[i]
		member := score_member[i+1]

		score, err := strconv.ParseFloat(score_str, 64)
		if err != nil {
			return 0, err
		}
		if m.data.Zset.insert(Z{member, score}) {
			inserts++
		}
	}

	rs.TouchWatchedKey(key)
	return inserts, nil
}
