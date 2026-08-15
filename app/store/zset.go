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

func (rs *RedisStore) Zadd(key string, scores []float64, member []string) (int, error) {
	val, ok := rs.Look(key)
	if !ok {
		val = NewValue(ZSET, 0)
		rs.Store[key] = val
	}
	z, err := As[Zset](val)
	if err != nil {
		return 0, err
	}

	inserts := 0
	for i := range scores {
		score := scores[i]
		member := member[i]
		if z.insert(Z{member, score}) {
			inserts++
		}
	}

	rs.TouchWatchedKey(key)
	return inserts, nil
}

func (rs *RedisStore) Zrank(key, member string) (int, error) {
	val, ok := rs.Look(key)
	if !ok {
		return -1, nil
	}
	z, err := As[Zset](val)
	if err != nil {
		return -1, err
	}
	if _, ok := z.dict[member]; !ok {
		return -1, nil
	}

	rank := 0
	cur := z.list.head.next
	for cur != nil {
		if cur.member == member {
			return rank, nil
		}
		rank++
		cur = cur.next
	}
	return -1, nil
}

func (rs *RedisStore) Zrange(key string, start int, end int) ([]string, error) {
	val, ok := rs.Look(key)
	if !ok {
		return []string{}, nil
	}
	z, err := As[Zset](val)
	if err != nil {
		return nil, err
	}

	n := z.len()
	if n == 0 {
		return []string{}, nil
	}
	if start < 0 {
		start = max(n+start, 0)
	}
	if end < 0 {
		end = max(n+end, 0)
	}
	if start >= n || start > end {
		return []string{}, nil
	}
	if end >= n {
		end = n - 1
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

	return list, nil
}

func (rs *RedisStore) Zcard(key string) (int, error) {
	val, ok := rs.Look(key)
	if !ok {
		return 0, nil
	}
	z, err := As[Zset](val)
	if err != nil {
		return 0, err
	}

	return z.len(), nil
}

func (rs *RedisStore) Zscore(key, member string) (float64, bool, error) {
	val, ok := rs.Look(key)
	if !ok {
		return 0, false, nil
	}
	z, err := As[Zset](val)
	if err != nil {
		return 0, false, err
	}

	v, ok := z.dict[member]
	if !ok {
		return 0, false, nil
	}
	return v.score, true, nil
}

func (rs *RedisStore) Zrem(key, member string) (int, error) {
	val, ok := rs.Look(key)
	if !ok {
		return 0, nil
	}
	z, err := As[Zset](val)
	if err != nil {
		return 0, err
	}

	v, ok := z.dict[member]
	if !ok {
		return 0, nil
	}
	z.remove(v)
	delete(z.dict, member)

	rs.TouchWatchedKey(key)
	return 1, nil
}
