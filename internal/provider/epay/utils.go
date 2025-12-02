package epay

import (
	"strings"

	"github.com/shopspring/decimal"
)

func normalizeLang(lang string) string {
	if lang == "" {
		return "rus"
	}

	lang = strings.ToLower(lang)
	switch lang {
	case "ru", "rus":
		return "rus"
	case "kz", "kaz":
		return "kaz"
	default:
		return "eng"
	}
}

func parseToRequestByCardID(data Request) (res RequestByCardID, err error) {
	_, err = decimal.NewFromString(data.Amount)
	if err != nil {
		return
	}

	return RequestByCardID{
		ID:          data.ID,
		IIN:         data.IIN,
		InvoiceID:   data.InvoiceID,
		Amount:      0.00,
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
