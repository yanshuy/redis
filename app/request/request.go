package request

import (
	"io"

	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
)

func ReadAndHandleRequest(conn io.ReadWriter) (n int, err error) {
	// TODO: request > 1024
	b := make([]byte, 1024)
	bLen := 0
	for {
		n, err := conn.Read(b[bLen:])
		if n > 0 {
			bLen += n
			r, o, err := Parse(b[:bLen])
			if err != nil {
				return bLen, err
			}
			if o > 0 {
				err := HandleRequest(conn, r)
				if err != nil {
					return bLen, err
				}
				copy(b, b[o:n])
				bLen -= o
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return bLen, err
		}
	}

	return bLen, nil
}

func HandleRequest(w io.Writer, rs []resp.DataType) (err error) {
	for _, r := range rs {
		var res resp.DataType
		switch r.Type {
		case resp.Array:
			if r.Arr[0].Is(resp.String) {
				res = HandleCmd(r.Arr[0].Str, r.Arr[1:])
			} else {
				res = resp.NewData(resp.Error, "invalid command")
			}
		case resp.String, resp.BulkString:
			res = HandleCmd(r.Str, nil)
		default:
			res = resp.NewData(resp.Error, "invalid command")
		}

		// fmt.Printf("res %+v\n", res)
		resBytes := res.ToResponse()
		_, err := w.Write(resBytes)
		if err != nil {
			return err
		}
	}

	return err
}
