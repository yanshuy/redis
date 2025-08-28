package store

import (
	"fmt"
	"log"
	"time"
)

type DataType int

const (
	String DataType = iota + 1
	List
	Stream
)

type DataStruct struct {
	Type   DataType
	String string
	List   []string
}

type StoreMember struct {
	ExpiryAt time.Time
	data     DataStruct
}

func NewStoreMember(t DataType) *StoreMember {
	m := &StoreMember{
		data: DataStruct{Type: t},
	}
	return m
}

func (m *StoreMember) AssignValue(val any) (err error) {
	defer func() {
		if err != nil {
			log.Fatal("ERROR during store member value assignment", err)
		}
	}()
	switch m.data.Type {
	case String:
		s, ok := val.(string)
		if !ok {
			return fmt.Errorf("expected string for String type, got %T", val)
		}
		m.data.String = s
	case List:
		switch v := val.(type) {
		case []string:
			m.data.List = append(m.data.List, v...)
		default:
			return fmt.Errorf("expected []string (or []any) for List type, got %T", val)
		}
	default:
		return fmt.Errorf("unknown data type: %d", m.data.Type)
	}
	return nil
}
