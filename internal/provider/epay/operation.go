package epay

import (
	"errors"
	"net/http"
	"net/url"
)

func (c *Client) Charge(transactionID, amount string) (err error) {
	// preparation of request params
	params := url.Values{}
	params.Add("amount", amount)

	// setup request handler
	path := c.credential.Endpoint + "/operation/" + transactionID + "/charge?" + params.Encode()
	resBytes, status, err := c.handler("POST", path, "", "", nil, true)
	if err != nil {
		return
	}

	// check response status
	if status != http.StatusOK {
		err = errors.New(string(resBytes))
	}

	return
}

func (c *Client) Cancel(transactionID string) (err error) {
	// setup request handler
	path := c.credential.Endpoint + "/operation/" + transactionID + "/cancel"
	resBytes, status, err := c.handler("POST", path, "", "", nil, true)
	if err != nil {
		return
	}

	// check response status
	if status != http.StatusOK {
		err = errors.New(string(resBytes))
	}

	return
}
