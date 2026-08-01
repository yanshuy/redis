package request

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"

	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
)

type User struct {
	name      string
	flags     Set
	passwords Set
}

func (u *User) ToRESP() resp.DataType {
	flags := resp.NewData(resp.BulkString, "flags")
	flags_v := resp.NewData(resp.Array, u.flags.ToSlice())

	passwords := resp.NewData(resp.BulkString, "passwords")
	passwords_v := resp.NewData(resp.Array, u.passwords.ToSlice())

	res := []resp.DataType{flags, flags_v, passwords, passwords_v}
	return resp.NewData(resp.Array, res)
}

var users = map[string]*User{
	"default": {
		name:      "default",
		flags:     NewSet("nopass"),
		passwords: Set{},
	},
}

var currentUser = "default"

func ACL_WHOAMI() resp.DataType {
	return resp.NewData(resp.BulkString, currentUser)
}

func ACL_GETUSER(username string) resp.DataType {
	if user, ok := users[username]; ok {
		return user.ToRESP()
	} else {
		panic("user does not exists")
	}
}

func ACL_SETUSER(username string, rule string) resp.DataType {
	if user, ok := users[username]; ok {
		if strings.HasPrefix(rule, ">") {
			pass := []byte(rule[1:])
			hash := sha256.Sum256(pass)

			encoded := hex.EncodeToString(hash[:])
			user.passwords.Add(encoded)

			user.flags.Remove("nopass")
		}
		return resp.NewData(resp.String, "OK")
	} else {
		panic("user does not exists")
	}
}

type Set map[string]struct{}

func NewSet(items ...string) Set {
	s := make(Set)
	s.Add(items...)
	return s
}

func (s Set) Add(items ...string) {
	for _, item := range items {
		s[item] = struct{}{}
	}
}

func (s Set) Remove(items ...string) {
	for _, item := range items {
		delete(s, item)
	}
}

func (s Set) Contains(item string) bool {
	_, ok := s[item]
	return ok
}

func (s Set) ToSlice() []string {
	result := make([]string, 0, len(s))
	for item := range s {
		result = append(result, item)
	}
	return result
}
