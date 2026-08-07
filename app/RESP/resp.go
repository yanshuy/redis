package resp

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
)

const (
	Empty          byte = 0
	Error          byte = '-'
	String         byte = '+'
	Integer        byte = ':'
	NullBulkString byte = '0'
	BulkString     byte = '$'
	Array          byte = '*'
)

type Data struct {
	Type byte
	Str  string
	Int  int64
	Arr  []Data
}

func (d *Data) Is(dataType byte) bool {
	return d.Type == dataType
}

func NewData(t byte, data ...any) Data {
	d := Data{Type: t}
	if len(data) == 0 {
		return d
	}
	datum := data[0]
	switch t {
	case Error:
		d.Str = datum.(string)
		return d
	case String, BulkString:
		d.Str = datum.(string)
		return d
	case NullBulkString:
		return d
	case Integer:
		switch v := datum.(type) {
		case int64:
			d.Int = v
		case int:
			d.Int = int64(v)
		default:
			panic("expecteed int got something else")
		}
		return d
	case Array:
		for _, datum := range data {
			switch v := datum.(type) {
			case Data:
				d.Arr = append(d.Arr, v)
			case []Data:
				d.Arr = v
			case string:
				s := NewData(BulkString, v)
				d.Arr = append(d.Arr, s)
			case []string:
				if len(v) == 0 {
					d.Arr = []Data{}
					return d
				}
				d.Arr = make([]Data, 0, len(v))
				for _, elem := range v {
					s := NewData(BulkString, elem)
					d.Arr = append(d.Arr, s)
				}
			case int64:
				s := NewData(Integer, v)
				d.Arr = append(d.Arr, s)
			case int:
				s := NewData(Integer, int64(v))
				d.Arr = append(d.Arr, s)
			default:
				panic("unhandled case")
			}
		}
		return d
	default:
		if err, ok := datum.(error); ok {
			d.Type = Error
			d.Str = err.Error()
			return d
		}
		log.Fatal("unknown data type encountered:", data)
		return d
	}
}

func (d *Data) String() string {
	switch d.Type {
	case Error:
		return d.Str
	case String, BulkString:
		return d.Str
	case Integer:
		return strconv.Itoa(int(d.Int))
	case Array:
		var str strings.Builder
		for _, sd := range d.Arr {
			str.WriteString(sd.String())
		}
		return str.String()
	default:
		return ""
	}
}

func (d *Data) Integer() (int64, error) {
	switch d.Type {
	case String, BulkString:
		i, err := strconv.ParseInt(d.Str, 10, 64)
		if err != nil {
			return 0, errors.New("bad integer conversion: " + err.Error())
		}
		return i, nil
	case Integer:
		return d.Int, nil
	default:
		return 0, errors.New("bad integer conversion: data is not of expected type")
	}
}

func (d *Data) ToResponse() []byte {
	crlf := "\r\n"
	switch d.Type {
	case Error:
		res := make([]byte, 0, 1+len(d.Str)+2)
		res = append(res, Error)
		res = fmt.Append(res, d.Str)
		res = fmt.Append(res, crlf)
		return res

	case String:
		res := make([]byte, 0, 1+len(d.Str)+2)
		res = append(res, String)
		res = fmt.Append(res, d.Str+crlf)
		return res

	case NullBulkString, Empty:
		return []byte("$-1\r\n")

	case BulkString:
		first := string(BulkString) + strconv.Itoa(len(d.Str)) + crlf
		res := make([]byte, 0, len(first)+len(d.Str)+2)
		res = fmt.Append(res, first)
		res = fmt.Append(res, d.Str+crlf)
		return res

	case Integer:
		intStr := strconv.FormatInt(d.Int, 10)
		res := make([]byte, 0, 1+len(intStr)+2)
		res = append(res, Integer)
		res = fmt.Append(res, intStr+crlf)
		return res

	case Array:
		if d.Arr == nil {
			return []byte("*-1\r\n")
		}
		n := strconv.Itoa(len(d.Arr))
		res := make([]byte, 0, 1+len(n)+2)
		res = append(res, Array)
		res = append(res, []byte(n)...)
		res = fmt.Append(res, crlf)
		for _, sd := range d.Arr {
			res = append(res, sd.ToResponse()...)
		}
		return res

	default:
		log.Fatal("unknown data type encountered")
		return []byte{}
	}
}

func None() Data {
	return NewData(Empty)
}

func Err(msg string) Data {
	return NewData(Error, "ERR "+msg)
}

func WrongPass() Data {
	return NewData(Error, "WRONGPASS invalid username-password pair or user is disabled.")
}

func NoAuth() Data {
	return NewData(Error, "NOAUTH Authentication required.")
}

func WrongArgs(cmd string) Data {
	return Err("wrong number of arguments for '" + cmd + "' command")
}
