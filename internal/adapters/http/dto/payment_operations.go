package dto

import (
	"library-service/internal/payments/domain"
	paymentops "library-service/internal/payments/operations/payment"
	savedcardops "library-service/internal/payments/operations/savedcard"
)

// CancelPaymentRequest represents a request to cancel a pending or processing domain.
//
// Only payments in "pending" or "processing" status can be cancelled. Completed,
// failed, or already-cancelled payments cannot be cancelled.
//
// Fields:
//   - Reason: Optional explanation for the cancellation (stored for audit purposes)
//
// Example JSON:
//
//	{
//	  "reason": "Customer requested cancellation"
//	}
type CancelPaymentRequest struct {
	Reason string `json:"reason,omitempty"`
}

// CancelPaymentResponse represents the result of a payment cancellation.
//
// After successfully cancelling a payment, this response confirms the cancellation
// and provides the updated status and timestamp.
//
// Fields:
//   - PaymentID: Identifier of the cancelled payment
//   - Status: Updated payment status (should be "cancelled")
//   - CancelledAt: Timestamp when cancellation occurred (ISO 8601)
type CancelPaymentResponse struct {
	PaymentID   string        `json:"payment_id"`
	Status      domain.Status `json:"status"`
	CancelledAt string        `json:"cancelled_at"`
}

// RefundPaymentRequest represents a request to refund a completed domain.
//
// Only successfully completed payments (status "completed") can be refunded.
// The system supports both full and partial refunds.
//
// Validation rules:
//   - RefundAmount: Optional
//   - If nil or omitted: Full refund of the original payment amount
//   - If specified: Partial refund of the specified amount (must be > 0 and <= original amount)
//   - Reason: Optional explanation for the refund (stored for audit and compliance)
//
// Business Rules:
//   - Payment must be in "completed" status
//   - Refund amount cannot exceed the original payment amount
//   - Multiple partial refunds may be allowed (depending on business logic)
//   - Member can only refund their own payments unless they are admin
//
// Example JSON (full refund):
//
//	{
//	  "reason": "Item not delivered"
//	}
//
// Example JSON (partial refund):
//
//	{
//	  "reason": "Partial cancellation",
//	  "refund_amount": 2500
//	}
type RefundPaymentRequest struct {
	Reason       string `json:"reason,omitempty"`
	RefundAmount *int64 `json:"refund_amount,omitempty"` // Optional: if nil, full refund; if specified, partial refund
}

// RefundPaymentResponse represents the result of a successful payment refund.
//
// After processing a refund request, this response confirms the refund was
// processed and provides details about the refunded amount and timing.
//
// Fields:
//   - PaymentID: Identifier of the refunded payment
//   - Status: Updated payment status (should be "refunded")
//   - RefundedAt: Timestamp when refund was processed (ISO 8601)
//   - Amount: Actual refunded amount in smallest currency unit
//   - Currency: Currency of the refund (matches original payment)
//
// Note: The refund may take 3-30 business days to appear in the customer's
// account depending on the payment gateway and bank processing times.
type RefundPaymentResponse struct {
	PaymentID  string        `json:"payment_id"`
	Status     domain.Status `json:"status"`
	RefundedAt string        `json:"refunded_at"`
	Amount     int64         `json:"amount"`
	Currency   string        `json:"currency"`
}

// PayWithSavedCardRequest represents a request to create a payment using a tokenized saved card.
//
// This DTO allows members to make payments without re-entering card details
// by using a previously saved and tokenized card. The actual card data is
// never stored; only the gateway-provided token is used.
//
// Validation rules:
//   - SavedCardID: Required, must reference a valid saved card belonging to the member
//   - Amount: Required, must be greater than 0 (in smallest currency unit)
//   - Currency: Required, exactly 3 characters (ISO 4217 code)
//   - PaymentType: Required, must be a valid payment type
//   - RelatedEntityID: Optional, references the related entity (reservation, subscription, etc.)
//
// Security:
//   - Only the card owner can use their saved cards
//   - Saved cards are validated before charging
//   - CVV may be required depending on gateway configuration
//
// Example JSON:
//
//	{
//	  "saved_card_id": "card_abc123",
//	  "amount": 5000,
//	  "currency": "KZT",
//	  "payment_type": "reservation",
//	  "related_entity_id": "res_456"
//	}
type PayWithSavedCardRequest struct {
	SavedCardID     string             `json:"saved_card_id" validate:"required"`
	Amount          int64              `json:"amount" validate:"required,gt=0"`
	Currency        string             `json:"currency" validate:"required,len=3"`
	PaymentType     domain.PaymentType `json:"payment_type" validate:"required"`
	RelatedEntityID *string            `json:"related_entity_id,omitempty"`
}

// PayWithSavedCardResponse represents the result of a saved card domain.
//
// When a payment is successfully initiated using a saved card, this response
// provides the payment details and confirmation.
//
// Fields:
//   - PaymentID: Unique identifier for the created payment
//   - InvoiceID: Gateway invoice number for tracking
//   - Status: Payment status (typically "processing" or "completed")
//   - Amount: Payment amount in smallest currency unit
//   - Currency: ISO 4217 currency code
//   - CardMask: Masked card number used for the payment (e.g., "4405-62**-****-1448")
//
// Note: Saved card payments may be processed synchronously (immediate result)
// or asynchronously (status updated via callback), depending on gateway configuration.
type PayWithSavedCardResponse struct {
	PaymentID string        `json:"payment_id"`
	InvoiceID string        `json:"invoice_id"`
	Status    domain.Status `json:"status"`
	Amount    int64         `json:"amount"`
	Currency  string        `json:"currency"`
	CardMask  string        `json:"card_mask"`
}

// ToCancelPaymentResponse converts a use case CancelPaymentResponse to DTO format.
//
// This helper maps the payment cancellation result to the HTTP API response format,
// including formatting the cancellation timestamp to ISO 8601 format.
//
// Used by: POST /payments/{id}/cancel handler
func ToCancelPaymentResponse(resp paymentops.CancelPaymentResponse) CancelPaymentResponse {
	return CancelPaymentResponse{
		PaymentID:   resp.PaymentID,
		Status:      resp.Status,
		CancelledAt: resp.CancelledAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// ToRefundPaymentResponse converts a use case RefundPaymentResponse to DTO format.
//
// This helper maps the payment refund result to the HTTP API response format,
// including formatting the refund timestamp to ISO 8601 format.
//
// The refund amount in the response represents the actual amount refunded, which
// may be a partial amount if the request specified RefundAmount.
//
// Used by: POST /payments/{id}/refund handler
func ToRefundPaymentResponse(resp paymentops.RefundPaymentResponse) RefundPaymentResponse {
	return RefundPaymentResponse{
		PaymentID:  resp.PaymentID,
		Status:     resp.Status,
		RefundedAt: resp.RefundedAt.Format("2006-01-02T15:04:05Z07:00"),
		Amount:     resp.Amount,
		Currency:   resp.Currency,
	}
}

// ToPayWithSavedCardResponse converts a use case PayWithSavedCardResponse to DTO format.
//
// This helper maps the saved card payment result to the HTTP API response format.
// It includes the masked card number for display to the user while never exposing
// the actual card details or token.
//
// Used by: POST /payments/pay-with-card handler
func ToPayWithSavedCardResponse(resp savedcardops.PayWithSavedCardResponse) PayWithSavedCardResponse {
	return PayWithSavedCardResponse{
		PaymentID: resp.PaymentID,
		InvoiceID: resp.InvoiceID,
		Status:    resp.Status,
		Amount:    resp.Amount,
		Currency:  resp.Currency,
		CardMask:  resp.CardMask,
	}
}
