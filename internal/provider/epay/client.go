package epay

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"library-service/internal/config"
)

type Credentials struct {
	Username string
	Password string
	Endpoint string
	OAuth    string
	JS       string

	AccessToken string
	ExpiresIn   int64
}

type Token struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    string `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
}

type Client struct {
	client     *http.Client
	configs    *config.Configs
	credential Credentials
}

func New(configs *config.Configs, credential Credentials) *Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	client.Timeout = 30 * time.Second

	return &Client{
		configs:    configs,
		client:     client,
		credential: credential,
	}
}

func (c *Client) InitTokenRefresher() (err error) {
	if err = c.updateToken(); err != nil {
		return
	}

	ticker := time.Duration(c.credential.ExpiresIn - 60)
	timer := time.NewTicker(ticker * time.Second)

	go func() {
		for {
			<-timer.C

			err = c.updateToken()
			if err != nil {
				return
			}
		}
	}()

	return
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
	// preparation of request params
	reqBytes := &bytes.Buffer{}
	writer := multipart.NewWriter(reqBytes)

	_ = writer.WriteField("grant_type", "client_credentials")
	_ = writer.WriteField("scope", "webapi usermanagement email_send verification statement statistics payment")
	_ = writer.WriteField("client_id", c.credential.Username)
	_ = writer.WriteField("client_secret", c.credential.Password)

	_ = writer.WriteField("invoiceID", requestSrc.InvoiceID)
	_ = writer.WriteField("amount", fmt.Sprint(requestSrc.Amount))
	_ = writer.WriteField("currency", requestSrc.Currency)
	_ = writer.WriteField("terminal", requestSrc.TerminalID)

	if err = writer.Close(); err != nil {
		return
	}

	// setup request handler
	path := c.credential.OAuth + "/oauth2/token"
	resBytes, status, err := c.handler("POST", path, writer.FormDataContentType(), "", reqBytes, true)
	if err != nil {
		return
	}

	// check response status
	switch status {
	case http.StatusOK:
		err = json.Unmarshal(resBytes, &responseSrc)
	default:
		err = errors.New(string(resBytes))
	}

	return
}

func (c *Client) handler(method, path, contentType, accessToken string, reqBytes io.Reader, repeat bool) (resBytes []byte, status int, err error) {
	// setup request
	req, err := http.NewRequest(method, path, reqBytes)
	if err != nil {
		return
	}

	// setup request header
	if contentType == "" {
		contentType = "application/json"
	}
	req.Header.Add("Content-Type", contentType)

	if accessToken == "" {
		req.Header.Add("Authorization", "Bearer "+c.credential.AccessToken)
	} else {
		req.Header.Add("Authorization", "Bearer "+accessToken)
	}

	// send request
	res, err := c.client.Do(req)
	if err != nil {
		return
	}
	status = res.StatusCode

	// read response body
	defer func() {
		if err = res.Body.Close(); err != nil {
			return
		}
	}()

	if res.StatusCode == 400 || res.StatusCode == 401 {
		if accessToken == "" {
			var token Token
			if token, err = c.getToken(Request{}); err != nil {
				return
			}

			c.credential.AccessToken = token.AccessToken
			if repeat {
				resBytes, status, err = c.handler(method, path, contentType, accessToken, reqBytes, false)
			}
		}
	}

	resBytes, err = io.ReadAll(res.Body)
	if err != nil {
		return
	}

	return
}
