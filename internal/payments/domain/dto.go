package domain

import (
	"errors"
	"net/http"
	"time"
)

// Request represents the request payload for creating a payment.
type Request struct {
	InvoiceID       string      `json:"invoice_id"`
	MemberID        string      `json:"member_id"`
	Amount          int64       `json:"amount"`
	Currency        string      `json:"currency"`
	PaymentType     PaymentType `json:"payment_type"`
	RelatedEntityID *string     `json:"related_entity_id,omitempty"`
}

// Bind validates the request payload.
func (r *Request) Bind(req *http.Request) error {
	if r.InvoiceID == "" {
		return errors.New("invoice_id: cannot be blank")
	}

	if r.MemberID == "" {
		return errors.New("member_id: cannot be blank")
	}

	if r.Amount <= 0 {
		return errors.New("amount: must be greater than 0")
	}

	if r.Currency == "" {
		return errors.New("currency: cannot be blank")
	}

	if r.PaymentType == "" {
		return errors.New("payment_type: cannot be blank")
	}

	// Validate payment type
	validTypes := map[PaymentType]bool{
		PaymentTypeFine:         true,
		PaymentTypeSubscription: true,
		PaymentTypeDeposit:      true,
	}
	if !validTypes[r.PaymentType] {
		return errors.New("payment_type: invalid payment type")
	}

	return nil
}

// Response represents the response payload for payment service.
type Response struct {
	ID                   string        `json:"id"`
	InvoiceID            string        `json:"invoice_id"`
	MemberID             string        `json:"member_id"`
	Amount               int64         `json:"amount"`
	Currency             string        `json:"currency"`
	Status               Status        `json:"status"`
	PaymentMethod        PaymentMethod `json:"payment_method"`
	PaymentType          PaymentType   `json:"payment_type"`
	RelatedEntityID      *string       `json:"related_entity_id,omitempty"`
	GatewayTransactionID *string       `json:"gateway_transaction_id,omitempty"`
	CardMask             *string       `json:"card_mask,omitempty"`
	ApprovalCode         *string       `json:"approval_code,omitempty"`
	ErrorCode            *string       `json:"error_code,omitempty"`
	ErrorMessage         *string       `json:"error_message,omitempty"`
	CreatedAt            time.Time     `json:"created_at"`
	UpdatedAt            time.Time     `json:"updated_at"`
	CompletedAt          *time.Time    `json:"completed_at,omitempty"`
	ExpiresAt            time.Time     `json:"expires_at"`
}

// ParseFromPayment converts a payment entity to a response payload.
func ParseFromPayment(data Payment) Response {
	return Response{
		ID:                   data.ID,
		InvoiceID:            data.InvoiceID,
		MemberID:             data.MemberID,
		Amount:               data.Amount,
		Currency:             data.Currency,
		Status:               data.Status,
		PaymentMethod:        data.PaymentMethod,
		PaymentType:          data.PaymentType,
		RelatedEntityID:      data.RelatedEntityID,
		GatewayTransactionID: data.GatewayTransactionID,
		CardMask:             data.CardMask,
		ApprovalCode:         data.ApprovalCode,
		ErrorCode:            data.ErrorCode,
		ErrorMessage:         data.ErrorMessage,
		CreatedAt:            data.CreatedAt,
		UpdatedAt:            data.UpdatedAt,
		CompletedAt:          data.CompletedAt,
		ExpiresAt:            data.ExpiresAt,
	}
}

// ParseFromPayments converts a list of payments to a list of response payloads.
func ParseFromPayments(data []Payment) []Response {
	res := make([]Response, len(data))
	for i, payment := range data {
		res[i] = ParseFromPayment(payment)
	}
	return res
}

// InitiatePaymentResponse represents the response for initiating a payment.
type InitiatePaymentResponse struct {
	PaymentID   string `json:"payment_id"`
	InvoiceID   string `json:"invoice_id"`
	AuthToken   string `json:"auth_token"`
	RedirectURL string `json:"redirect_url,omitempty"`
}

// SaveCardRequest represents the request to save a card.
type SaveCardRequest struct {
	CardToken   string `json:"card_token"`
	CardMask    string `json:"card_mask"`
	CardType    string `json:"card_type"`
	ExpiryMonth int    `json:"expiry_month"`
	ExpiryYear  int    `json:"expiry_year"`
}

// Bind validates the save card request.
func (r *SaveCardRequest) Bind(req *http.Request) error {
	if r.CardToken == "" {
		return errors.New("card_token: cannot be blank")
	}

	if r.CardMask == "" {
		return errors.New("card_mask: cannot be blank")
	}

	if r.CardType == "" {
		return errors.New("card_type: cannot be blank")
	}

	if r.ExpiryMonth < 1 || r.ExpiryMonth > 12 {
		return errors.New("expiry_month: must be between 1 and 12")
	}

	currentYear := time.Now().Year()
	if r.ExpiryYear < currentYear {
		return errors.New("expiry_year: card has expired")
	}

	return nil
}

// SavedCardResponse represents a saved card response.
type SavedCardResponse struct {
	ID          string     `json:"id"`
	CardMask    string     `json:"card_mask"`
	CardType    string     `json:"card_type"`
	ExpiryMonth int        `json:"expiry_month"`
	ExpiryYear  int        `json:"expiry_year"`
	IsDefault   bool       `json:"is_default"`
	IsActive    bool       `json:"is_active"`
	IsExpired   bool       `json:"is_expired"`
	CreatedAt   time.Time  `json:"created_at"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
}

// ParseFromSavedCard converts a saved card entity to a response.
func ParseFromSavedCard(card SavedCard) SavedCardResponse {
	return SavedCardResponse{
		ID:          card.ID,
		CardMask:    card.CardMask,
		CardType:    card.CardType,
		ExpiryMonth: card.ExpiryMonth,
		ExpiryYear:  card.ExpiryYear,
		IsDefault:   card.IsDefault,
		IsActive:    card.IsActive,
		IsExpired:   card.IsExpired(),
		CreatedAt:   card.CreatedAt,
		LastUsedAt:  card.LastUsedAt,
	}
}

// ParseFromSavedCards converts a list of saved cards to responses.
func ParseFromSavedCards(cards []SavedCard) []SavedCardResponse {
	responses := make([]SavedCardResponse, len(cards))
	for i, card := range cards {
		responses[i] = ParseFromSavedCard(card)
	}
	return responses
}
