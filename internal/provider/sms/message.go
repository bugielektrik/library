package sms

import (
	"context"
	"net/url"
)

func (c *Client) SendMessage(ctx context.Context, phone, message string) (err error) {
	path, err := url.Parse(c.credentials.URL)
	if err != nil {
		return
	}
	path = path.JoinPath("/sys/send.php")

	params := url.Values{
		"login":  []string{c.credentials.Login},
		"psw":    []string{c.credentials.Password},
		"phones": []string{phone},
		"mes":    []string{message},
	}
	path.RawQuery = params.Encode()

	return c.request(ctx, "GET", path.String(), nil, nil, nil)
}
