package request

import (
	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
)

type User struct {
	name      string
	flags     []string
	passwords []string
}

func (u *User) ToRESP() resp.DataType {
	flags := resp.NewData(resp.BulkString, "flags")
	flags_v := resp.NewData(resp.Array, u.flags)

	passwords := resp.NewData(resp.BulkString, "passwords")
	passwords_v := resp.NewData(resp.Array, u.passwords)

	res := []resp.DataType{flags, flags_v, passwords, passwords_v}
	return resp.NewData(resp.Array, res)
}

var users = map[string]User{
	"default": {
		name:      "default",
		flags:     []string{"nopass"},
		passwords: []string{},
	},
}

func ACL_WHOAMI() resp.DataType {
	return resp.NewData(resp.BulkString, "default")
}

func ACL_GETUSER(user resp.DataType) resp.DataType {
	if user.Type != resp.BulkString {
		return resp.NewData(resp.BulkString, "-1")
	}
	username := user.Str

	if user, ok := users[username]; ok {
		return user.ToRESP()
	} else {
		panic("user does not exists")
	}
}
