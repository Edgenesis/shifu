package client

import (
	"github.com/nats-io/nats.go"
)

type Client struct {
	natsClient *nats.Conn
}

func New(address string) (*Client, error) {
	natsClient, err := nats.Connect(address)
	if err != nil {
		return nil, err
	}

	return &Client{
		natsClient: natsClient,
	}, nil
}

func (c *Client) Publish(topic string, message []byte) error {
	err := c.natsClient.Publish(topic, message)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Subscribe(topic string, callback nats.MsgHandler) error {
	_, err := c.natsClient.Subscribe(topic, callback)
	if err != nil {
		return err
	}

	return nil
}
