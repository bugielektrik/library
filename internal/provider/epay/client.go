package epay

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

	OAuthURL       string
	PaymentPageURL string
	GlobalToken    TokenResponse
}

type Client struct {
	httpClient  *http.Client
	credentials Credentials
}

func New(credentials Credentials) (client Client, err error) {
	httpClient := http.DefaultClient
	httpClient.Timeout = 30 * time.Second

	client = Client{
		httpClient:  httpClient,
		credentials: credentials,
	}
	err = client.initGlobalTokenRefresher()

	return
}

func (c *Client) request(ctx context.Context, repeat bool, method, url string, body io.Reader, headers map[string]string, out interface{}) (err error) {
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

	// check unauthorized status
	if res.StatusCode == http.StatusUnauthorized && repeat {
		if err = c.initGlobalTokenRefresher(); err != nil {
			return
		}
		return c.request(ctx, false, method, url, body, headers, out)
	}

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
