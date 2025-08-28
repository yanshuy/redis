package store

var Streams = make([]map[string]string, 0)

func (rs *RedisStore) Xadd(stream_key string, key_vals []string) {
	entry := make(map[string]string)
	entry["stream_key"] = stream_key
	for i := 0; i < len(key_vals); i += 2 {
		key := key_vals[i]
		val := key_vals[i+1]
		entry[key] = val
	}
	Streams = append(Streams, entry)
}
