package paymentops

import (
	"context"

	"go.uber.org/zap"

	"library-service/internal/domain/payment"
)

// updatePaymentFields updates optional payment fields from gateway response
func updatePaymentFields(paymentEntity *payment.Payment, transaction payment.GatewayTransactionDetails) bool {
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
	paymentRepo payment.Repository,
	paymentEntity *payment.Payment,
	logger *zap.Logger,
) {
	if err := paymentRepo.UpdateStatus(ctx, paymentEntity.ID, payment.StatusFailed); err != nil {
		logger.Error("failed to update expired payment status", zap.Error(err))
	} else {
		paymentEntity.Status = payment.StatusFailed
	}
}

// isPaymentUpdatable checks if payment status allows updates
func isPaymentUpdatable(status payment.Status) bool {
	return status == payment.StatusPending || status == payment.StatusProcessing
}
