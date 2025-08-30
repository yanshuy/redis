package store

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"time"
)

const (
	FA = 0xFA // AUX: Auxiliary Field, key-value settings
	FE = 0xFE // database index
	FD = 0xFD // expire time in seconds
	FC = 0xFC // expire time in milliseconds
	FB = 0xFB // hash table sizes
	FF = 0xFF // end of the file

	REDIS_VERSION = "0011"
)

func (rs *RedisStore) GetRDBFile(flag int) *os.File {
	dir := rs.Config["dir"]
	// cwd, _ := os.Getwd()
	// if !filepath.IsAbs(dir) {
	// 	dir = filepath.Join(cwd, dir)
	// }
	err := os.MkdirAll(dir, 0o755)
	if err != nil {
		log.Fatalf("mkdir %s: %v", dir, err)
	}
	filePath := path.Join(dir, rs.Config["dbfilename"])
	fmt.Println(filePath)
	file, err := os.OpenFile(filePath, flag, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	return file
}

func (rs *RedisStore) SaveRDBSnapshot() (err error) {
	file := rs.GetRDBFile(os.O_CREATE | os.O_WRONLY | os.O_TRUNC)
	defer file.Close()

	w := bufio.NewWriter(file)

	//header
	_, err = w.Write([]byte("REDIS")) // Magic Line
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(REDIS_VERSION))
	if err != nil {
		return err
	}

	//meta-data
	err = w.WriteByte(FA)
	if err != nil {
		return err
	}
	err = writeEncodedString(w, "redis-ver")
	if err != nil {
		return err
	}
	err = writeEncodedString(w, "6.0.16")
	if err != nil {
		return err
	}

	// database
	err = w.WriteByte(FE)
	if err != nil {
		return err
	}
	err = w.WriteByte(byte(0))
	if err != nil {
		return err
	}

	//map size
	err = w.WriteByte(FB)
	if err != nil {
		return err
	}
	storeLenBytes, err := encodeLength(len(rs.Store))
	if err != nil {
		return err
	}
	_, err = w.Write(storeLenBytes)
	if err != nil {
		return err
	}
	_, err = w.Write(storeLenBytes)
	if err != nil {
		return err
	}

	//data
	for key, value := range rs.Store {
		nowMs := time.Now().UnixMilli()
		if value.ExpiryAt > 0 && value.ExpiryAt <= nowMs {
			continue
		}
		if value.data.Type != STRING {
			log.Println("ignoring type other than string")
			continue
		}
		if value.ExpiryAt != 0 {
			err = w.WriteByte(FC)
			if err != nil {
				return err
			}
			var buf [8]byte
			binary.LittleEndian.PutUint64(buf[:], uint64(value.ExpiryAt))
			if _, err = w.Write(buf[:]); err != nil {
				return err
			}
		}

		err := w.WriteByte(byte(STRING))
		if err != nil {
			return err
		}
		err = writeEncodedString(w, key)
		if err != nil {
			return err
		}
		err = writeEncodedString(w, value.data.String)
		if err != nil {
			return err
		}

	}
	// EOF
	if err = w.WriteByte(FF); err != nil {
		return err
	}

	err = w.Flush()
	if err != nil {
		return err
	}
	return nil
}

func writeEncodedString(w io.Writer, s string) error {
	lenBytes, err := encodeLength(len(s))
	if err != nil {
		return err
	}
	_, err = w.Write(lenBytes)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(s))
	if err != nil {
		return err
	}
	return nil
}

func encodeLength(n int) ([]byte, error) {
	if n < 0 {
		return nil, fmt.Errorf("encoding length < 0")
	}
	switch {
	case n < 1<<6:
		return []byte{byte(n)}, nil
	case n < 1<<14:
		fb := byte(n>>8 | 0x40)
		sb := byte(n)
		return []byte{fb, sb}, nil
	case n < 1<<32:
		b := make([]byte, 5)
		b[0] = 0x80
		binary.LittleEndian.PutUint32(b[1:], uint32(n))
		return b, nil
	default:
		return nil, fmt.Errorf("encoding length >= 1<<32")
	}
}

func decodeLength(b []byte) (int, int, bool) {
	if len(b) == 0 {
		return 0, 0, false
	}
	f := uint8(b[0])
	switch f >> 6 {
	case 0:
		l := f & 0xFF
		return int(l), 1, true

	case 1:
		var l uint16
		b[0] = f & 0x3f
		_, err := binary.Decode(b, binary.BigEndian, l)
		if err != nil {
			return 0, 0, false
		}
		return int(l), 2, true

	case 2:
		var l uint32
		_, err := binary.Decode(b[1:], binary.BigEndian, l)
		if err != nil {
			return 0, 0, false
		}
		return int(l), 5, true

	default:
		log.Println("special case not supported")
		return 0, 0, false
	}
}

const (
	header   = 0
	metadata = 1
	database = 2
	keyVals  = 3
)

func (rs *RedisStore) RestoreRDBSnapshot() (err error) {
	file := rs.GetRDBFile(os.O_RDONLY | os.O_CREATE)
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		return err
	}
	if fi.Size() == 0 {
		return nil
	}

	state := header
	StateMachine := func(b []byte) (n int, err error) {
		defer recover()

	loop:
		for len(b) > 0 {
			fmt.Println(len(b), state)
			if b[0] == FF {
				n += 1
				return n, nil
			}
			switch state {
			case header:
				i := bytes.IndexByte(b, FA)
				if i == -1 {
					return 0, nil
				}
				headers := string(b[:i])
				if !strings.Contains(headers, "REDIS0011") {
					return i, fmt.Errorf("unsupported format or version")
				}
				b = b[i+1:]
				n += i + 1
				state = metadata

			case metadata:
				i := bytes.IndexByte(b, FE)
				if i == -1 {
					break loop
				}
				metadata := string(b[:i])
				log.Println("read metadata:", "t", metadata, "t")

				b = b[i+1:]
				n += i + 1
				state = database

			case database:
				if len(b) < 2 {
					break loop
				}
				dbIndex := uint8(b[0])
				log.Println("db index is:", dbIndex)

				if b[1] != FB {
					log.Println("no FB")
					break loop
				}
				i := 2

				tl, used, ok := decodeLength(b[i:])
				if !ok {
					break loop
				}
				log.Println("table len:", tl)
				i += used

				etl, used, ok := decodeLength(b[i:])
				if !ok {
					break loop
				}
				log.Println("table len:", etl)
				i += used

				n += i
				b = b[i:]
				state = keyVals

			case keyVals:
				i := 0
				var expiry uint64
				if b[i] == FC {
					log.Println("got FC")
					i += 1
					if len(b) < i+8 {
						break loop
					}
					expiry = binary.LittleEndian.Uint64(b[i : i+8])
					i += 8
				} else if b[i] == FD {
					i += 1
					log.Println("got FD")
					if len(b) < i+8 {
						break loop
					}
					sec := binary.LittleEndian.Uint32(b[i : i+4])
					expiry = uint64(sec) * 1000
					i += 4
				}
				if len(b) < i+1 {
					break loop
				}

				valType := b[i]
				i += 1
				fmt.Println("hi", string(b[i:]))

				switch valType {
				case byte(STRING):
					kLen, used, ok := decodeLength(b[i:])
					if !ok {
						break loop
					}
					i += used
					if len(b) < i+kLen {
						break loop
					}
					keyBytes := b[i : i+kLen]
					i += kLen

					vLen, used, ok := decodeLength(b[i:])
					if !ok {
						break loop
					}
					i += used
					if len(b) < i+vLen {
						break loop
					}
					valBytes := b[i : i+vLen]
					i += vLen

					rs.Store[string(keyBytes)] = &StoreMember{
						data: Data{
							Type:   STRING,
							String: string(valBytes),
						},
						ExpiryAt: int64(expiry),
					}
					fmt.Println(rs.Store[string(keyBytes)])

				default:
					return i, fmt.Errorf("unsupported value type: 0x%02X", valType)
				}
				n += i
				b = b[i:]
			}
		}
		return n, nil
	}

	buf := make([]byte, 4096)
	bufLen := 0
	for {
		n, err := file.Read(buf[bufLen:])
		if n > 0 {
			bufLen += n
			o, err := StateMachine(buf[:bufLen])
			if err != nil {
				return err
			}
			if o > 0 {
				copy(buf, buf[o:n])
				bufLen -= o
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	return nil
}
