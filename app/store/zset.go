package store

import (
	"fmt"
	"math"
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

const (
	MIN_LATITUDE  = -85.05112878
	MAX_LATITUDE  = 85.05112878
	MIN_LONGITUDE = -180.0
	MAX_LONGITUDE = 180.0

	LATITUDE_RANGE  = MAX_LATITUDE - MIN_LATITUDE
	LONGITUDE_RANGE = MAX_LONGITUDE - MIN_LONGITUDE
)

func spreadInt32ToInt64(v uint32) uint64 {
	result := uint64(v)
	result = (result | (result << 16)) & 0x0000FFFF0000FFFF
	result = (result | (result << 8)) & 0x00FF00FF00FF00FF
	result = (result | (result << 4)) & 0x0F0F0F0F0F0F0F0F
	result = (result | (result << 2)) & 0x3333333333333333
	result = (result | (result << 1)) & 0x5555555555555555
	return result
}

func interleave(x, y uint32) uint64 {
	xSpread := spreadInt32ToInt64(x)
	ySpread := spreadInt32ToInt64(y)

	return xSpread | (ySpread << 1)
}

func encode(longitude, latitude float64) uint64 {
	normalizedLatitude := math.Pow(2, 26) * (latitude - MIN_LATITUDE) / LATITUDE_RANGE
	normalizedLongitude := math.Pow(2, 26) * (longitude - MIN_LONGITUDE) / LONGITUDE_RANGE

	latInt := uint32(normalizedLatitude)
	lonInt := uint32(normalizedLongitude)

	return interleave(latInt, lonInt)
}

func compactInt64ToInt32(v uint64) uint32 {
	result := v
	result = (result | (result >> 1)) & 0x3333333333333333
	result = (result | (result >> 2)) & 0x0F0F0F0F0F0F0F0F
	result = (result | (result >> 4)) & 0x00FF00FF00FF00FF
	result = (result | (result >> 8)) & 0x0000FFFF0000FFFF
	result = (result | (result >> 16)) & 0x00000000FFFFFFFF
	return uint32(result)
}

func convertGridNumbersToCoordinates(gridLatitudeNumber, gridLongitudeNumber uint32) Coordinates {
	gridLatitudeMin := MIN_LATITUDE + LATITUDE_RANGE*(float64(gridLatitudeNumber)/math.Pow(2, 26))
	gridLatitudeMax := MIN_LATITUDE + LATITUDE_RANGE*(float64(gridLatitudeNumber+1)/math.Pow(2, 26))
	gridLongitudeMin := MIN_LONGITUDE + LONGITUDE_RANGE*(float64(gridLongitudeNumber)/math.Pow(2, 26))
	gridLongitudeMax := MIN_LONGITUDE + LONGITUDE_RANGE*(float64(gridLongitudeNumber+1)/math.Pow(2, 26))

	latitude := (gridLatitudeMin + gridLatitudeMax) / 2
	longitude := (gridLongitudeMin + gridLongitudeMax) / 2

	return Coordinates{Latitude: latitude, Longitude: longitude}
}

func decode(geoCode uint64) Coordinates {
	x := geoCode & 0x5555555555555555
	y := (geoCode >> 1) & 0x5555555555555555

	gridLatitudeNumber := compactInt64ToInt32(x)
	gridLongitudeNumber := compactInt64ToInt32(y)

	return convertGridNumbersToCoordinates(gridLatitudeNumber, gridLongitudeNumber)
}

type Coordinates struct {
	Longitude float64
	Latitude  float64
}

func inRange(long float64, lat float64) bool {
	return (long >= -180 && long <= 180) && (lat >= -85.05112878 && lat <= 85.05112878)
}

func (rs *RedisStore) Geoadd(key string, long float64, lat float64, member string) (int, error) {
	val, ok := rs.Look(key)
	if !ok {
		val = NewValue(ZSET, 0)
		rs.Store[key] = val
	}
	zset, err := As[Zset](val)
	if err != nil {
		return 0, err
	}

	if !inRange(long, lat) {
		return 0, fmt.Errorf("invalid longitude,latitude pair %f, %f", long, lat)
	}

	score := encode(long, lat)
	fmt.Println(score, float64(score))
	inserts := 0
	if zset.insert(Z{member: member, score: float64(score)}) {
		inserts++
	}

	rs.TouchWatchedKey(key)
	return inserts, nil
}

func (rs *RedisStore) GeoPos(key string, members []string) ([]*Coordinates, error) {
	val, ok := rs.Look(key)
	if !ok {
		return nil, nil
	}
	zset, err := As[Zset](val)
	if err != nil {
		return nil, err
	}

	answers := make([]*Coordinates, len(members))

	for _, member := range members {
		z, ok := zset.dict[member]
		if ok {
			geoCode := uint64(z.score)
			cords := decode(geoCode)
			if inRange(cords.Longitude, cords.Latitude) {
				answers = append(answers, &cords)
			} else {
				answers = append(answers, &Coordinates{})
			}
		} else {
			answers = append(answers, nil)
		}
	}

	return answers, nil
}
