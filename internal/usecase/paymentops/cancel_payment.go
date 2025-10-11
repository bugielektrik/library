package paymentops

import (
	"context"
	"time"

	"go.uber.org/zap"

	"library-service/internal/domain/payment"
	"library-service/pkg/errors"
	"library-service/pkg/logutil"
)

// CancelPaymentRequest represents the input for cancelling a payment.
type CancelPaymentRequest struct {
	PaymentID string
	MemberID  string // For authorization check
	Reason    string
}

// CancelPaymentResponse represents the output of cancelling a payment.
type CancelPaymentResponse struct {
	PaymentID   string
	Status      payment.Status
	CancelledAt time.Time
}

// CancelPaymentUseCase handles the cancellation of a payment.
type CancelPaymentUseCase struct {
	paymentRepo    payment.Repository
	paymentService *payment.Service
}

// NewCancelPaymentUseCase creates a new instance of CancelPaymentUseCase.
func NewCancelPaymentUseCase(
	paymentRepo payment.Repository,
	paymentService *payment.Service,
) *CancelPaymentUseCase {
	return &CancelPaymentUseCase{
		paymentRepo:    paymentRepo,
		paymentService: paymentService,
	}
}

// Execute cancels a payment if it's in a cancellable state.
func (uc *CancelPaymentUseCase) Execute(ctx context.Context, req CancelPaymentRequest) (CancelPaymentResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "payment", "cancel_payment")

	// Retrieve payment
	paymentEntity, err := uc.paymentRepo.GetByID(ctx, req.PaymentID)
	if err != nil {
		logger.Error("failed to retrieve payment", zap.Error(err))
		return CancelPaymentResponse{}, errors.NotFound("payment")
	}

	// Verify payment belongs to member
	if paymentEntity.MemberID != req.MemberID {
		logger.Warn("unauthorized cancellation attempt",
			zap.String("payment_member_id", paymentEntity.MemberID),
			zap.String("requesting_member_id", req.MemberID),
		)
		return CancelPaymentResponse{}, errors.NotFoundWithID("payment", req.PaymentID)
	}

	// Check if payment can be cancelled
	if paymentEntity.Status == payment.StatusCompleted {
		logger.Warn("cannot cancel completed payment")
		return CancelPaymentResponse{}, errors.ErrInvalidPaymentStatus.
			WithDetails("status", string(paymentEntity.Status)).
			WithDetails("reason", "completed payments cannot be cancelled, use refund instead")
	}

	if paymentEntity.Status == payment.StatusCancelled {
		logger.Warn("payment already cancelled")
		return CancelPaymentResponse{}, errors.ErrPaymentAlreadyProcessed.
			WithDetails("status", string(paymentEntity.Status))
	}

	if paymentEntity.Status == payment.StatusRefunded {
		logger.Warn("cannot cancel refunded payment")
		return CancelPaymentResponse{}, errors.ErrInvalidPaymentStatus.
			WithDetails("status", string(paymentEntity.Status)).
			WithDetails("reason", "refunded payments cannot be cancelled")
	}

	// Validate status transition
	if err := uc.paymentService.ValidateStatusTransition(paymentEntity.Status, payment.StatusCancelled); err != nil {
		logger.Warn("invalid status transition", zap.Error(err))
		return CancelPaymentResponse{}, err
	}

	// Update payment status to cancelled
	if err := uc.paymentRepo.UpdateStatus(ctx, req.PaymentID, payment.StatusCancelled); err != nil {
		logger.Error("failed to update payment status", zap.Error(err))
		return CancelPaymentResponse{}, errors.Database("database operation", err)
	}

	now := time.Now()

	logger.Info("payment cancelled successfully")

	return CancelPaymentResponse{
		PaymentID:   req.PaymentID,
		Status:      payment.StatusCancelled,
		CancelledAt: now,
	}, nil
}
