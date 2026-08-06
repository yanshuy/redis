package store

import (
	"fmt"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/client"
)

type RedisStore struct {
	Store       map[string]*Value
	BlockedKeys map[string][]chan BlockResult
	WatchedKeys map[string][]*client.Client

	dir        string
	dbfilename string
}

type BlockResult struct {
	Key   string
	Value string
}

func InitializeStore(dir string, dbfilename string) (*RedisStore, error) {
	store := &RedisStore{
		Store:       make(map[string]*Value),
		BlockedKeys: make(map[string][]chan BlockResult),
		WatchedKeys: make(map[string][]*client.Client),
		dir:         dir,
		dbfilename:  dbfilename,
	}
	err := RestoreRDBSnapshot(store)
	if err != nil {
		return nil, err
	}
	return store, nil
}

type ObjType int

const (
	STRING ObjType = iota
	LIST
	STREAM
	ZSET
)

type Value struct {
	ExpiryAt int64
	Type     ObjType
	Obj      any
}

func (rs *RedisStore) Type(key string) string {
	if m, ok := rs.Look(key); ok {
		switch m.Type {
		case STRING:
			return "string"
		case LIST:
			return "list"
		case STREAM:
			return "stream"
		case ZSET:
			return "zset"
		}
	}
	return "none"
}

func NewValue(t ObjType, expiryAt int64) *Value {
	var obj any
	switch t {
	case STRING:
		obj = ""
	case LIST:
		obj = []string{}
	case STREAM:
		obj = Stream{}
	case ZSET:
		obj = NewZset()
	default:
		panic(fmt.Sprintf("unexpected store.ObjType: %#v", t))
	}
	return &Value{ExpiryAt: expiryAt, Type: t, Obj: obj}
}

type List = []string
type RedisValue interface {
	~string | List | Stream | Zset
}

func As[T RedisValue](val *Value) (T, error) {
	v, ok := val.Obj.(T)
	if !ok {
		return v, fmt.Errorf("expected %T, got %T", *new(T), val.Obj)
	}
	return v, nil
}

func (rs *RedisStore) Look(key string) (*Value, bool) {
	m, ok := rs.Store[key]
	return m, ok
}

func (rs *RedisStore) RemoveMemberAfter(ttl_ms int64, key string) {
	timer := time.NewTimer(time.Duration(ttl_ms) * time.Millisecond)
	<-timer.C
	delete(rs.Store, key)
}

func (rs *RedisStore) Set(key string, val string, ttl_ms int64) {
	var expiryAt int64
	if ttl_ms > 0 {
		expiryAt = time.Now().UnixMilli() + ttl_ms
		go rs.RemoveMemberAfter(ttl_ms, key)
	}
	obj := NewValue(STRING, expiryAt)
	obj.Obj = val
	rs.Store[key] = obj

	rs.TouchWatchedKey(key)
}

func (rs *RedisStore) Get(key string) string {
	val, ok := rs.Store[key]
	if !ok {
		return ""
	}
	if val.ExpiryAt > 0 && val.ExpiryAt <= time.Now().UnixMilli() {
		delete(rs.Store, key)
		return ""
	}
	str, err := As[string](val)
	if err != nil {
		return ""
	}
	return str
}

func (rs *RedisStore) Keys(pattern string) []string {
	subStr := strings.ReplaceAll(pattern, "*", "")

	ans := make([]string, 0)
	for key := range rs.Store {
		if strings.Contains(key, subStr) {
			ans = append(ans, key)
		}
	}
	return ans
}

func (rs *RedisStore) TouchWatchedKey(key string) {
	for _, c := range rs.WatchedKeys[key] {
		c.CASDirty = true
	}
}
