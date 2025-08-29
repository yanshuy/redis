package resp

import (
	"bytes"
	"errors"
	"strconv"
)

var CRLF = []byte("\r\n")

// returns (value, consumedBytes, error).
// If data is incomplete, returns (zeroValue, 0, nil).
func Parse(b []byte) (r []DataType, n int, err error) {
	if len(b) == 0 {
		return r, 0, nil
	}

	for n < len(b) {
		d, o, err := R(b[n:])
		if err != nil {
			return r, 0, err
		}
		n += o
		r = append(r, d)
	}

	return r, n, nil
}

func R(b []byte) (d DataType, n int, err error) {
	i := bytes.Index(b, CRLF)
	if i == -1 {
		return d, 0, errors.New("No CRLF terminator")
	}
	next := i + len(CRLF)

	d.Type = b[0]
	switch d.Type {
	case String:
		d.Str = string(b[1:i])

	case Integer:
		num, _ := strconv.ParseInt(string(b[1:i]), 10, 64)
		d.Int = num

	case BulkString:
		l, _ := strconv.Atoi(string(b[1:i]))
		// TODO: max l 512MB
		d.Str = string(b[next : next+l])
		next += l + len(CRLF)

	case Array:
		l, _ := strconv.Atoi(string(b[1:i]))
		d.Arr = make([]DataType, 0, l)

		for range l {
			v, o, err := R(b[next:])
			if err != nil {
				return v, 0, err
			}
			next += o
			d.Arr = append(d.Arr, v)
		}

	default:
		return d, 0, errors.New("unsupported type")
	}

	return d, next, nil
}
