package client

import (
	"crypto/sha256"
	"fmt"
	"io"
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
	fmt.Printf("%+v\n", resp.NewData(resp.Array, res))
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

func NewClient(conn io.WriteCloser) *Client {
	var user *User
	if DefaultUser.flags.Contains("nopass") {
		user = DefaultUser
	}
	doneChan := make(chan struct{})

	return &Client{
		conn:       conn,
		authAsUser: user,
		done:       doneChan,
	}
}

var currentUser = "default"

func ACL_WHOAMI() resp.Data {
	return resp.NewData(resp.BulkString, currentUser)
}

func ACL_GETUSER(username string) resp.Data {
	if user, ok := users[username]; ok {
		return user.ToRESP()
	} else {
		return resp.Err("User does not exist.")
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
		return resp.Err("User does not exist.")
	}
}

func Authenticate(c *Client, username string, password string) resp.Data {
	if user, ok := users[username]; ok {
		calc := sha256.Sum256([]byte(password))

		if user.passwords.Contains(calc) {
			c.authAsUser = user
			return resp.NewData(resp.String, "OK")
		}
		return resp.WrongPass()

	} else {
		return resp.Err("User does not exist.")
	}
}
