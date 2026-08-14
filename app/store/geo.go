package store

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

	geo.dict[member] = long_lat{
		long: long,
		lat:  lat,
	}

	rs.TouchWatchedKey(key)
	return geo.len(), nil
}
