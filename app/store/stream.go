package store

import (
	"errors"
	"strconv"
	"strings"
)

func (rs *RedisStore) Xadd(key, stream_key string, key_vals []string) error {
	parts := strings.Split(stream_key, "-")
	if len(parts) != 2 {
		return errors.New("invalid stream key")
	}
	time_ms, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return errors.New("invalid stream key")
	}
	var sqNo int
	m, exists := rs.Look(key)
	if exists && parts[0] == "*" {
		for _, item := range m.data.Stream {
			if item.time_ms == time_ms {
				sqNo = item.sequenceNo + 1
			}
		}
	}
	sqNo, err = strconv.Atoi(parts[1])
	if err != nil {
		return errors.New("invalid stream key")
	}
	if time_ms == 0 {
		sqNo = 1
	}

	if time_ms < 0 || sqNo < 0 {
		return errors.New("The ID specified in XADD must be greater than 0-0")
	}

	entry := StreamEntry{
		sequenceNo: sqNo,
		time_ms:    time_ms,
	}

	if exists {
		last := m.data.Stream[len(m.data.Stream)-1]
		if time_ms < last.time_ms {
			if time_ms == 0 && sqNo == 0 {
				return errors.New("The ID specified in XADD must be greater than 0-0")
			}
			return errors.New("The ID specified in XADD is equal or smaller than the target stream top item")
		}
		if time_ms == last.time_ms && sqNo <= last.sequenceNo {
			return errors.New("The ID specified in XADD is equal or smaller than the target stream top item")
		}
		m.data.Stream = append(m.data.Stream, entry)
	} else {
		mNew := rs.NewStoreMember(key, Stream)
		mNew.data.Stream = append(mNew.data.Stream, entry)
	}

	if len(key_vals) > 1 {
		entry.fields = make(map[string]string)
		for i := 0; i < len(key_vals); i += 2 {
			key := key_vals[i]
			val := key_vals[i+1]
			entry.fields[key] = val
		}
	}
	return nil
}
