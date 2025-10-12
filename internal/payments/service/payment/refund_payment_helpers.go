package payment

import (
	errors2 "library-service/internal/pkg/errors"

	"go.uber.org/zap"

	"library-service/internal/payments/domain"
)

// validateRefundAuthorization checks if the member is authorized to refund the payment
func validateRefundAuthorization(req RefundPaymentRequest, paymentEntity domain.Payment, logger *zap.Logger) error {
	// Admin can refund any payment
	if req.IsAdmin {
		return nil
	}

	// Member can only refund their own payment
	if paymentEntity.MemberID != req.MemberID {
		logger.Warn("unauthorized refund attempt",
			zap.String("payment_member_id", paymentEntity.MemberID),
			zap.String("requesting_member_id", req.MemberID),
		)
		return errors2.NotFoundWithID("payment", req.PaymentID)
	}

	return nil
}

// validateRefundAmount validates and returns the refund amount
func validateRefundAmount(req RefundPaymentRequest, paymentEntity domain.Payment, logger *zap.Logger) (int64, bool, error) {
	// Default to full refund if no amount specified
	if req.RefundAmount == nil {
		return paymentEntity.Amount, false, nil
	}

	requestedAmount := *req.RefundAmount

	// Check if amount is positive
	if requestedAmount <= 0 {
		logger.Warn("invalid refund amount",
			zap.Int64("requested_amount", requestedAmount),
		)
		return 0, false, errors2.NewError(errors2.CodeValidation).
			WithDetail("field", "refund_amount").
			WithDetail("reason", "refund amount must be positive").
			Build()
	}

	// Check if amount exceeds payment amount
	if requestedAmount > paymentEntity.Amount {
		logger.Warn("refund amount exceeds payment amount",
			zap.Int64("requested_amount", requestedAmount),
			zap.Int64("payment_amount", paymentEntity.Amount),
		)
		return 0, false, errors2.NewError(errors2.CodeValidation).
			WithDetail("field", "refund_amount").
			WithDetail("reason", "refund amount cannot exceed payment amount").
			Build()
	}

	// Valid partial refund
	logger.Info("processing partial refund")
	return requestedAmount, true, nil
}

// validateRefundEligibility checks if the payment can be refunded
func validateRefundEligibility(paymentEntity domain.Payment, paymentService *domain.Service, logger *zap.Logger) error {
	// Check if payment can be refunded using domain logic
	if !paymentEntity.CanBeRefunded() {
		logger.Warn("payment cannot be refunded",
			zap.String("status", string(paymentEntity.Status)),
			zap.Bool("is_expired", paymentEntity.IsExpired()),
		)
		return errors2.ErrInvalidPaymentStatus.
			WithDetails("status", string(paymentEntity.Status)).
			WithDetails("reason", "only completed payments within 180 days can be refunded")
	}

	// Validate status transition
	if err := paymentService.ValidateStatusTransition(paymentEntity.Status, domain.StatusRefunded); err != nil {
		logger.Warn("invalid status transition", zap.Error(err))
		return err
	}

	return nil
}
