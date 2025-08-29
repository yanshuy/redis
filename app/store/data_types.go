package store

import (
	"sync"
	"time"
)

type DataType int

const (
	String DataType = iota + 1
	List
	Stream
)

type Data struct {
	Type   DataType
	String string
	Stream *StreamObj
	List   []string
}

type RedisStore struct {
	Store     map[string]*StoreMember
	Config    map[string]string
	Listeners map[string][]chan struct{}
	mu        sync.Mutex
}

func (rs *RedisStore) Look(key string) (*StoreMember, bool) {
	m, ok := rs.Store[key]
	return m, ok
}

type StoreMember struct {
	ExpiryAt time.Time
	data     Data
}

func (rs *RedisStore) NewStoreMember(key string, t DataType) *StoreMember {
	m := &StoreMember{
		data: Data{Type: t},
	}
	rs.Store[key] = m
	return m
}

// func (m *StoreMember) AssignValue(val any) (err error) {
// 	defer func() {
// 		if err != nil {
// 			log.Fatal("ERROR during store member value assignment", err)
// 		}
// 	}()
// 	switch m.data.Type {
// 	case String:
// 		s, ok := val.(string)
// 		if !ok {
// 			return fmt.Errorf("expected string for String type, got %T", val)
// 		}
// 		m.data.String = s
// 	case List:
// 		switch v := val.(type) {
// 		case []string:
// 			m.data.List = append(m.data.List, v...)
// 		default:
// 			return fmt.Errorf("expected []string (or []any) for List type, got %T", val)
// 		}
// 	default:
// 		return fmt.Errorf("unknown data type: %d", m.data.Type)
// 	}
// 	return nil
// }
