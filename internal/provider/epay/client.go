package epay

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"library-service/config"
	"net/http"
	"strconv"
	"time"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	client     *resty.Client
	configs    config.Configs
	credential Credentials
}

func New(configs config.Configs, credential Credentials) *Client {
	client := resty.New()
	client.SetTimeout(30 * time.Second)
	client.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})

	return &Client{
		configs:    configs,
		client:     client,
		credential: credential,
	}
}

func (c *Client) updateToken() (err error) {
	token, err := c.getToken(Request{})
	if err != nil {
		return
	}

	c.credential.AccessToken = token.AccessToken

	expiresIn, err := strconv.ParseInt(token.ExpiresIn, 10, 64)
	if err != nil {
		c.credential.ExpiresIn = 3600
	} else {
		c.credential.ExpiresIn = expiresIn
	}

	return
}

func (c *Client) getToken(requestSrc Request) (responseSrc Token, err error) {
	path := c.credential.OAuth + "/oauth2/token"

	formData := map[string]string{
		"grant_type":    "client_credentials",
		"scope":         "webapi usermanagement email_send verification statement statistics payment",
		"client_id":     c.credential.Username,
		"client_secret": c.credential.Password,
		"invoiceID":     requestSrc.InvoiceID,
		"amount":        fmt.Sprint(requestSrc.Amount),
		"currency":      requestSrc.Currency,
		"terminal":      requestSrc.TerminalID,
	}

	resBytes, status, err := c.handler("POST", path, "", "", formData, true)
	if err != nil {
		return
	}

	switch status {
	case http.StatusOK:
		err = json.Unmarshal(resBytes, &responseSrc)
	default:
		err = errors.New(string(resBytes))
	}

	return
}

func (c *Client) handler(method, path, contentType, accessToken string, body interface{}, repeat bool) (resBytes []byte, status int, err error) {
	req := c.client.R()

	if formData, ok := body.(map[string]string); ok {
		req.SetFormData(formData)
	} else if body != nil {
		if contentType == "" {
			contentType = "application/json"
		}
		req.SetHeader("Content-Type", contentType)
		req.SetBody(body)
	} else {
		if contentType == "" {
			contentType = "application/json"
		}
		req.SetHeader("Content-Type", contentType)
	}

	token := c.credential.AccessToken
	if accessToken != "" {
		token = accessToken
	}
	if token != "" {
		req.SetHeader("Authorization", "Bearer "+token)
	}

	resp, err := req.Execute(method, path)
	if err != nil {
		return
	}

	status = resp.StatusCode()
	resBytes = resp.Body()

	if (status == 400 || status == 401) && accessToken == "" && repeat {
		var newToken Token
		if newToken, err = c.getToken(Request{}); err != nil {
			return
		}

		c.credential.AccessToken = newToken.AccessToken

		resBytes, status, err = c.handler(method, path, contentType, "", body, false)
	}

	return
}
