package resp

import "strconv"

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

func (d *DataType) String() string {
	switch d.Type {
	case Error:
		return d.Str
	case String, BulkString:
		return d.Str
	case Integer:
		return strconv.Itoa(int(d.Int))
	case Array:
		str := ""
		for _, sd := range d.Arr {
			str += sd.String()
		}
		return str
	default:
		return ""
	}
}
