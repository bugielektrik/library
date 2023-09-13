package epay

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/url"
	"time"

	"github.com/shopspring/decimal"
)

type TokenResponse struct {
	Scope        string          `json:"scope"`
	ExpiresIn    decimal.Decimal `json:"expires_in"`
	TokenType    string          `json:"token_type"`
	AccessToken  string          `json:"access_token"`
	RefreshToken string          `json:"refresh_token"`
}

func (c *Client) initGlobalTokenRefresher() (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c.credentials.GlobalToken, err = c.GetPaymentToken(ctx, nil)
	if err != nil {
		return
	}
	ticker := time.NewTicker(time.Duration(c.credentials.GlobalToken.ExpiresIn.IntPart()-60) * time.Second)

	go func() {
		for {
			<-ticker.C
			c.credentials.GlobalToken, err = c.GetPaymentToken(ctx, nil)
		}
	}()

	return
}

func (c *Client) GetPaymentToken(ctx context.Context, src *PaymentRequest) (dst TokenResponse, err error) {
	path, err := url.Parse(c.credentials.OAuthURL)
	if err != nil {
		return
	}
	path = path.JoinPath("/oauth2/token")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	defer writer.Close()

	_ = writer.WriteField("client_id", c.credentials.Login)
	_ = writer.WriteField("client_secret", c.credentials.Password)
	_ = writer.WriteField("grant_type", "client_credentials")
	_ = writer.WriteField("scope", "webapi usermanagement email_send verification statement statistics payment")

	if src != nil {
		_ = writer.WriteField("amount", src.Amount)
		_ = writer.WriteField("currency", src.Currency)
		_ = writer.WriteField("invoiceID", src.InvoiceID)
		_ = writer.WriteField("terminal", src.TerminalID)
	}

	headers := map[string]string{
		"Content-Type": writer.FormDataContentType(),
	}
	err = c.request(ctx, true, "POST", path.String(), body, headers, &dst)

	return
}
