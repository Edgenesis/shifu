package client

import (
	"time"

	"github.com/nats-io/nats.go"
)

type Client struct {
	natsClient *nats.Conn
}

type ClientOption struct {
	Name              string
	Address           string
	EnableReconnect   bool
	MaxReconnectTimes int
	ReconnectWaitSec  int
	TimeoutSec        int
}

func New(option ClientOption) (*Client, error) {
	opts := nats.Options{
		Name:           option.Name,
		Url:            option.Address,
		AllowReconnect: option.EnableReconnect,
		MaxReconnect:   option.MaxReconnectTimes,
		ReconnectWait:  time.Duration(option.ReconnectWaitSec) * time.Second,
		Timeout:        time.Duration(option.TimeoutSec) * time.Second,
	}

	client, err := opts.Connect()
	if err != nil {
		return nil, err
	}

	return &Client{
		natsClient: client,
	}, nil
}

func (c *Client) Publish(topic string, message []byte) error {
	err := c.natsClient.Publish(topic, message)
	if err != nil {
		return err
	}

	if err := c.natsClient.Flush(); err != nil {
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

func (c *Client) Close() {
	c.natsClient.Close()
}
