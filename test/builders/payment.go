// Package builders provides test fixture builders for creating test data.
// These builders implement the Builder pattern to create domain entities
// with sensible defaults that can be overridden as needed.
package builders

import (
	"time"

	"library-service/internal/domain/payment"
)

// PaymentBuilder provides a fluent interface for building Payment test fixtures.
type PaymentBuilder struct {
	payment payment.Payment
}

// NewPayment creates a PaymentBuilder with sensible defaults.
func NewPayment() *PaymentBuilder {
	return &PaymentBuilder{
		payment: payment.Payment{
			ID:            "test-payment-id",
			InvoiceID:     "test-invoice-id",
			MemberID:      "test-member-id",
			Amount:        10000, // 100.00 in smallest currency unit
			Currency:      payment.CurrencyKZT,
			Status:        payment.StatusPending,
			PaymentType:   payment.PaymentTypeFine,
			PaymentMethod: payment.PaymentMethodCard,
			CreatedAt:     time.Now(),
		},
	}
}

// WithID sets the payment ID.
func (b *PaymentBuilder) WithID(id string) *PaymentBuilder {
	b.payment.ID = id
	return b
}

// WithInvoiceID sets the invoice ID.
func (b *PaymentBuilder) WithInvoiceID(invoiceID string) *PaymentBuilder {
	b.payment.InvoiceID = invoiceID
	return b
}

// WithMemberID sets the member ID.
func (b *PaymentBuilder) WithMemberID(memberID string) *PaymentBuilder {
	b.payment.MemberID = memberID
	return b
}

// WithAmount sets the payment amount.
func (b *PaymentBuilder) WithAmount(amount int64) *PaymentBuilder {
	b.payment.Amount = amount
	return b
}

// WithCurrency sets the currency.
func (b *PaymentBuilder) WithCurrency(currency string) *PaymentBuilder {
	b.payment.Currency = currency
	return b
}

// WithStatus sets the payment status.
func (b *PaymentBuilder) WithStatus(status payment.Status) *PaymentBuilder {
	b.payment.Status = status
	return b
}

// WithPaymentType sets the payment type.
func (b *PaymentBuilder) WithPaymentType(paymentType payment.PaymentType) *PaymentBuilder {
	b.payment.PaymentType = paymentType
	return b
}

// WithPaymentMethod sets the payment method.
func (b *PaymentBuilder) WithPaymentMethod(method payment.PaymentMethod) *PaymentBuilder {
	b.payment.PaymentMethod = method
	return b
}

// WithCardMask sets the card mask.
func (b *PaymentBuilder) WithCardMask(cardMask string) *PaymentBuilder {
	b.payment.CardMask = &cardMask
	return b
}

// WithGatewayTransactionID sets the gateway transaction ID.
func (b *PaymentBuilder) WithGatewayTransactionID(txnID string) *PaymentBuilder {
	b.payment.GatewayTransactionID = &txnID
	return b
}

// WithCompletedStatus sets status to completed and adds completion time.
func (b *PaymentBuilder) WithCompletedStatus() *PaymentBuilder {
	now := time.Now()
	b.payment.Status = payment.StatusCompleted
	b.payment.CompletedAt = &now
	return b
}

// WithFailedStatus sets status to failed.
func (b *PaymentBuilder) WithFailedStatus() *PaymentBuilder {
	b.payment.Status = payment.StatusFailed
	return b
}

// WithCancelledStatus sets status to cancelled.
func (b *PaymentBuilder) WithCancelledStatus() *PaymentBuilder {
	b.payment.Status = payment.StatusCancelled
	return b
}

// WithRefundedStatus sets status to refunded.
func (b *PaymentBuilder) WithRefundedStatus() *PaymentBuilder {
	b.payment.Status = payment.StatusRefunded
	return b
}

// Build returns the constructed Payment.
func (b *PaymentBuilder) Build() payment.Payment {
	return b.payment
}

// SavedCardBuilder provides a fluent interface for building SavedCard test fixtures.
type SavedCardBuilder struct {
	card payment.SavedCard
}

// NewSavedCard creates a SavedCardBuilder with sensible defaults.
func NewSavedCard() *SavedCardBuilder {
	return &SavedCardBuilder{
		card: payment.SavedCard{
			ID:          "test-card-id",
			MemberID:    "test-member-id",
			CardToken:   "test-token-123",
			CardMask:    "****1234",
			ExpiryMonth: 12,
			ExpiryYear:  2025,
			CardType:    "visa",
			IsActive:    true,
			IsDefault:   false,
			CreatedAt:   time.Now(),
		},
	}
}

// WithID sets the card ID.
func (b *SavedCardBuilder) WithID(id string) *SavedCardBuilder {
	b.card.ID = id
	return b
}

// WithMemberID sets the member ID.
func (b *SavedCardBuilder) WithMemberID(memberID string) *SavedCardBuilder {
	b.card.MemberID = memberID
	return b
}

// WithCardToken sets the card token.
func (b *SavedCardBuilder) WithCardToken(token string) *SavedCardBuilder {
	b.card.CardToken = token
	return b
}

// WithCardMask sets the card mask.
func (b *SavedCardBuilder) WithCardMask(mask string) *SavedCardBuilder {
	b.card.CardMask = mask
	return b
}

// WithExpiry sets the expiry date.
func (b *SavedCardBuilder) WithExpiry(month, year int) *SavedCardBuilder {
	b.card.ExpiryMonth = month
	b.card.ExpiryYear = year
	return b
}

// WithInactive sets the card as inactive.
func (b *SavedCardBuilder) WithInactive() *SavedCardBuilder {
	b.card.IsActive = false
	return b
}

// WithDefault sets the card as default.
func (b *SavedCardBuilder) WithDefault() *SavedCardBuilder {
	b.card.IsDefault = true
	return b
}

// WithExpired sets the card as expired.
func (b *SavedCardBuilder) WithExpired() *SavedCardBuilder {
	b.card.ExpiryMonth = 1
	b.card.ExpiryYear = 2020
	return b
}

// Build returns the constructed SavedCard.
func (b *SavedCardBuilder) Build() payment.SavedCard {
	return b.card
}

// ReceiptBuilder provides a fluent interface for building Receipt test fixtures.
type ReceiptBuilder struct {
	receipt payment.Receipt
}

// NewReceipt creates a ReceiptBuilder with sensible defaults.
func NewReceipt() *ReceiptBuilder {
	return &ReceiptBuilder{
		receipt: payment.Receipt{
			ID:            "test-receipt-id",
			ReceiptNumber: "RCP-2025-001",
			PaymentID:     "test-payment-id",
			MemberID:      "test-member-id",
			Amount:        10000,
			Currency:      payment.CurrencyKZT,
			ReceiptDate:   time.Now(),
		},
	}
}

// WithID sets the receipt ID.
func (b *ReceiptBuilder) WithID(id string) *ReceiptBuilder {
	b.receipt.ID = id
	return b
}

// WithPaymentID sets the payment ID.
func (b *ReceiptBuilder) WithPaymentID(paymentID string) *ReceiptBuilder {
	b.receipt.PaymentID = paymentID
	return b
}

// WithMemberID sets the member ID.
func (b *ReceiptBuilder) WithMemberID(memberID string) *ReceiptBuilder {
	b.receipt.MemberID = memberID
	return b
}

// WithAmount sets the receipt amount.
func (b *ReceiptBuilder) WithAmount(amount int64) *ReceiptBuilder {
	b.receipt.Amount = amount
	return b
}

// Build returns the constructed Receipt.
func (b *ReceiptBuilder) Build() payment.Receipt {
	return b.receipt
}
