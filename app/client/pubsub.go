package client

import (
	"sync"

	resp "github.com/codecrafters-io/redis-starter-go/app/RESP"
)

type Channels struct {
	Channels map[string][]*Client
	mu       sync.Mutex
}

var Chans = Channels{
	Channels: make(map[string][]*Client),
}

func (c *Channels) Subscribers(channel string) int {
	return len(c.Channels[channel])
}

func (c *Client) ListenMessages() {
	for pub := range c.messageChan {
		res := resp.NewData(resp.Array, "message", pub.channel, pub.message)
		c.RespChan <- res
	}
}

func (c *Channels) Subscribe(client *Client, channels ...string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, channel := range channels {
		if _, ok := client.subscriptions[channel]; ok {
			continue
		}
		client.subscriptions.Add(channel)
		c.Channels[channel] = append(c.Channels[channel], client)
	}
}

type PubMessage struct {
	channel string
	message string
}

func (c *Channels) Publish(channel string, message string) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	delivered := 0
	subs := c.Channels[channel]
	for _, cl := range subs {
		cl.messageChan <- PubMessage{
			channel: channel,
			message: message,
		}
		delivered++
	}
	return delivered
}

func (c *Channels) Unsubscribe(client *Client, channel string) {
	if ok := client.subscriptions.Contains(channel); !ok {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	subs := c.Channels[channel]
	subs = filter(subs, client)

	if len(subs) == 0 {
		delete(c.Channels, channel)
	} else {
		c.Channels[channel] = subs
	}

	client.subscriptions.Remove(channel)

	// if len(client.subscriptions) == 0 {
	// 	close(client.messageChan)
	// }
}

func filter[T comparable](array []T, elem T) []T {
	for i, item := range array {
		if elem == item {
			return append(array[:i], array[i+1:]...)
		}
	}
	return array
}
