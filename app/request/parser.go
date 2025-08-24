package request

import (
	"bytes"
	"errors"
	"strconv"

	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
)

var CRLF = []byte("\r\n")

// returns (value, consumedBytes, error).
// If data is incomplete, returns (zeroValue, 0, nil).
func Parse(b []byte) (r []resp.DataType, n int, err error) {
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

func R(b []byte) (d resp.DataType, n int, err error) {
	i := bytes.Index(b, CRLF)
	if i == -1 {
		return d, 0, errors.New("No CRLF terminator")
	}
	next := i + len(CRLF)

	d.Type = b[0]
	switch d.Type {
	case resp.String:
		d.Str = string(b[1:i])

	case resp.Integer:
		num, _ := strconv.ParseInt(string(b[1:i]), 10, 64)
		d.Int = num

	case resp.BulkString:
		l, _ := strconv.Atoi(string(b[1:i]))
		// TODO: max l 512MB
		d.Str = string(b[next : next+l])
		next += l + len(CRLF)

	case resp.Array:
		l, _ := strconv.Atoi(string(b[1:i]))
		d.Arr = make([]resp.DataType, 0, l)

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
