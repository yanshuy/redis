package resp

import (
	"errors"
	"fmt"
	"log"
	"strconv"
)

const (
	Error      byte = '-'
	String     byte = '+'
	Integer    byte = ':'
	BulkString byte = '$'
	Array      byte = '*'
)

type DataType struct {
	Type byte
	Str  string
	Int  int64
	Arr  []DataType
}

func (d *DataType) Is(dataType byte) bool {
	if dataType == String {
		return d.Type == BulkString
	}
	return d.Type == dataType
}

func NewData(t byte, data any) DataType {
	d := DataType{Type: t}
	switch t {
	case Error:
		d.Str = data.(string)
		return d
	case String, BulkString:
		d.Str = data.(string)
		return d
	case Integer:
		d.Int = data.(int64)
		return d
	case Array:
		elems := data.([]string)
		for _, elem := range elems {
			data := NewData(BulkString, elem)
			d.Arr = append(d.Arr, data)
		}
		return d

	default:
		if err, ok := data.(error); ok {
			d.Str = err.Error()
			return d
		}
		log.Fatal("unknown data type encountered:", string(t))
		return d
	}
}

// func (d *DataType) String() string {
// 	switch d.Type {
// 	case Error:
// 		return d.Str
// 	case String, BulkString:
// 		return d.Str
// 	case Integer:
// 		return strconv.Itoa(int(d.Int))
// 	case Array:
// 		str := ""
// 		for _, sd := range d.Arr {
// 			str += sd.String()
// 		}
// 		return str
// 	default:
// 		return ""
// 	}
// }

func (d *DataType) Integer() (int64, error) {
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

func (d *DataType) ToResponse() []byte {
	crlf := "\r\n"
	switch d.Type {
	case Error:
		res := make([]byte, 0, 1+6+len(d.Str)+2)
		res = append(res, Error)
		res = fmt.Append(res, "ERROR ")
		res = fmt.Append(res, d.Str+crlf)
		return res

	case String:
		res := make([]byte, 0, 1+len(d.Str)+2)
		res = append(res, String)
		res = fmt.Append(res, d.Str+crlf)
		return res

	case BulkString:
		first := string(BulkString) + strconv.Itoa(len(d.Str)) + crlf
		res := make([]byte, 0, len(first)+len(d.Str)+2)
		if d.Str == "" {
			res = append(res, BulkString)
			res = fmt.Append(res, "-1"+crlf)
			return res
		}
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
