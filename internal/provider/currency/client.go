package currency

import (
	"context"
	"encoding/xml"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/patrickmn/go-cache"
)

type Credentials struct {
	URL string
}

type Client struct {
	caches      *cache.Cache
	httpClient  *http.Client
	Credentials Credentials
}

func New(credentials Credentials) *Client {
	// Cache with 5 minutes expiration and 10 minutes cleanup interval
	caches := cache.New(5*time.Minute, 10*time.Minute)

	httpClient := http.DefaultClient
	httpClient.Timeout = 30 * time.Second

	client := &Client{
		caches:     caches,
		httpClient: httpClient,

		Credentials: credentials,
	}
	client.initCacheRefresher()

	return client
}

func (c *Client) request(ctx context.Context, method, url string, out interface{}) (err error) {
	// create new request
	request, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return
	}

	// setup request header
	headers := map[string]string{
		"Content-Type": "text/xml",
	}

	for key, value := range headers {
		request.Header.Add(key, value)
	}

	// send request
	res, err := c.httpClient.Do(request)
	if err != nil {
		return
	}
	defer res.Body.Close()

	// read response body
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return
	}

	// check response status
	if res.StatusCode != http.StatusOK {
		return errors.New(string(data))
	}
	err = xml.Unmarshal(data, &out)

	return
}
