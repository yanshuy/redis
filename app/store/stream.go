package store

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	resp "github.com/codecrafters-io/redis-starter-go/app/Resp"
	"github.com/codecrafters-io/redis-starter-go/app/client"
)

type StreamID struct {
	MS  int64
	Seq uint64
}

type StreamEntry struct {
	Id     StreamID
	Fields []string
}

type Stream struct {
	LastID  StreamID
	Entries []StreamEntry
}

func streamIDLess(a, b StreamID) bool {
	if a.MS < b.MS {
		return true
	}
	if a.MS == b.MS {
		return a.Seq < b.Seq
	}
	return false
}

func streamIDLessOrEqual(a, b StreamID) bool {
	if a.MS < b.MS {
		return true
	}
	if a.MS == b.MS {
		return a.Seq <= b.Seq
	}
	return false
}

func (rs *RedisStore) Xadd(key, stream_key string, key_vals []string) (string, error) {
	val, ok := rs.Look(key)
	if !ok {
		val = NewValue(STREAM, 0)
		rs.Store[key] = val
	}
	stream, err := As[Stream](val)
	if err != nil {
		return "", err
	}

	hasEntries := len(stream.Entries) > 0

	var time_ms int64
	var seqNo uint64 = 0

	if stream_key == "*" {
		time_ms = time.Now().UnixMilli()
		if hasEntries {
			if time_ms < stream.LastID.MS {
				time_ms = stream.LastID.MS
			}
			if time_ms == stream.LastID.MS {
				seqNo = stream.LastID.Seq + 1
			}
		}
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
			if hasEntries && stream.LastID.MS == time_ms {
				seqNo = stream.LastID.Seq + 1
			} else if time_ms == 0 {
				seqNo = 1
			} else {
				seqNo = 0
			}
		} else {
			seqNo, err = strconv.ParseUint(parts[1], 10, 64)
			if err != nil {
				return "", errors.New("invalid stream key")
			}
		}
	}

	if time_ms == 0 && seqNo == 0 {
		return "", errors.New("The ID specified in XADD must be greater than 0-0")
	}

	if hasEntries {
		if time_ms < stream.LastID.MS || (time_ms == stream.LastID.MS && seqNo <= stream.LastID.Seq) {
			return "", errors.New("The ID specified in XADD is equal or smaller than the target stream top item")
		}
	}

	streamId := StreamID{MS: time_ms, Seq: seqNo}
	stream.LastID = streamId
	entry := StreamEntry{Id: streamId, Fields: key_vals}

	if !rs.NotifyBlockedOnStream(key, entry) {
		stream.Entries = append(stream.Entries, StreamEntry{Id: streamId, Fields: key_vals})
	}
	val.Obj = stream

	rs.TouchWatchedKey(key)
	return fmt.Sprintf("%d-%d", time_ms, seqNo), nil
}

func (rs *RedisStore) NotifyBlockedOnStream(key string, entry StreamEntry) bool {
	blocked := rs.BlockedOnKeys[key]
	if len(blocked) == 0 {
		return false
	}
	var client *client.Client
	for _, c := range blocked {
		if !c.Blop.IsTimedOut() {
			client = c
			break
		}
	}
	if client == nil {
		return false
	}

	id := fmt.Sprintf("%d-%d", entry.Id.MS, entry.Id.Seq)
	fields := resp.NewData(resp.Array, entry.Fields)

	entryData := resp.NewData(resp.Array, id, fields)
	entries := resp.NewData(resp.Array, entryData)
	keyData := resp.NewData(resp.Array, key, entries)

	client.QueueMessage(resp.NewData(resp.Array, keyData))
	client.Blop.Cancel()

	return true
}

func (rs *RedisStore) Xrange(key string, startStr string, endStr string) ([]StreamEntry, error) {
	val, exists := rs.Look(key)
	if !exists {
		return nil, nil
	}
	stream, err := As[Stream](val)
	if err != nil {
		return nil, err
	}

	startID, err := parseStreamIDBound(startStr, false)
	if err != nil {
		return nil, err
	}

	endID, err := parseStreamIDBound(endStr, true)
	if err != nil {
		return nil, err
	}

	var entries []StreamEntry
	for _, entry := range stream.Entries {
		if streamIDLessOrEqual(startID, entry.Id) && streamIDLessOrEqual(entry.Id, endID) {
			entries = append(entries, entry)
		}
	}

	return entries, nil
}

func parseStreamIDBound(s string, isEnd bool) (StreamID, error) {
	s = strings.TrimSpace(s)
	if s == "-" {
		return StreamID{MS: 0, Seq: 0}, nil
	}
	if s == "+" {
		return StreamID{MS: math.MaxInt64, Seq: math.MaxInt}, nil
	}

	parts := strings.Split(s, "-")
	if len(parts) == 1 {
		ms, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return StreamID{}, errors.New("invalid arguments")
		}
		if isEnd {
			return StreamID{MS: ms, Seq: math.MaxInt}, nil
		}
		return StreamID{MS: ms, Seq: 0}, nil
	} else if len(parts) == 2 {
		ms, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return StreamID{}, errors.New("invalid arguments")
		}
		seq, err := strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			return StreamID{}, errors.New("invalid arguments")
		}
		return StreamID{MS: ms, Seq: seq}, nil
	}
	return StreamID{}, errors.New("invalid arguments")
}

func (rs *RedisStore) Xread(key string, startIDStr string) ([]StreamEntry, error) {
	val, ok := rs.Look(key)
	if !ok {
		return nil, nil
	}
	stream, err := As[Stream](val)
	if err != nil {
		return nil, err
	}

	var entries []StreamEntry

	var startID StreamID
	if startIDStr == "$" {
		return entries, nil
	} else {
		var err error
		startID, err = parseStreamIDBound(startIDStr, false)
		if err != nil {
			return nil, err
		}
	}

	for _, entry := range stream.Entries {
		if streamIDLess(startID, entry.Id) {
			entries = append(entries, entry)
		}
	}

	return entries, nil
}
