package store

import (
	"fmt"
)

type long_lat struct {
	long float64
	lat  float64
}

type Geo struct {
	dict map[string]long_lat
}

func (g Geo) len() int {
	return len(g.dict)
}

func NewGeo() Geo {
	return Geo{
		dict: make(map[string]long_lat),
	}
}

func inRange(long float64, lat float64) bool {
	return (long >= -180 && long <= 180) && (lat >= -85.05112878 && lat <= 85.05112878)
}

func (rs *RedisStore) Geoadd(key string, long float64, lat float64, member string) (int, error) {
	val, ok := rs.Look(key)
	if !ok {
		val = NewValue(GEO, 0)
		rs.Store[key] = val
	}
	geo, err := As[Geo](val)
	if err != nil {
		return 0, err
	}

	if !inRange(long, lat) {
		return 0, fmt.Errorf("invalid longitude,latitude pair %f, %f", long, lat)
	}

	geo.dict[member] = long_lat{
		long: long,
		lat:  lat,
	}

	rs.TouchWatchedKey(key)
	return geo.len(), nil
}
