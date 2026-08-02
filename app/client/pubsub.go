package client

import (
	"sync"

	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
)

type Channels struct {
	Channels map[string][]*Client
	mu       sync.Mutex
}

func (c *Channels) Subscribers(channel string) int {
	return len(c.Channels[channel])
}

func (c *Channels) Subscribe(channel string, client *Client) {
	if client.subscriptions == nil {
		client.subscriptions = make(map[string]chan string)
	}
	if _, ok := client.subscriptions[channel]; ok {
		return
	}
	cn := make(chan string, 1)
	client.subscriptions[channel] = cn

	c.mu.Lock()
	c.Channels[channel] = append(c.Channels[channel], client)
	c.mu.Unlock()

	go func() {
		for {
			select {
			case msg, ok := <-cn:
				if !ok {
					return
				}
				res := resp.NewData(resp.Array, "message", channel, msg)
				client.conn.Write(res.ToResponse())
			case <-client.done:
				return
			}
		}
	}()
}

func (c *Channels) Unsubscribe(channel string, client *Client) {
	ch, ok := client.subscriptions[channel]
	if !ok {
		return
	}

	c.mu.Lock()

	subs := c.Channels[channel]
	for i, cl := range subs {
		if cl == client {
			subs = append(subs[:i], subs[i+1:]...)
			break
		}
	}
	if len(subs) == 0 {
		delete(c.Channels, channel)
	} else {
		c.Channels[channel] = subs
	}

	c.mu.Unlock()

	close(ch)
	delete(client.subscriptions, channel)

	if len(client.subscriptions) == 0 {
		client.done <- struct{}{}
	}
}

func (c *Channels) Publish(channel string, message string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	delivered := 0
	subs := c.Channels[channel]
	for _, cl := range subs {
		ch := cl.subscriptions[channel]
		ch <- message
		delivered++
	}
	return delivered
}

var Chans = Channels{
	Channels: make(map[string][]*Client),
}
