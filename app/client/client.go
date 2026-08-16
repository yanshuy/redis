package client

import (
	"context"
	"fmt"
	"net"

	resp "github.com/codecrafters-io/redis-starter-go/app/Resp"
)

type Command struct {
	Name string
	Args []string
}

type Client struct {
	Conn       net.Conn
	authAsUser *User
	Role       Role

	Command Command

	subscriptions Set[string]
	messageChan   chan PubMessage

	InMulti    bool
	QueuedCmds []Command
	CASDirty   bool

	WatchingKeys Set[string]

	Blop struct {
		Cancel      context.CancelFunc
		Reploffset  int
		Numreplicas int
	}

	Blocked  bool
	Unblock  chan struct{}
	respChan chan resp.Data
	Reader   *Reader
}

func NewClient(conn net.Conn, role Role) *Client {
	var user *User
	if DefaultUser.flags.Contains("nopass") {
		user = DefaultUser
	}

	c := &Client{
		Conn:          conn,
		Role:          role,
		authAsUser:    user,
		subscriptions: make(Set[string]),
		messageChan:   make(chan PubMessage),
		WatchingKeys:  make(Set[string]),
		respChan:      make(chan resp.Data, 100),
		Reader:        NewReader(),
	}

	go c.ListenMessages()
	go c.WriteLoop()
	return c
}

func (c *Client) MakeSlave() {
	c.Role = SLAVE
	close(c.messageChan)
}

func (c *Client) Close() error {
	for channel := range c.subscriptions {
		Chans.Unsubscribe(c, channel)
	}
	close(c.respChan)
	return c.Conn.Close()
}

func (c *Client) QueueMessage(message resp.Data) {
	if !message.Is(resp.Empty) {
		c.respChan <- message
	}
}

func (c *Client) WriteLoop() error {
	for resp := range c.respChan {
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

type Role int

func (r Role) String() string {
	switch r {
	case CLIENT:
		return "client"
	case MASTER:
		return "master"
	case SLAVE:
		return "slave"
	default:
		panic(fmt.Sprintf("unexpected client.Role: %#v", r))
	}
}

const (
	CLIENT Role = iota
	SLAVE
	MASTER
)

func (cmd Command) ToRESP() resp.Data {
	str := append([]string{cmd.Name}, cmd.Args...)
	return resp.NewData(resp.Array, str)
}
