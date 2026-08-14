package resp

import (
	"bytes"
	"errors"
	"strconv"
)

var CRLF = []byte("\r\n")

const MaxBulkStringSize = 512 * 1024 * 1024

// returns (value, consumedBytes, error).
// If data is incomplete, returns (zeroValue, 0, nil).
func Parse(b []byte) (r Data, n int, err error) {
	if len(b) == 0 {
		return r, 0, nil
	}

	d, o, err := R(b[n:])
	if err != nil {
		return d, 0, err
	}
	n += o

	return d, n, nil
}

func R(b []byte) (d Data, n int, err error) {
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
		l, err := strconv.Atoi(string(b[1:i]))
		if err != nil {
			return d, 0, err
		}
		if len(b) < next+l+2 || len(b) > MaxBulkStringSize {
			return d, 0, nil // incomplete
		}

		d.Str = string(b[next : next+l])
		next += l

		if !bytes.Equal(b[next:next+2], CRLF) {
			return d, 0, errors.New("invalid bulk string terminator")
		}
		next += 2

	case Array:
		l, _ := strconv.Atoi(string(b[1:i]))
		d.Arr = make([]Data, 0, l)

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

func ReadBulkLength(b []byte) (len int, n int, err error) {
	i := bytes.Index(b, CRLF)
	if i == -1 {
		return 0, 0, errors.New("No CRLF terminator")
	}
	l, err := strconv.Atoi(string(b[1:i]))
	if err != nil {
		return 0, 0, err
	}
	return l, i + 2, nil
}
