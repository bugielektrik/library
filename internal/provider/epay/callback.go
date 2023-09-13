package epay

import (
	"time"
)

type CallbackRequest struct {
	ID             string    `json:"id"`
	DateTime       time.Time `json:"dateTime"`
	InvoiceID      string    `json:"invoiceId"`
	InvoiceIDAlt   string    `json:"invoiceIdAlt"`
	Amount         int       `json:"amount"`
	Currency       string    `json:"currency"`
	ApprovalCode   string    `json:"approvalCode"`
	Terminal       string    `json:"terminal"`
	AccountID      string    `json:"accountId"`
	Description    string    `json:"description"`
	Language       string    `json:"language"`
	CardMask       string    `json:"cardMask"`
	CardType       string    `json:"cardType"`
	Issuer         string    `json:"issuer"`
	Reference      string    `json:"reference"`
	Secure         string    `json:"secure"`
	TokenRecipient string    `json:"tokenRecipient"`
	Code           string    `json:"code"`
	Reason         string    `json:"reason"`
	ReasonCode     int       `json:"reasonCode"`
	Name           string    `json:"name"`
	Email          string    `json:"email"`
	Phone          string    `json:"phone"`
	IP             string    `json:"ip"`
	IPCountry      string    `json:"ipCountry"`
	IPCity         string    `json:"ipCity"`
	IPRegion       string    `json:"ipRegion"`
	IPDistrict     string    `json:"ipDistrict"`
	IPLongitude    float64   `json:"ipLongitude"`
	IPLatitude     float64   `json:"ipLatitude"`
	CardID         string    `json:"cardId"`
}
