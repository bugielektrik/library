package sms

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"
)

type Credentials struct {
	URL      string
	Login    string
	Password string
}

type Client struct {
	httpClient  *http.Client
	credentials Credentials
}

func New(credentials Credentials) Client {
	httpClient := http.DefaultClient
	httpClient.Timeout = 30 * time.Second

	return Client{
		httpClient:  httpClient,
		credentials: credentials,
	}
}

func (c *Client) request(ctx context.Context, method, url string, body io.Reader, headers map[string]string, out interface{}) (err error) {
	// setup http request
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return
	}

	// setup request header
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	// send http request
	res, err := c.httpClient.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	// read response body
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return
	}

	// check error status
	if res.StatusCode != http.StatusOK {
		return errors.New(string(data))
	}
	err = json.Unmarshal(data, &out)

	return
}
