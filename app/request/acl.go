package request

import (
	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
)

func ACL_WHOAMI() resp.DataType {
	return resp.NewData(resp.BulkString, "default")
}

func ACL_GETUSER(user resp.DataType) resp.DataType {
	if user.Type != resp.BulkString {
		return resp.NewData(resp.BulkString, "-1")
	}
	flags := resp.NewData(resp.BulkString, "flags")
	flags_v := resp.NewData(resp.Array, []string{})

	res := []resp.DataType{flags, flags_v}
	return resp.NewData(resp.Array, res)
}
