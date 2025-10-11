package payment

import (
	"fmt"
	"strings"
	"time"

	"library-service/pkg/errors"
)

// Service encapsulates business logic for payments that doesn't naturally
// belong to a single entity. This is a domain service in DDD terms.
//
// Key Responsibilities:
//   - Payment validation (amount, currency, type)
//   - Status transition rules (state machine)
//   - Gateway status mapping
//   - Invoice ID generation
//   - Refund policy enforcement
//
// See Also:
//   - Use case example: internal/usecase/paymentops/initiate_payment.go (orchestrates this service)
//   - Gateway interface: internal/usecase/paymentops/initiate_payment.go:35 (PaymentGateway)
//   - Similar services: internal/domain/book/service.go (comprehensive domain service example)
//   - ADR: .claude/adr/003-domain-services-vs-infrastructure.md (domain vs infrastructure services)
//   - ADR: .claude/adr/005-payment-gateway-interface.md (gateway abstraction)
//   - Test: internal/usecase/paymentops/initiate_payment_test.go
type Service struct {
	// Domain services are typically stateless
	// If state is needed, it should be passed as parameters
}

// NewService creates a new payment domain service.
func NewService() *Service {
	return &Service{}
}

// ValidatePayment validates a payment entity according to business rules.
func (s *Service) Validate(payment Payment) error {
	if payment.MemberID == "" {
		return errors.ErrValidation.WithDetails("field", "member_id")
	}

	if payment.InvoiceID == "" {
		return errors.ErrValidation.WithDetails("field", "invoice_id")
	}

	if payment.Amount <= 0 {
		return errors.ErrValidation.WithDetails("field", "amount").WithDetails("reason", "amount must be greater than 0")
	}

	if payment.Currency == "" {
		return errors.ErrValidation.WithDetails("field", "currency")
	}

	// Validate currency code (ISO 4217)
	if !s.isValidCurrency(payment.Currency) {
		return errors.ErrValidation.WithDetails("field", "currency").WithDetails("reason", "invalid currency code")
	}

	// Validate payment type
	validTypes := map[PaymentType]bool{
		PaymentTypeFine:         true,
		PaymentTypeSubscription: true,
		PaymentTypeDeposit:      true,
	}
	if !validTypes[payment.PaymentType] {
		return errors.ErrValidation.WithDetails("field", "payment_type")
	}

	return nil
}

// ValidateStatusTransition validates if a status transition is allowed.
func (s *Service) ValidateStatusTransition(currentStatus, newStatus Status) error {
	// Define allowed status transitions
	allowedTransitions := map[Status][]Status{
		StatusPending: {
			StatusProcessing,
			StatusCancelled,
			StatusFailed,
		},
		StatusProcessing: {
			StatusCompleted,
			StatusFailed,
			StatusCancelled,
		},
		StatusCompleted: {
			StatusRefunded,
		},
		StatusFailed: {
			StatusPending, // Allow retry
		},
		StatusCancelled: {},
		StatusRefunded:  {},
	}

	allowed, exists := allowedTransitions[currentStatus]
	if !exists {
		return errors.ErrValidation.WithDetails("reason", fmt.Sprintf("invalid current status: %s", currentStatus))
	}

	for _, status := range allowed {
		if status == newStatus {
			return nil
		}
	}

	return errors.ErrValidation.WithDetails(
		"reason",
		fmt.Sprintf("status transition from %s to %s is not allowed", currentStatus, newStatus),
	)
}

// CalculateAmount calculates the payment amount based on payment type and related entity.
// This is a placeholder for complex business logic.
func (s *Service) CalculateAmount(paymentType PaymentType, relatedEntityID string) (int64, error) {
	// In a real system, this would:
	// - For fines: look up the fine amount
	// - For subscriptions: look up the subscription plan price
	// - For deposits: use a configured deposit amount

	// Placeholder implementation
	switch paymentType {
	case PaymentTypeFine:
		// Would query fine repository
		return 0, errors.ErrValidation.WithDetails("reason", "fine calculation not implemented")
	case PaymentTypeSubscription:
		// Would query subscription plan
		return 0, errors.ErrValidation.WithDetails("reason", "subscription calculation not implemented")
	case PaymentTypeDeposit:
		// Would use configured deposit amount
		return 0, errors.ErrValidation.WithDetails("reason", "deposit calculation not implemented")
	default:
		return 0, errors.ErrValidation.WithDetails("reason", "invalid payment type")
	}
}

// IsExpired checks if a payment has expired.
func (s *Service) IsExpired(payment Payment) bool {
	return payment.IsExpired()
}

// CanRefund checks if a payment can be refunded based on business rules.
func (s *Service) CanRefund(payment Payment, refundPolicy time.Duration) error {
	if !payment.CanBeRefunded() {
		return errors.ErrValidation.WithDetails("reason", "payment cannot be refunded in current status")
	}

	if payment.CompletedAt == nil {
		return errors.ErrValidation.WithDetails("reason", "payment has no completion date")
	}

	// Check if payment is within refund policy period
	if time.Since(*payment.CompletedAt) > refundPolicy {
		return errors.ErrValidation.WithDetails("reason", "payment is outside refund policy period")
	}

	return nil
}

// GenerateInvoiceID generates a unique invoice ID.
func (s *Service) GenerateInvoiceID(memberID string, paymentType PaymentType) string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%s-%s-%d", paymentType, memberID, timestamp)
}

// isValidCurrency validates if a currency code is valid.
func (s *Service) isValidCurrency(currency string) bool {
	// Common currencies - expand as needed
	validCurrencies := map[string]bool{
		"KZT": true, // Kazakhstani Tenge
		"USD": true, // US Dollar
		"EUR": true, // Euro
		"RUB": true, // Russian Ruble
	}

	return validCurrencies[currency]
}

// FormatAmount formats an amount for display (converts from smallest unit).
func (s *Service) FormatAmount(amount int64, currency string) string {
	switch currency {
	case "KZT":
		return fmt.Sprintf("%.2f KZT", float64(amount)/100)
	case "USD", "EUR":
		return fmt.Sprintf("%.2f %s", float64(amount)/100, currency)
	case "RUB":
		return fmt.Sprintf("%.2f RUB", float64(amount)/100)
	default:
		return fmt.Sprintf("%d %s", amount, currency)
	}
}

// CallbackData represents callback data from payment gateway.
type CallbackData struct {
	TransactionID   string
	CardMask        *string
	ApprovalCode    *string
	ErrorCode       *string
	ErrorMessage    *string
	GatewayResponse *string
	NewStatus       Status
}

// UpdateStatusFromCallback updates a payment entity based on callback data.
//
// This method encapsulates the business rules for status transitions:
//   - Sets the new status
//   - Updates gateway-provided fields
//   - Sets CompletedAt timestamp for completed payments
//   - Updates the UpdatedAt timestamp
//
// The caller should validate the status transition before calling this method.
func (s *Service) UpdateStatusFromCallback(payment *Payment, data CallbackData) {
	// Update status
	payment.Status = data.NewStatus

	// Update gateway fields
	payment.GatewayTransactionID = &data.TransactionID
	payment.CardMask = data.CardMask
	payment.ApprovalCode = data.ApprovalCode
	payment.ErrorCode = data.ErrorCode
	payment.ErrorMessage = data.ErrorMessage
	payment.GatewayResponse = data.GatewayResponse

	// Update timestamps
	payment.UpdatedAt = time.Now()

	// Set completion time if payment is completed
	if data.NewStatus == StatusCompleted {
		now := time.Now()
		payment.CompletedAt = &now
	}
}

// MapGatewayStatus maps gateway status strings to internal payment status.
//
// This method encodes the business rules for interpreting payment gateway responses:
//   - "success", "approved" → StatusCompleted
//   - "failed", "declined" → StatusFailed
//   - "cancelled" → StatusCancelled
//   - "processing" → StatusProcessing
//   - unknown values → StatusFailed (safe default)
//
// The mapping is defensive - unknown statuses default to Failed to prevent
// payments from being incorrectly marked as successful.
func (s *Service) MapGatewayStatus(gatewayStatus string) Status {
	// Normalize to lowercase for case-insensitive comparison
	status := strings.ToLower(gatewayStatus)

	switch status {
	case "success", "approved", "completed":
		return StatusCompleted
	case "failed", "declined":
		return StatusFailed
	case "cancelled", "voided":
		return StatusCancelled
	case "processing", "auth":
		return StatusProcessing
	case "pending":
		return StatusPending
	case "refunded":
		return StatusRefunded
	default:
		// Default to failed for unknown statuses (defensive programming)
		return StatusFailed
	}
}

// IsFinalStatus checks if a status is final (cannot transition further).
//
// Final statuses are:
//   - StatusCompleted (payment succeeded)
//   - StatusRefunded (payment was refunded)
//   - StatusCancelled (payment was cancelled)
//
// Payments in final statuses should not be updated by callbacks to ensure idempotency.
func (s *Service) IsFinalStatus(status Status) bool {
	return status == StatusCompleted ||
		status == StatusRefunded ||
		status == StatusCancelled
}
