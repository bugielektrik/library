package epay

import (
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

type PaymentCardID struct {
	ID interface{} `json:"id"`
}

type InvoiceRequest struct {
	ID             string          `json:"id"`
	DateTime       string          `json:"dateTime"`
	InvoiceID      string          `json:"invoiceId"`
	Amount         decimal.Decimal `json:"amount"`
	AmountBonus    decimal.Decimal `json:"amountBonus"`
	Currency       string          `json:"currency"`
	Terminal       string          `json:"terminal"`
	AccountID      string          `json:"accountId"`
	Description    string          `json:"description"`
	Language       string          `json:"language"`
	CardMask       string          `json:"cardMask"`
	CardType       string          `json:"cardType"`
	Issuer         string          `json:"issuer"`
	Reference      string          `json:"reference"`
	IntReference   string          `json:"intReference"`
	Secure         string          `json:"secure"`
	Secure3D       string          `json:"secure3D"`
	TokenRecipient string          `json:"tokenRecipient"`
	Code           string          `json:"code"`
	Reason         string          `json:"reason"`
	ReasonCode     int             `json:"reasonCode"`
	Name           string          `json:"name"`
	Email          string          `json:"email"`
	Phone          string          `json:"phone"`
	IP             string          `json:"ip"`
	IPCountry      string          `json:"ipCountry"`
	IPCity         string          `json:"ipCity"`
	IPRegion       string          `json:"ipRegion"`
	IPDistrict     string          `json:"ipDistrict"`
	IPLongitude    decimal.Decimal `json:"ipLongitude"`
	IPLatitude     decimal.Decimal `json:"ipLatitude"`
	CardID         string          `json:"cardId"`
}

type StatusResponse struct {
	InvoiceID     string `json:"invoiceID"`
	ResultCode    string `json:"resultCode"`
	ResultMessage string `json:"resultMessage"`
	Transaction   struct {
		ID                string          `json:"id"`
		Date              string          `json:"-"`
		Datetime          time.Time       `json:"dateTime"`
		CreatedDate       time.Time       `json:"createdDate"`
		InvoiceID         string          `json:"invoiceID"`
		Amount            decimal.Decimal `json:"amount"`
		AmountBonus       decimal.Decimal `json:"amountBonus"`
		PayoutAmount      decimal.Decimal `json:"payoutAmount"`
		Currency          string          `json:"currency"`
		Terminal          string          `json:"terminal"`
		AccountID         string          `json:"accountID"`
		Description       string          `json:"description"`
		Language          string          `json:"language"`
		CardMask          string          `json:"cardMask"`
		CardType          string          `json:"cardType"`
		Issuer            string          `json:"issuer"`
		Reference         string          `json:"reference"`
		IntReference      string          `json:"intReference"`
		Secure            bool            `json:"secure"`
		Status            string          `json:"status"`
		StatusID          string          `json:"statusID"`
		StatusName        string          `json:"statusName"`
		StatusTitle       string          `json:"statusTitle"`
		StatusDescription string          `json:"statusDescription"`
		ReasonCode        string          `json:"reasonCode"`
		Reason            string          `json:"reason"`
		Code              string          `json:"code"`
		Name              string          `json:"name"`
		Email             string          `json:"email"`
		Phone             string          `json:"phone"`
		CardID            string          `json:"cardID"`
	} `json:"transaction"`
}
