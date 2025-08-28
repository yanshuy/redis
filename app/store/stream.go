package store

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func (rs *RedisStore) Xadd(key, stream_key string, key_vals []string) (s string, err error) {
	m, exists := rs.Look(key)

	var time_ms int64
	var sqNo int

	if stream_key == "*" {
		time_ms = time.Now().Unix() * 1000
	} else {
		parts := strings.Split(stream_key, "-")
		if len(parts) != 2 {
			return "", errors.New("invalid stream key")
		}
		time_ms, err = strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return "", errors.New("invalid stream key")
		}
		if parts[1] == "*" {
			if exists {
				for _, item := range m.data.Stream {
					if item.time_ms == time_ms {
						sqNo = item.sequenceNo + 1
					}
				}
			}
		} else {
			sqNo, err = strconv.Atoi(parts[1])
			if err != nil {
				return "", errors.New("invalid stream key")
			}
		}
	}

	if time_ms < 0 || sqNo < 0 {
		return "", errors.New("The ID specified in XADD must be greater than 0-0")
	}
	if exists {
		if time_ms == 0 && sqNo == 0 { // chutiya codecrafters case
			return "", errors.New("The ID specified in XADD must be greater than 0-0")
		}
		last := m.data.Stream[len(m.data.Stream)-1]
		if time_ms < last.time_ms || (time_ms == last.time_ms && sqNo <= last.sequenceNo) {
			return "", errors.New("The ID specified in XADD is equal or smaller than the target stream top item")
		}
	}
	if time_ms == 0 {
		sqNo = 1
	}

	entry := StreamEntry{
		sequenceNo: sqNo,
		time_ms:    time_ms,
	}

	if exists {
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

	return fmt.Sprintf("%d-%d", time_ms, sqNo), nil
}
