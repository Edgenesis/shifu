package client

import (
	"os"

	"github.com/nats-io/nats.go"
)

type Client struct {
	natsClient *nats.Conn
}

func New() (*Client, error) {
	natsClient, err := nats.Connect(os.Getenv("NATS_URL"))
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
