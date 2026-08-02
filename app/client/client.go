package client

import (
	"io"

	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
)

type Command struct {
	Name string
	Args []resp.Data
}

type Client struct {
	conn          io.WriteCloser
	subscriptions map[string]chan string
	authAsUser    *User
	InMulti       bool
	QueuedCmds    []Command
	WatchKeys     []string
	CASDirty      bool
	done          chan struct{}
	Quit          bool
}

func (c *Client) InSubscribeMode() bool {
	return len(c.subscriptions) > 0
}

func (c *Client) SubscriptionCount() int {
	return len(c.subscriptions)
}

func (c *Client) InWatch() bool {
	return len(c.WatchKeys) > 0
}

func (c *Client) Write(b []byte) (int, error) {
	return c.conn.Write(b)
}

func (c *Client) Close() error {
	for channel := range c.subscriptions {
		Chans.Unsubscribe(channel, c)
	}
	return c.conn.Close()
}

func (c *Client) IsAuthenticated() bool {
	return c.authAsUser != nil
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
