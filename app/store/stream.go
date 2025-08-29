package store

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type StreamID struct {
	MS  int64
	Seq int
}

type StreamEntry struct {
	Id     StreamID
	Fields []string
}

type StreamObj struct {
	LastID  StreamID
	Entries []StreamEntry
}

func (rs *RedisStore) Xadd(key, stream_key string, key_vals []string) (s string, err error) {
	m, exists := rs.Look(key)

	var time_ms int64
	var seqNo int

	if stream_key == "*" {
		time_ms = time.Now().Unix() * 1000
	} else {
		parts := strings.Split(stream_key, "-")
		if len(parts) != 2 {
			return "", errors.New("invalid stream key")
		}
		time_ms, err = strconv.ParseInt(parts[0], 10, 64)
		if err != nil || time_ms < 0 {
			return "", errors.New("invalid stream key")
		}
		if parts[1] == "*" {
			if exists && m.data.Stream.LastID.MS == time_ms {
				seqNo = m.data.Stream.LastID.Seq + 1
			} else if time_ms == 0 {
				seqNo = 1
			}
		} else {
			seqNo, err = strconv.Atoi(parts[1])
			if err != nil || seqNo < 0 {
				return "", errors.New("invalid stream key")
			}
		}
	}

	if time_ms == 0 && seqNo == 0 {
		return "", errors.New("The ID specified in XADD must be greater than 0-0")
	}

	streamId := StreamID{time_ms, seqNo}
	if !exists {
		m := rs.NewStoreMember(key, Stream)
		s := &StreamObj{
			LastID:  streamId,
			Entries: []StreamEntry{{streamId, key_vals}},
		}
		m.data.Stream = s
		return fmt.Sprintf("%d-%d", time_ms, seqNo), nil
	}

	stream := m.data.Stream
	if time_ms < stream.LastID.MS || (time_ms == stream.LastID.MS && seqNo <= stream.LastID.Seq) {
		return "", errors.New("The ID specified in XADD is equal or smaller than the target stream top item")
	}
	stream.LastID = streamId
	stream.Entries = append(stream.Entries, StreamEntry{streamId, key_vals})
	return fmt.Sprintf("%d-%d", time_ms, seqNo), nil
}

func (rs *RedisStore) XRange(key string, startStr string, endStr string) ([]StreamEntry, error) {
	m, exists := rs.Look(key)
	if !exists {
		return nil, nil
	}
	stream := m.data.Stream

	startStr = strings.TrimSpace(startStr)
	endStr = strings.TrimSpace(endStr)
	var start, end int64
	var startSeq, endSeq int
	var err error

	startParts := strings.Split(startStr, "-")
	endParts := strings.Split(endStr, "-")
	fmt.Println(startParts)
	if len(endParts) > 2 || len(startParts) > 2 {
		return nil, errors.New("invalid arguments")
	}

	if startStr == "-" {
		start = stream.Entries[0].Id.MS
		startParts = nil
	} else {
		start, err = strconv.ParseInt(startParts[0], 10, 64)
		if err != nil {
			return nil, errors.New("invalid arguments")
		}
	}

	if endStr == "+" {
		end = stream.Entries[len(stream.Entries)-1].Id.MS
		endParts = nil
	} else {
		end, err = strconv.ParseInt(endParts[0], 10, 64)
		if err != nil {
			return nil, errors.New("invalid arguments")
		}
	}

	if start > end {
		return nil, nil
	}

	var entries []StreamEntry
	for _, entry := range stream.Entries {
		if entry.Id.MS >= start && entry.Id.MS <= end {
			entries = append(entries, entry)
		}
	}

	lo := 0
	if len(startParts) == 2 {
		startSeq, err = strconv.Atoi(startParts[1])
		if err != nil {
			return nil, errors.New("invalid arguments")
		}
		for lo < len(entries) {
			e := entries[lo]
			if e.Id.MS == start && e.Id.Seq < startSeq {
				lo++
			} else if e.Id.MS < start {
				lo++
			} else {
				break
			}
		}
		entries = entries[lo:]
	}

	if len(endParts) == 2 {
		endSeq, err = strconv.Atoi(endParts[1])
		if err != nil {
			return nil, errors.New("invalid arguments")
		}
		hi := len(entries) - 1
		for hi >= lo {
			e := entries[hi]
			if e.Id.MS == end && e.Id.Seq > endSeq {
				hi--
			} else if e.Id.MS > end {
				hi--
			} else {
				break
			}
		}
		entries = entries[:hi+1]
	}

	return entries, nil
}
