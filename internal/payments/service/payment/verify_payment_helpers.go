package payment

import (
	"context"

	"go.uber.org/zap"

	"library-service/internal/payments/domain"
)

// updatePaymentFields updates optional payment fields from provider response
func updatePaymentFields(paymentEntity *domain.Payment, transaction domain.GatewayTransactionDetails) bool {
	needsUpdate := false

	// Update transaction ID if changed
	if shouldUpdateStringField(transaction.ID, paymentEntity.GatewayTransactionID) {
		paymentEntity.GatewayTransactionID = &transaction.ID
		needsUpdate = true
	}

	// Update card mask if changed
	if shouldUpdateStringField(transaction.CardMask, paymentEntity.CardMask) {
		paymentEntity.CardMask = &transaction.CardMask
		needsUpdate = true
	}

	// Update approval code if changed
	if shouldUpdateStringField(transaction.ApprovalCode, paymentEntity.ApprovalCode) {
		paymentEntity.ApprovalCode = &transaction.ApprovalCode
		needsUpdate = true
	}

	return needsUpdate
}

// shouldUpdateStringField checks if a string field should be updated
func shouldUpdateStringField(newValue string, currentValue *string) bool {
	if newValue == "" {
		return false
	}
	return currentValue == nil || *currentValue != newValue
}

// handleExpiredPayment updates expired payment status
func handleExpiredPayment(
	ctx context.Context,
	paymentRepo domain.Repository,
	paymentEntity *domain.Payment,
	logger *zap.Logger,
) {
	if err := paymentRepo.UpdateStatus(ctx, paymentEntity.ID, domain.StatusFailed); err != nil {
		logger.Error("failed to update expired payment status", zap.Error(err))
	} else {
		paymentEntity.Status = domain.StatusFailed
	}
}

// isPaymentUpdatable checks if payment status allows updates
func isPaymentUpdatable(status domain.Status) bool {
	return status == domain.StatusPending || status == domain.StatusProcessing
}
