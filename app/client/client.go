package client

import (
	"encoding/hex"
	"fmt"
	"log"
	"net"

	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
)

type Command struct {
	Name string
	Args []string
}

type ClientState struct {
	Conn       net.Conn
	Server     any
	authAsUser *User

	subscriptions Set[string]
	messageChan   chan PubMessage

	InMulti      bool
	QueuedCmds   []Command
	WatchingKeys Set[string]
	CASDirty     bool

	Blocked  bool
	Unblock  chan struct{}
	RespChan chan resp.Data
}

type Client struct {
	*ClientState
	Command Command
}

func NewClient(conn net.Conn) *Client {
	var user *User
	if DefaultUser.flags.Contains("nopass") {
		user = DefaultUser
	}

	state := &ClientState{
		Conn:          conn,
		authAsUser:    user,
		subscriptions: make(Set[string]),
		messageChan:   make(chan PubMessage),
		WatchingKeys:  make(Set[string]),
		RespChan:      make(chan resp.Data, 100),
	}
	c := &Client{ClientState: state}
	go c.ListenMessages()
	return c
}

func (c *Client) WithCommand(cmd Command) *Client {
	return &Client{
		ClientState: c.ClientState,
		Command:     cmd,
	}
}

func (c *Client) Close() error {
	for channel := range c.subscriptions {
		Chans.Unsubscribe(c, channel)
	}
	// close(c.done)
	close(c.RespChan)
	return c.Conn.Close()
}

func (c *Client) WriteLoop() error {
	for resp := range c.RespChan {
		resBytes := resp.ToResponse()
		_, err := c.Conn.Write(resBytes)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) Block() {
	c.Blocked = true
	c.Unblock = make(chan struct{})
}

func (c *Client) UnBlock() {
	c.Blocked = false
	close(c.Unblock)
}

func (c *Client) InSubscribeMode() bool {
	return len(c.subscriptions) > 0
}

func (c *Client) SubscriptionCount() int {
	return len(c.subscriptions)
}

func (c *Client) InWatch() bool {
	return len(c.WatchingKeys) > 0
}

func (c *Client) IsAuthenticated() bool {
	return c.authAsUser != nil
}

var RDB, _ = hex.DecodeString("524544495330303131fa0972656469732d76657205372e322e30fa0a72656469732d62697473c040fa056374696d65c26d08bc65fa08757365642d6d656dc2b0c41000fa08616f662d62617365c000fff06e3bfec0ff5aa2")

func (c *Client) SendRDB() {
	buf := fmt.Appendf(nil, "$%d\r\n", len(RDB))
	buf = append(buf, RDB...)

	_, err := c.Conn.Write(buf)
	if err != nil {
		log.Fatal("error sending RDB to slave", err.Error())
	}
}
