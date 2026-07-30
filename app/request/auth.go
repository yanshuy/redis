package request

import resp "github.com/codecrafters-io/redis-starter-go/app/RESP"

func ACL_WHOAMI() resp.DataType {
	return resp.NewData(resp.BulkString, "default")
}
