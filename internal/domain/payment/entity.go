package payment

import (
	"time"
)

// Status represents the status of a payment.
// State transitions: pending → processing → completed/failed
// Terminal states: completed, failed, cancelled, refunded
type Status string

const (
	StatusPending    Status = "pending"    // Initial state
	StatusProcessing Status = "processing" // Gateway processing
	StatusCompleted  Status = "completed"  // Successfully paid
	StatusFailed     Status = "failed"     // Payment failed
	StatusCancelled  Status = "cancelled"  // User cancelled
	StatusRefunded   Status = "refunded"   // Admin refunded
)

// PaymentMethod represents the method used for payment (card, wallet, etc.)
type PaymentMethod string

// PaymentType represents the purpose of the payment
type PaymentType string

const (
	PaymentTypeFine         PaymentType = "fine"         // Library fines
	PaymentTypeSubscription PaymentType = "subscription" // Membership fees
	PaymentTypeDeposit      PaymentType = "deposit"      // Refundable deposits
)

// Payment represents a payment transaction entity.
// Integrates with epayment.kz gateway via webhooks.
// Amount stored in smallest currency unit (tenge for KZT).
// State transitions managed by payment.Service.
type Payment struct {
	// ID is the unique identifier for the payment.
	ID string `db:"id" bson:"_id"`

	// InvoiceID is the unique invoice identifier used with the payment gateway.
	InvoiceID string `db:"invoice_id" bson:"invoice_id"`

	// MemberID is the ID of the member making the payment.
	MemberID string `db:"member_id" bson:"member_id"`

	// Amount is the payment amount in the smallest currency unit (e.g., tenge).
	Amount int64 `db:"amount" bson:"amount"`

	// Currency is the currency code (e.g., KZT, USD).
	Currency string `db:"currency" bson:"currency"`

	// Status is the current status of the payment.
	Status Status `db:"status" bson:"status"`

	// PaymentMethod is the method used for payment.
	PaymentMethod PaymentMethod `db:"payment_method" bson:"payment_method"`

	// PaymentType is the purpose of the payment.
	PaymentType PaymentType `db:"payment_type" bson:"payment_type"`

	// RelatedEntityID is the ID of the related entity (e.g., fine ID, subscription ID).
	RelatedEntityID *string `db:"related_entity_id" bson:"related_entity_id"`

	// GatewayTransactionID is the transaction ID from the payment gateway.
	GatewayTransactionID *string `db:"gateway_transaction_id" bson:"gateway_transaction_id"`

	// GatewayResponse is the full response from the payment gateway (JSON).
	GatewayResponse *string `db:"gateway_response" bson:"gateway_response"`

	// CardMask is the masked card number (e.g., "****1234").
	CardMask *string `db:"card_mask" bson:"card_mask"`

	// ApprovalCode is the approval code from the payment gateway.
	ApprovalCode *string `db:"approval_code" bson:"approval_code"`

	// ErrorCode is the error code if payment failed.
	ErrorCode *string `db:"error_code" bson:"error_code"`

	// ErrorMessage is the error message if payment failed.
	ErrorMessage *string `db:"error_message" bson:"error_message"`

	// CreatedAt is the timestamp when the payment was created.
	CreatedAt time.Time `db:"created_at" bson:"created_at"`

	// UpdatedAt is the timestamp when the payment was last updated.
	UpdatedAt time.Time `db:"updated_at" bson:"updated_at"`

	// CompletedAt is the timestamp when the payment was completed.
	CompletedAt *time.Time `db:"completed_at" bson:"completed_at"`

	// ExpiresAt is the timestamp when the payment will expire if not completed.
	ExpiresAt time.Time `db:"expires_at" bson:"expires_at"`
}

// New creates a new Payment instance.
func New(req Request) Payment {
	now := time.Now()
	// Default expiration is 30 minutes from creation
	expiresAt := now.Add(30 * time.Minute)

	return Payment{
		InvoiceID:       req.InvoiceID,
		MemberID:        req.MemberID,
		Amount:          req.Amount,
		Currency:        req.Currency,
		Status:          StatusPending,
		PaymentType:     req.PaymentType,
		PaymentMethod:   PaymentMethodCard, // Default to card
		RelatedEntityID: req.RelatedEntityID,
		CreatedAt:       now,
		UpdatedAt:       now,
		ExpiresAt:       expiresAt,
	}
}

// IsPending returns true if the payment is still pending.
func (p Payment) IsPending() bool {
	return p.Status == StatusPending
}

// IsProcessing returns true if the payment is being processed.
func (p Payment) IsProcessing() bool {
	return p.Status == StatusProcessing
}

// IsCompleted returns true if the payment is completed.
func (p Payment) IsCompleted() bool {
	return p.Status == StatusCompleted
}

// IsFailed returns true if the payment has failed.
func (p Payment) IsFailed() bool {
	return p.Status == StatusFailed
}

// IsExpired returns true if the payment has expired based on current time.
func (p Payment) IsExpired() bool {
	return (p.Status == StatusPending || p.Status == StatusProcessing) && time.Now().After(p.ExpiresAt)
}

// CanBeRetried returns true if the payment can be retried.
func (p Payment) CanBeRetried() bool {
	return p.Status == StatusFailed || p.IsExpired()
}

// CanBeCancelled returns true if the payment can be cancelled.
func (p Payment) CanBeCancelled() bool {
	return p.Status == StatusPending || p.Status == StatusProcessing
}

// CanBeRefunded returns true if the payment can be refunded.
func (p Payment) CanBeRefunded() bool {
	return p.Status == StatusCompleted
}
