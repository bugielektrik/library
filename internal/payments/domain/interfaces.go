package domain

import (
	"context"
	"time"
)

// Gateway defines the interface for payment gateway operations.
//
// This interface abstracts external payment gateway integration (epayment.kz, Stripe, etc.)
// following the Dependency Inversion Principle:
//   - Domain layer defines the contract (this interface)
//   - Infrastructure layer implements it (epayment adapter)
//   - Use cases depend on the interface, not the implementation
//
// Design Decisions:
//   - Type-safe responses (no interface{} returns)
//   - Separation of operations vs configuration (see GatewayConfig)
//   - Context-aware for cancellation and timeout
//   - Minimal interface (only methods use cases actually need)
//
// See Also:
//   - Implementation: internal/adapters/payment/epayment/gateway.go
//   - Usage: internal/usecase/paymentops/ (all payment use cases)
//   - ADR: .claude/adr/005-payment-gateway-interface.md
type Gateway interface {
	// GetAuthToken retrieves an authentication token from the payment gateway.
	// The token is typically cached and refreshed automatically before expiry.
	//
	// Returns an error if authentication fails (invalid credentials, network issues, etc.)
	GetAuthToken(ctx context.Context) (string, error)

	// CheckPaymentStatus queries the gateway for the current status of a payment.
	// This is used to verify payments, handle callbacks, and check transaction state.
	//
	// Parameters:
	//   - invoiceID: Unique invoice identifier for the payment
	//
	// Returns detailed transaction status or an error if the query fails.
	CheckPaymentStatus(ctx context.Context, invoiceID string) (*GatewayStatusResponse, error)

	// RefundPayment initiates a refund for a completed payment.
	// Supports both full refunds and partial refunds.
	//
	// Parameters:
	//   - transactionID: Gateway transaction identifier
	//   - amount: Amount to refund (nil for full refund)
	//   - externalID: Optional tracking identifier for reconciliation
	//
	// Returns an error if the refund fails (payment not completed, insufficient funds, etc.)
	RefundPayment(ctx context.Context, transactionID string, amount *float64, externalID string) error

	// CancelPayment cancels a pending payment before it's processed.
	// Only works for payments in pending status.
	//
	// Returns an error if cancellation fails (payment already processed, etc.)
	CancelPayment(ctx context.Context, transactionID string) error

	// ChargeCard charges a previously saved card token.
	// Used for recurring payments or one-click checkout.
	//
	// Returns payment response with transaction ID and status.
	ChargeCard(ctx context.Context, req *CardChargeRequest) (*CardChargeResponse, error)
}

// GatewayConfig provides gateway configuration details.
//
// Separated from Gateway to distinguish operations from configuration.
// Configuration methods don't require context and never fail.
type GatewayConfig interface {
	// GetTerminal returns the merchant terminal ID.
	GetTerminal() string

	// GetBackLink returns the URL where users are redirected after payment.
	GetBackLink() string

	// GetPostLink returns the URL where the gateway sends payment callbacks.
	GetPostLink() string

	// GetWidgetURL returns the JavaScript widget URL for embedded payments.
	GetWidgetURL() string
}

// GatewayStatusResponse represents a standardized payment status check response.
//
// This structure provides a gateway-agnostic view of payment status.
// Gateway-specific implementations map their response format to this structure.
type GatewayStatusResponse struct {
	// ResultCode indicates the result of the status check (success, error, etc.)
	ResultCode string

	// ResultMessage provides a human-readable description of the result
	ResultMessage string

	// Transaction contains detailed transaction information
	Transaction GatewayTransactionDetails
}

// GatewayTransactionDetails contains detailed information about a payment transaction.
type GatewayTransactionDetails struct {
	// ID is the gateway's internal transaction identifier
	ID string

	// InvoiceID is our invoice identifier (matches payment.InvoiceID)
	InvoiceID string

	// Amount is the transaction amount in smallest currency unit (e.g., tenge)
	Amount int64

	// Currency is the three-letter currency code (e.g., "KZT")
	Currency string

	// Status is the gateway's status string (success, failed, pending, etc.)
	// Note: Use payment.Service.MapGatewayStatus() to convert to domain Status
	Status string

	// CardMask is the masked card number (e.g., "400000******0002")
	CardMask string

	// ApprovalCode is the bank approval code for successful transactions
	ApprovalCode string

	// Reference is the gateway's reference number for tracking
	Reference string
}

// CardChargeRequest represents a request to charge a previously saved card.
type CardChargeRequest struct {
	// InvoiceID is our unique invoice identifier for this payment
	InvoiceID string

	// Amount is the charge amount in smallest currency unit (e.g., tenge)
	Amount int64

	// Currency is the three-letter currency code (e.g., "KZT")
	Currency string

	// CardID is the saved card token from the gateway
	CardID string

	// Description is a human-readable payment description
	Description string
}

// CardChargeResponse represents the response from charging a saved card.
type CardChargeResponse struct {
	// ID is the gateway's internal payment identifier
	ID string

	// TransactionID is the gateway's transaction identifier
	TransactionID string

	// Status is the payment status (success, failed, pending, etc.)
	Status string

	// Reference is the gateway's reference number
	Reference string

	// ApprovalCode is the bank approval code (present on success)
	ApprovalCode string

	// ErrorCode is the error code if the charge failed
	ErrorCode string

	// ErrorMessage is a human-readable error message
	ErrorMessage string

	// ProcessedAt is when the transaction was processed
	ProcessedAt *time.Time
}
