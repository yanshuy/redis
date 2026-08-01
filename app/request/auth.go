package request

import (
	"crypto/sha256"
	"fmt"
	"strings"

	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
)

type User struct {
	name      string
	flags     Set[string]
	passwords Set[[32]byte]
}

func (u *User) ToRESP() resp.Data {
	flags := resp.NewData(resp.BulkString, "flags")
	flags_v := resp.NewData(resp.Array, u.flags.ToSlice())

	passwords := resp.NewData(resp.BulkString, "passwords")
	passStrs := make([]string, 0, len(u.passwords))
	for p := range u.passwords {
		passStrs = append(passStrs, fmt.Sprintf("%x", p))
	}
	passwords_v := resp.NewData(resp.Array, passStrs)

	res := []resp.Data{flags, flags_v, passwords, passwords_v}
	return resp.NewData(resp.Array, res)
}

var DefaultUser = &User{
	name:      "default",
	flags:     NewSet[string]("nopass"),
	passwords: Set[[32]byte]{},
}

var users = map[string]*User{
	"default": DefaultUser,
}

var currentUser = "default"

func ACL_WHOAMI() resp.Data {
	return resp.NewData(resp.BulkString, currentUser)
}

func ACL_GETUSER(username string) resp.Data {
	if user, ok := users[username]; ok {
		return user.ToRESP()
	} else {
		return Err("User does not exist.")
	}
}

func ACL_SETUSER(username string, rule string) resp.Data {
	if user, ok := users[username]; ok {
		if strings.HasPrefix(rule, ">") {
			pass := []byte(rule[1:])
			hash := sha256.Sum256(pass)
			user.passwords.Add(hash)

			user.flags.Remove("nopass")
		}
		return resp.NewData(resp.String, "OK")
	} else {
		return Err("User does not exist.")
	}
}

func Authenticate(username string, password string) resp.Data {
	if user, ok := users[username]; ok {
		calc := sha256.Sum256([]byte(password))

		if user.passwords.Contains(calc) {
			return resp.NewData(resp.String, "OK")
		}
		return WrongPass()

	} else {
		return Err("User does not exist.")
	}
}

type Set[T comparable] map[T]struct{}

func NewSet[T comparable](items ...T) Set[T] {
	s := make(Set[T])
	s.Add(items...)
	return s
}

func (s Set[T]) Add(items ...T) {
	for _, item := range items {
		s[item] = struct{}{}
	}
}

func (s Set[T]) Remove(items ...T) {
	for _, item := range items {
		delete(s, item)
	}
}

func (s Set[T]) Contains(item T) bool {
	_, ok := s[item]
	return ok
}

func (s Set[T]) ToSlice() []T {
	result := make([]T, 0, len(s))
	for item := range s {
		result = append(result, item)
	}
	return result
}
