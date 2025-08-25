package store

import (
	"fmt"
	"log"
	"time"
)

type DataStructType int

const (
	String DataStructType = iota
	List
)

type DataStruct struct {
	Type   DataStructType
	String string
	List   []string
}

type StoreMember struct {
	ExpiryAt time.Time
	data     DataStruct
}

func NewStoreMember(t DataStructType, val any) (*StoreMember, error) {
	m := &StoreMember{
		data: DataStruct{Type: t},
	}

	var err error
	defer log.Println("NewStoreMember:", err)

	switch t {
	case String:
		s, ok := val.(string)
		if !ok {
			err = fmt.Errorf("expected string for String type, got %T", val)
			return nil, err
		}
		m.data.String = s
	case List:
		switch v := val.(type) {
		case []string:
			m.data.List = append(m.data.List, v...)
		default:
			err = fmt.Errorf("expected []string (or []any) for List type, got %T", val)
			return nil, err
		}
	default:
		err = fmt.Errorf("unknown data type: %d", t)
		return nil, err
	}
	return m, nil
}
