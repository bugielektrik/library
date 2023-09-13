package epay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

type PaymentCardID struct {
	ID string `json:"id"`
}

type PaymentRequest struct {
	Amount          string        `json:"amount"`
	Currency        string        `json:"currency"`
	Name            string        `json:"name"`
	TerminalID      string        `json:"terminalId"`
	InvoiceID       string        `json:"invoiceId"`
	InvoiceIDAlt    string        `json:"invoiceIdAlt"`
	Description     string        `json:"description"`
	AccountID       string        `json:"accountId"`
	Email           string        `json:"email"`
	Phone           string        `json:"phone"`
	BackLink        string        `json:"backLink"`
	FailureBackLink string        `json:"failureBackLink"`
	PostLink        string        `json:"postLink"`
	FailurePostLink string        `json:"failurePostLink"`
	Language        string        `json:"language"`
	PaymentType     string        `json:"paymentType"`
	CardID          PaymentCardID `json:"cardId"`

	HomebankToken  string `json:"-"`
	PaymentPageURL string `json:"-"`

	Token  TokenResponse  `json:"-"`
	Status StatusResponse `json:"-"`
}

type PaymentResponse struct {
	ID           string      `json:"id,omitempty"`
	AccountID    string      `json:"accountId,omitempty"`
	Amount       int         `json:"amount,omitempty"`
	AmountBonus  int         `json:"amountBonus,omitempty"`
	Currency     string      `json:"currency,omitempty"`
	Description  string      `json:"description,omitempty"`
	Email        string      `json:"email,omitempty"`
	InvoiceID    string      `json:"invoiceID,omitempty"`
	Language     string      `json:"language,omitempty"`
	Phone        string      `json:"phone,omitempty"`
	Reference    string      `json:"reference,omitempty"`
	IntReference string      `json:"intReference,omitempty"`
	Secure3D     interface{} `json:"secure3D,omitempty"`
	CardID       string      `json:"cardID,omitempty"`
	PaymentLink  string      `json:"paymentLink,omitempty"`
}

func (c *Client) PayByPaymentPage(ctx context.Context, w http.ResponseWriter, src PaymentRequest, dueDate time.Time) (err error) {
	rootDir, err := os.Getwd()
	if err != nil {
		return
	}

	if !dueDate.IsZero() && (time.Now().Unix() > dueDate.Add(-1500*time.Second).Unix()) {
		src.Status.Transaction.StatusName = "EXPIRED"
		src.Status.Transaction.StatusDescription = "Истек срок оплаты"
	}

	templateName := ""
	switch src.Status.Transaction.StatusName {
	case "NEW", "AUTH", "EXPIRED":
		templateName = filepath.Join(rootDir, "pkg", "provider", "epay", "template", "pending.html")
	case "CHARGE":
		templateName = filepath.Join(rootDir, "pkg", "provider", "epay", "template", "success.html")
	case "CANCEL", "REFUND":
		templateName = filepath.Join(rootDir, "pkg", "provider", "epay", "template", "cancelled.html")
	case "REJECT", "FAILED", "3D", "CANCEL_OLD":
		templateName = filepath.Join(rootDir, "pkg", "provider", "epay", "template", "failed.html")
	default:
		templateName = filepath.Join(rootDir, "pkg", "provider", "epay", "template", "payment.html")

		src.Token, err = c.GetPaymentToken(ctx, &src)
		if err != nil {
			return
		}
		src.PaymentPageURL = c.credentials.PaymentPageURL
	}

	tmpl, err := template.ParseFiles(templateName)
	if err != nil {
		return
	}

	return tmpl.Execute(w, src)
}

func (c *Client) PayBySavedCard(ctx context.Context, src PaymentRequest) (dst PaymentResponse, err error) {
	path, err := url.Parse(c.credentials.URL)
	if err != nil {
		return
	}
	path = path.JoinPath("/payments/cards/auth")

	payload, err := json.Marshal(src)
	if err != nil {
		return
	}
	body := bytes.NewReader(payload)

	token, err := c.GetPaymentToken(ctx, &src)
	if err != nil {
		return
	}

	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %s", token.AccessToken),
	}
	err = c.request(ctx, true, "POST", path.String(), body, headers, &dst)

	return
}
