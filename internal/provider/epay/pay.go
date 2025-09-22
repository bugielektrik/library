package epay

import (
	"bytes"
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type Request struct {
	CreatedAt       *time.Time     `json:"createdAt,omitempty" db:"created_at"`
	UpdatedAt       *time.Time     `json:"updatedAt,omitempty" db:"updated_at"`
	ID              string         `json:"id,omitempty" db:"id"`
	IIN             string         `json:"iin" db:"iin" validate:"required,len=12"`
	CorrelationID   string         `json:"correlationId" db:"correlation_id" validate:"required,uuid4"`
	InvoiceID       string         `json:"invoiceId" db:"invoice_id" validate:"required"`
	Amount          string         `json:"amount" db:"amount" validate:"required"`
	Currency        string         `json:"currency" db:"currency" validate:"required"`
	TerminalID      string         `json:"terminalId" db:"terminal_id" validate:"required"`
	Description     string         `json:"description" db:"description" validate:"required"`
	AccountID       string         `json:"accountId" db:"account_id"`
	Name            string         `json:"name" db:"name"`
	Email           string         `json:"email" db:"email"`
	Phone           string         `json:"phone" db:"phone" validate:"required"`
	BackLink        string         `json:"backLink" db:"back_link" validate:"required"`
	FailureBackLink string         `json:"failureBackLink" db:"failure_back_link"`
	PostLink        string         `json:"postLink" db:"post_link" validate:"required"`
	FailurePostLink string         `json:"failurePostLink" db:"failure_post_link"`
	Language        string         `json:"language" db:"language"`
	Data            string         `json:"data" db:"data"`
	CardSave        bool           `json:"cardSave" db:"card_save"`
	PaymentType     string         `json:"paymentType" db:"payment_type"`
	CardID          interface{}    `json:"cardId,omitempty" db:"card_id"`
	DueDate         *time.Time     `json:"-"`
	PaymentJsLink   string         `json:"-"`
	HomebankToken   string         `json:"-"`
	Token           Token          `json:"-"`
	Status          StatusResponse `json:"-"`
	ReceiptLink     string         `json:"-" db:"-"`

	RetryLink string
	Retry     bool
}

type RequestByCardID struct {
	ID          string      `json:"id,omitempty" db:"id"`
	IIN         string      `json:"iin" db:"iin" validate:"required,len=12"`
	InvoiceID   string      `json:"invoiceId" db:"invoice_id" validate:"required"`
	Amount      float64     `json:"amount" db:"amount" validate:"required"`
	Currency    string      `json:"currency" db:"currency" validate:"required"`
	TerminalID  string      `json:"terminalId" db:"terminal_id" validate:"required"`
	Description string      `json:"description" db:"description" validate:"required"`
	PaymentType string      `json:"paymentType"`
	AccountID   string      `json:"accountId" db:"account_id"`
	Name        string      `json:"name" db:"name"`
	Email       string      `json:"email" db:"email"`
	Phone       string      `json:"phone" db:"phone" validate:"required"`
	CardID      interface{} `json:"cardId,omitempty" db:"card_id"`
	PostLink    string      `json:"postLink"`
	Data        string      `json:"data" db:"data"`
}

func parseToRequestByCardID(data Request) (res RequestByCardID, err error) {
	amount, err := decimal.NewFromString(data.Amount)
	if err != nil {
		return
	}

	return RequestByCardID{
		ID:          data.ID,
		IIN:         data.IIN,
		InvoiceID:   data.InvoiceID,
		Amount:      amount.InexactFloat64(),
		Currency:    data.Currency,
		TerminalID:  data.TerminalID,
		Description: data.Description,
		AccountID:   data.AccountID,
		Name:        data.Name,
		CardID:      data.CardID,
		PostLink:    data.PostLink,
		PaymentType: "cardId",
		Data:        data.Data,
	}, nil
}

type PaymentCardID struct {
	ID interface{} `json:"id"`
}

type Response struct {
	ID           string      `json:"id,omitempty"`
	AccountID    string      `json:"accountId,omitempty"`
	Amount       float32     `json:"amount,omitempty"`
	AmountBonus  int         `json:"amountBonus,omitempty"`
	Currency     string      `json:"currency,omitempty"`
	Description  string      `json:"description,omitempty"`
	Email        string      `json:"email,omitempty"`
	InvoiceID    string      `json:"invoiceId,omitempty"`
	CardIssuer   string      `json:"issuer"`
	Language     string      `json:"language,omitempty"`
	Phone        string      `json:"phone,omitempty"`
	Reference    string      `json:"reference,omitempty"`
	IntReference string      `json:"intReference,omitempty"`
	Secure3D     interface{} `json:"secure3D,omitempty"`
	CardID       string      `json:"cardId,omitempty"`
	CardMask     string      `json:"cardMask"`
	CardType     string      `json:"cardType"`
	TerminalID   string      `json:"terminal"`
	PaymentLink  string      `json:"paymentLink,omitempty"`
	Status       string      `json:"status"`
	Code         string      `json:"code"`
	ApprovalCode string      `json:"approvalCode"`
}

type ResponseCardID struct {
	ID                string      `json:"id,omitempty"`
	AccountID         string      `json:"accountId,omitempty"`
	Amount            int         `json:"amount,omitempty"`
	AmountBonus       int         `json:"amountBonus,omitempty"`
	Currency          string      `json:"currency,omitempty"`
	Description       string      `json:"description,omitempty"`
	Email             string      `json:"email,omitempty"`
	InvoiceID         string      `json:"invoiceId,omitempty"`
	Issuer            string      `json:"issuer"`
	Language          string      `json:"language,omitempty"`
	Phone             string      `json:"phone,omitempty"`
	Reference         string      `json:"reference,omitempty"`
	IntReference      string      `json:"intReference,omitempty"`
	Secure3D          interface{} `json:"secure3D,omitempty"`
	CardID            string      `json:"cardId,omitempty"`
	CardMask          string      `json:"cardMask"`
	Terminal          string      `json:"terminal"`
	PaymentLink       string      `json:"paymentLink,omitempty"`
	Status            string      `json:"status"`
	Code              int         `json:"code"`
	ApprovalCode      string      `json:"approvalCode"`
	Message           string      `json:"message"`
	StatusID          string      `json:"statusID"`
	StatusDescription string      `json:"statusDescription"`
}

func (c *Client) PayByTemplate(w http.ResponseWriter, requestSrc Request) (err error) {
	// preparation of request params
	rootDir, err := os.Getwd()
	if err != nil {
		return
	}

	if requestSrc.DueDate != nil && requestSrc.Status.Transaction.StatusName == "" {
		localtime := time.Now().Add(6 * time.Hour)
		dueTime := requestSrc.DueDate.Add(-1500 * time.Second)

		if localtime.Unix() > dueTime.Unix() {
			requestSrc.Status.Transaction.StatusName = "EXPIRED"
			requestSrc.Status.Transaction.StatusTitle = "Истек срок оплаты"

			requestSrc.Status.Transaction.Amount, err = decimal.NewFromString(requestSrc.Amount)
			if err != nil {
				return
			}
		}
	}

	filenames := ""
	switch requestSrc.Status.Transaction.StatusName {
	case "NEW", "AUTH", "EXPIRED":
		requestSrc.Status.Transaction.Status = "pending"
		filenames = filepath.Join(rootDir, "templates", "status.html")
	case "CHARGE":
		requestSrc.Status.Transaction.Status = "success"
		filenames = filepath.Join(rootDir, "templates", "status.html")

		if strings.EqualFold(requestSrc.Description, "Привязка банковской карты") {
			filenames = filepath.Join(rootDir, "templates", "redirect.html")
		}
	case "CANCEL", "CANCEL_OLD", "REFUND":
		requestSrc.Status.Transaction.Status = "cancel"
		filenames = filepath.Join(rootDir, "templates", "status.html")
	case "REJECT", "FAILED", "3D":
		requestSrc.Status.Transaction.Status = "failed"
		filenames = filepath.Join(rootDir, "templates", "status.html")
	case "":
		filenames = filepath.Join(rootDir, "templates", "payment.html")

		token, err := c.getToken(requestSrc)
		if err != nil {
			return err
		}

		requestSrc.Token = token
		requestSrc.PaymentJsLink = c.credential.JS

		requestSrc.BackLink = c.configs.APP.Host + "/invoices/" + requestSrc.ID + "/pay"

	default:
		requestSrc.Status.Transaction.Status = "failed"
		filenames = filepath.Join(rootDir, "templates", "status.html")
	}

	requestSrc.Status.Transaction.Date = requestSrc.Status.Transaction.CreatedDate.Format("02.01.2006 15:04")
	if requestSrc.Status.Transaction.StatusName == "EXPIRED" {
		requestSrc.Status.Transaction.Date = requestSrc.DueDate.Format("02.01.2006 15:04")
	}

	// setup request handler
	tmpl, err := template.ParseFiles(filenames)
	if err != nil {
		return
	}

	return tmpl.Execute(w, requestSrc)
}

var ErrFromEpay = errors.New("Произошла ошибка при попытке оплаты, проверьте статус")

func (c *Client) PayByCard(requestSrc Request) (responseSrc *ResponseCardID, err error) {
	//preparation of request params
	token, err := c.getToken(requestSrc)
	if err != nil {
		return
	}

	req, err := parseToRequestByCardID(requestSrc)
	if err != nil {
		return
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return
	}
	reqBytes := bytes.NewReader(reqBody)

	// setup request handler
	path := c.credential.Endpoint + "/payments/cards/auth"
	resBytes, status, err := c.handler("POST", path, "", token.AccessToken, reqBytes, true)
	if err != nil {
		return
	}

	// check response status
	switch status {
	case http.StatusOK:
		err = json.Unmarshal(resBytes, &responseSrc)
	default:
		if ok := json.Unmarshal(resBytes, &responseSrc); ok != nil {
			err = errors.New(string(resBytes))
			return
		}

		err = ErrFromEpay
		return
	}

	return
}
