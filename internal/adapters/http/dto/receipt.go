package dto

import (
	"library-service/internal/payments/domain"
)

// GenerateReceiptRequest represents the request to generate a receipt
type GenerateReceiptRequest struct {
	PaymentID string `json:"payment_id" validate:"required"`
	Notes     string `json:"notes,omitempty"`
}

// ReceiptResponse represents a receipt response
type ReceiptResponse struct {
	ID            string        `json:"id"`
	ReceiptNumber string        `json:"receipt_number"`
	PaymentID     string        `json:"payment_id"`
	MemberID      string        `json:"member_id"`
	Amount        int64         `json:"amount"`
	Currency      string        `json:"currency"`
	PaymentType   string        `json:"payment_type"`
	PaymentMethod string        `json:"payment_method"`
	TransactionID string        `json:"transaction_id,omitempty"`
	PaymentDate   string        `json:"payment_date"`
	ReceiptDate   string        `json:"receipt_date"`
	Status        string        `json:"status"`
	Description   string        `json:"description,omitempty"`
	MemberName    string        `json:"member_name"`
	MemberEmail   string        `json:"member_email"`
	CardMask      string        `json:"card_mask,omitempty"`
	Items         []ReceiptItem `json:"items"`
	TaxAmount     int64         `json:"tax_amount"`
	TotalAmount   int64         `json:"total_amount"`
	Notes         string        `json:"notes,omitempty"`
	CreatedAt     string        `json:"created_at"`
}

// ReceiptItem represents a receipt line item
type ReceiptItem struct {
	Description string `json:"description"`
	Quantity    int    `json:"quantity"`
	UnitPrice   int64  `json:"unit_price"`
	Amount      int64  `json:"amount"`
}

// ListReceiptsResponse represents the response for listing receipts
type ListReceiptsResponse struct {
	Receipts []ReceiptResponse `json:"receipts"`
	Total    int               `json:"total"`
}

// FromReceiptEntity converts a domain.Receipt to ReceiptResponse
func FromReceiptEntity(entity domain.Receipt) ReceiptResponse {
	items := make([]ReceiptItem, len(entity.Items))
	for i, item := range entity.Items {
		items[i] = ReceiptItem{
			Description: item.Description,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			Amount:      item.Amount,
		}
	}

	cardMask := ""
	if entity.CardMask != nil {
		cardMask = *entity.CardMask
	}

	return ReceiptResponse{
		ID:            entity.ID,
		ReceiptNumber: entity.ReceiptNumber,
		PaymentID:     entity.PaymentID,
		MemberID:      entity.MemberID,
		Amount:        entity.Amount,
		Currency:      entity.Currency,
		PaymentType:   string(entity.PaymentType),
		PaymentMethod: string(entity.PaymentMethod),
		TransactionID: entity.TransactionID,
		PaymentDate:   entity.PaymentDate.Format("2006-01-02 15:04:05"),
		ReceiptDate:   entity.ReceiptDate.Format("2006-01-02 15:04:05"),
		Status:        string(entity.Status),
		Description:   entity.Description,
		MemberName:    entity.MemberName,
		MemberEmail:   entity.MemberEmail,
		CardMask:      cardMask,
		Items:         items,
		TaxAmount:     entity.TaxAmount,
		TotalAmount:   entity.TotalAmount,
		Notes:         entity.Notes,
		CreatedAt:     entity.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

// FromReceiptEntities converts a slice of domain.Receipt to slice of ReceiptResponse
func FromReceiptEntities(entities []domain.Receipt) []ReceiptResponse {
	responses := make([]ReceiptResponse, len(entities))
	for i, entity := range entities {
		responses[i] = FromReceiptEntity(entity)
	}
	return responses
}
