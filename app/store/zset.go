package store

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

func (rs *RedisStore) Zadd(key string, scores []float64, member []string) int {
	m, _ := rs.Look(key)
	if m == nil {
		m = rs.NewStoreMember(key, ZSET)
	}

	z := m.Data.Zset

	inserts := 0
	for i := range scores {
		score := scores[i]
		member := member[i]
		if z.insert(Z{member, score}) {
			inserts++
		}
	}

	rs.TouchWatchedKey(key)
	return inserts
}

func (rs *RedisStore) Zrank(key, member string) (int, bool) {
	m, _ := rs.Look(key)
	z := m.Data.Zset

	if _, ok := z.dict[member]; !ok {
		return 0, false
	}

	rank := -1
	cur := z.list.head
	for cur != nil {
		if cur.member == member {
			break
		}
		rank++
		cur = cur.next
	}
	return rank, true
}

func (rs *RedisStore) Zrange(key string, start int, end int) []string {
	m, _ := rs.Look(key)
	z := m.Data.Zset

	if z.len() < start {
		return []string{}
	}

	cur := z.list.head.next
	for range start {
		cur = cur.next
	}

	list := make([]string, 0, end-start+1)
	for i := start; cur != nil && i <= end; i++ {
		list = append(list, cur.Z.member)
		cur = cur.next
	}

	return list
}
