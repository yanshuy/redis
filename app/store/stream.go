package store

func (rs *RedisStore) Xadd(key, stream_key string, key_vals []string) {
	entry := make(map[string]string)
	entry["id"] = stream_key
	for i := 0; i < len(key_vals); i += 2 {
		key := key_vals[i]
		val := key_vals[i+1]
		entry[key] = val
	}
	m := NewStoreMember(Stream)
	m.data.Stream = append(m.data.Stream, entry)
	rs.Store[key] = m
}
