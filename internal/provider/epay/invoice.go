package epay

import (
	"github.com/shopspring/decimal"
)

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
