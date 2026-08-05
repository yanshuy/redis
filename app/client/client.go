package client

import (
	"io"

	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
)

type Command struct {
	Name string
	Args []string
}

type Client struct {
	Conn       io.ReadWriteCloser
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

func NewClient(conn io.ReadWriteCloser) *Client {
	var user *User
	if DefaultUser.flags.Contains("nopass") {
		user = DefaultUser
	}

	c := &Client{
		Conn:          conn,
		authAsUser:    user,
		subscriptions: make(Set[string]),
		WatchingKeys:  make(Set[string]),
		messageChan:   make(chan PubMessage, 100),
		RespChan:      make(chan resp.Data, 100),
	}
	go c.ListenMessages()
	return c
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
	c.Unblock <- struct{}{}
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
