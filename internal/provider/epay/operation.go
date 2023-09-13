package epay

import (
	"context"
	"fmt"
	"net/url"
)

func (c *Client) Charge(ctx context.Context, token, transactionID, amount string) (err error) {
	path, err := url.Parse(c.credentials.URL)
	if err != nil {
		return
	}
	path = path.JoinPath("/operation", transactionID, "/charge")

	params := url.Values{
		"amount": []string{amount},
	}
	path.RawQuery = params.Encode()

	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", token),
	}

	return c.request(ctx, true, "POST", path.String(), nil, headers, nil)
}

func (c *Client) Cancel(ctx context.Context, token, transactionID string) (err error) {
	path, err := url.Parse(c.credentials.URL)
	if err != nil {
		return
	}
	path = path.JoinPath("/operation", transactionID, "/cancel")

	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", token),
	}

	return c.request(ctx, true, "POST", path.String(), nil, headers, nil)
}
