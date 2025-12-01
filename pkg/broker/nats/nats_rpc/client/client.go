package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"library-service/pkg/broker/nats/nats_rpc"
	"time"

	"github.com/nats-io/nats.go"
)

const (
	_defaultWaitTime = 5 * time.Second
	_defaultAttempts = 10
	_defaultTimeout  = 2 * time.Second
)

// Client -.
type Client struct {
	subject    string
	connection *nats.Conn

	timeout time.Duration
}

// New -.
func New(
	url string,
	serverSubject string,
	opts ...Option,
) (*Client, error) {
	connection, err := nats.Connect(
		url,
		nats.ReconnectWait(_defaultWaitTime),
		nats.MaxReconnects(_defaultAttempts),
		nats.Timeout(_defaultWaitTime),
	)
	if err != nil {
		return nil, fmt.Errorf("nats_rpc client - NewClient - nats.Connect: %w", err)
	}

	c := &Client{
		subject:    serverSubject,
		connection: connection,
		timeout:    _defaultTimeout,
	}

	// Custom options
	for _, opt := range opts {
		opt(c)
	}

	c.connection = connection

	return c, nil
}

// Shutdown -.
func (c *Client) Shutdown() error {
	c.connection.Close()

	return nil
}

// RemoteCall -.
func (c *Client) RemoteCall(handler string, request, response interface{}) error {
	var (
		requestBody []byte
		err         error
	)

	if request != nil {
		requestBody, err = json.Marshal(request)
		if err != nil {
			return err
		}
	}

	requestMessage := nats.Msg{
		Subject: c.subject,
		Header: nats.Header{
			"Handler": []string{handler},
		},
		Data: requestBody,
	}

	message, err := c.connection.RequestMsg(&requestMessage, c.timeout)
	if errors.Is(err, context.DeadlineExceeded) {
		return nats_rpc.ErrTimeout
	}

	if err != nil {
		return fmt.Errorf("nats_rpc client - Client - RemoteCall - c.connection.Conn.Request: %w", err)
	}

	switch message.Header.Get("Status") {
	case nats_rpc.Success:
		err = json.Unmarshal(message.Data, &response)
		if err != nil {
			return fmt.Errorf("nats_rpc client - Client - RemoteCall - json.Unmarshal: %w", err)
		}
	case nats_rpc.ErrBadHandler.Error():
		return nats_rpc.ErrBadHandler
	case nats_rpc.ErrInternalServer.Error():
		return nats_rpc.ErrInternalServer
	}

	return nil
}
