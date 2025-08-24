package request

import (
	"io"
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
