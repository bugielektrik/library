package payment

import (
	"library-service/internal/payments/domain"
	paymentops "library-service/internal/payments/service/payment"
)

// ============================================================================
// Core Payment DTOs
// ============================================================================

// InitiatePaymentRequest represents the request to initiate a payment.
//
// This DTO is used when a member initiates a payment transaction. It contains
// the payment amount, currency, type, and optional related entity reference.
//
// Validation rules:
//   - Amount: Required, must be greater than 0 (in smallest currency unit, e.g., tenge)
//   - Currency: Required, must be exactly 3 characters (ISO 4217 code, e.g., "KZT")
//   - PaymentType: Required, one of the domain-defined payment types
//   - RelatedEntityID: Optional, references associated entity (e.g., reservation, subscription)
//
// Example JSON:
//
//	{
//	  "amount": 5000,
//	  "currency": "KZT",
//	  "payment_type": "reservation",
//	  "related_entity_id": "res_123456"
//	}
type InitiatePaymentRequest struct {
	Amount          int64              `json:"amount" validate:"required,gt=0"`
	Currency        string             `json:"currency" validate:"required,len=3"`
	PaymentType     domain.PaymentType `json:"payment_type" validate:"required"`
	RelatedEntityID *string            `json:"related_entity_id,omitempty"`
}

// InitiatePaymentResponse represents the response for initiating a payment.
//
// After a payment is initiated, this response provides all necessary information
// for the client to complete the payment transaction through the payment provider's
// widget or API.
//
// Fields:
//   - PaymentID: Internal system payment identifier (UUID)
//   - InvoiceID: Gateway-specific invoice/order identifier
//   - AuthToken: OAuth token for provider widget authentication
//   - Terminal: Merchant terminal ID for the transaction
//   - Amount: Payment amount in smallest currency unit
//   - Currency: ISO 4217 currency code (e.g., "KZT")
//   - BackLink: URL where user is redirected after payment completion
//   - PostLink: Webhook URL for payment status callbacks
//   - WidgetURL: JavaScript widget URL for frontend integration
//
// The client should use WidgetURL, AuthToken, and other fields to initialize
// the payment widget or redirect the user to the payment provider.
type InitiatePaymentResponse struct {
	PaymentID string `json:"payment_id"`
	InvoiceID string `json:"invoice_id"`
	AuthToken string `json:"auth_token"`
	Terminal  string `json:"terminal"`
	Amount    int64  `json:"amount"`
	Currency  string `json:"currency"`
	BackLink  string `json:"back_link"`
	PostLink  string `json:"post_link"`
	WidgetURL string `json:"widget_url"`
}

// PaymentResponse represents a complete payment record with all details.
//
// This DTO is returned when retrieving a payment by ID or verifying payment status.
// It includes all payment transaction details including provider-specific information,
// timestamps, and error details if the payment failed.
//
// Fields:
//   - ID: Unique payment identifier (UUID)
//   - InvoiceID: Gateway invoice/order number for external tracking
//   - MemberID: Member who initiated the payment
//   - Amount: Payment amount in smallest currency unit (e.g., tenge)
//   - Currency: ISO 4217 currency code (e.g., "KZT")
//   - Status: Current payment status (pending, processing, completed, failed, cancelled, refunded)
//   - PaymentMethod: How the payment was made (card, saved_card, etc.)
//   - PaymentType: Purpose of payment (reservation, subscription, fine, etc.)
//   - RelatedEntityID: Optional reference to related entity (reservation ID, subscription ID)
//   - GatewayTransactionID: Gateway-assigned transaction identifier
//   - CardMask: Masked card number (e.g., "4405-62**-****-1448") for card payments
//   - ApprovalCode: Bank approval code for successful transactions
//   - ErrorCode: Gateway error code if payment failed
//   - ErrorMessage: Human-readable error description if payment failed
//   - CreatedAt: Payment initiation timestamp (ISO 8601)
//   - UpdatedAt: Last modification timestamp (ISO 8601)
//   - CompletedAt: Payment completion timestamp (ISO 8601), null if not completed
//   - ExpiresAt: Payment expiration timestamp (ISO 8601)
type PaymentResponse struct {
	ID                   string               `json:"id"`
	InvoiceID            string               `json:"invoice_id"`
	MemberID             string               `json:"member_id"`
	Amount               int64                `json:"amount"`
	Currency             string               `json:"currency"`
	Status               domain.Status        `json:"status"`
	PaymentMethod        domain.PaymentMethod `json:"payment_method"`
	PaymentType          domain.PaymentType   `json:"payment_type"`
	RelatedEntityID      *string              `json:"related_entity_id,omitempty"`
	GatewayTransactionID *string              `json:"gateway_transaction_id,omitempty"`
	CardMask             *string              `json:"card_mask,omitempty"`
	ApprovalCode         *string              `json:"approval_code,omitempty"`
	ErrorCode            *string              `json:"error_code,omitempty"`
	ErrorMessage         *string              `json:"error_message,omitempty"`
	CreatedAt            string               `json:"created_at"`
	UpdatedAt            string               `json:"updated_at"`
	CompletedAt          *string              `json:"completed_at,omitempty"`
	ExpiresAt            string               `json:"expires_at"`
}

// PaymentSummaryResponse represents a condensed payment record for list views.
//
// This DTO is used when returning multiple payments (e.g., member payment history)
// where full details are not needed. It includes only the most important fields
// to reduce payload size and improve performance.
//
// Fields:
//   - ID: Unique payment identifier (UUID)
//   - InvoiceID: Gateway invoice number for tracking
//   - Amount: Payment amount in smallest currency unit
//   - Currency: ISO 4217 currency code
//   - Status: Current payment status
//   - PaymentType: Purpose of payment (reservation, subscription, etc.)
//   - CreatedAt: Payment creation timestamp (ISO 8601)
//   - CompletedAt: Payment completion timestamp (ISO 8601), null if not completed
//
// To get full payment details including card info and error messages, use the
// GET /payments/{id} endpoint which returns PaymentResponse.
type PaymentSummaryResponse struct {
	ID          string             `json:"id"`
	InvoiceID   string             `json:"invoice_id"`
	Amount      int64              `json:"amount"`
	Currency    string             `json:"currency"`
	Status      domain.Status      `json:"status"`
	PaymentType domain.PaymentType `json:"payment_type"`
	CreatedAt   string             `json:"created_at"`
	CompletedAt *string            `json:"completed_at,omitempty"`
}

// ListPaymentsResponse represents the response for listing member payments.
//
// This DTO wraps a collection of payment summaries, typically returned when
// a member views their payment history.
type ListPaymentsResponse struct {
	Payments []PaymentSummaryResponse `json:"payments"`
}

// ============================================================================
// Core DTO Converters
// ============================================================================

// ToInitiatePaymentResponse converts a use case InitiatePaymentResponse to DTO format.
//
// This helper eliminates manual field mapping when converting from use case layer
// responses to HTTP API responses. It performs a one-to-one field mapping with
// no transformation logic.
//
// Used by: POST /payments/initiate handler
func ToInitiatePaymentResponse(resp paymentops.InitiatePaymentResponse) InitiatePaymentResponse {
	return InitiatePaymentResponse{
		PaymentID: resp.PaymentID,
		InvoiceID: resp.InvoiceID,
		AuthToken: resp.AuthToken,
		Terminal:  resp.Terminal,
		Amount:    resp.Amount,
		Currency:  resp.Currency,
		BackLink:  resp.BackLink,
		PostLink:  resp.PostLink,
		WidgetURL: resp.WidgetURL,
	}
}

// ToPaymentResponse converts a use case VerifyPaymentResponse to DTO format.
//
// This helper maps payment verification results to the HTTP API response format.
// Note: Some fields from the use case response may be omitted if they are not
// populated (e.g., MemberID, PaymentMethod, PaymentType, timestamps).
//
// Used by: GET /payments/{id} handler (verify payment endpoint)
func ToPaymentResponse(resp paymentops.VerifyPaymentResponse) PaymentResponse {
	return PaymentResponse{
		ID:                   resp.PaymentID,
		InvoiceID:            resp.InvoiceID,
		Status:               resp.Status,
		Amount:               resp.Amount,
		Currency:             resp.Currency,
		GatewayTransactionID: resp.GatewayTransactionID,
		CardMask:             resp.CardMask,
		ApprovalCode:         resp.ApprovalCode,
		ErrorCode:            resp.ErrorCode,
		ErrorMessage:         resp.ErrorMessage,
	}
}

// ToPaymentSummaryResponse converts a use case PaymentSummary to DTO format.
//
// This helper creates a condensed payment representation suitable for list views.
// It maps only the essential fields needed for payment history or summary lists.
//
// Used by: ToPaymentSummaryResponses for batch conversion
func ToPaymentSummaryResponse(resp paymentops.PaymentSummary) PaymentSummaryResponse {
	return PaymentSummaryResponse{
		ID:          resp.ID,
		InvoiceID:   resp.InvoiceID,
		Amount:      resp.Amount,
		Currency:    resp.Currency,
		Status:      resp.Status,
		PaymentType: resp.PaymentType,
		CreatedAt:   resp.CreatedAt,
		CompletedAt: resp.CompletedAt,
	}
}

// ToPaymentSummaryResponses converts a slice of use case PaymentSummary to DTO format.
//
// This helper performs batch conversion of multiple payment summaries, typically
// used when returning a member's payment history or listing payments.
//
// The function pre-allocates the result slice for optimal performance and
// converts each summary using ToPaymentSummaryResponse.
//
// Used by: GET /payments/member/{memberId} handler
func ToPaymentSummaryResponses(summaries []paymentops.PaymentSummary) []PaymentSummaryResponse {
	responses := make([]PaymentSummaryResponse, len(summaries))
	for i, s := range summaries {
		responses[i] = ToPaymentSummaryResponse(s)
	}
	return responses
}
