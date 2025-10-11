package paymentops

import (
	"context"
	"time"

	"go.uber.org/zap"

	"library-service/internal/domain/payment"
	"library-service/pkg/logutil"
)

// ExpirePaymentsRequest represents the input for expiring payments.
type ExpirePaymentsRequest struct {
	// Optional: limit the number of payments to expire in one batch
	BatchSize int
}

// ExpirePaymentsResponse represents the output of expiring payments.
type ExpirePaymentsResponse struct {
	ExpiredCount int
	FailedCount  int
	Errors       []string
}

// ExpirePaymentsUseCase handles expiring old pending/processing payments.
type ExpirePaymentsUseCase struct {
	paymentRepo    payment.Repository
	paymentService *payment.Service
}

// NewExpirePaymentsUseCase creates a new instance of ExpirePaymentsUseCase.
func NewExpirePaymentsUseCase(
	paymentRepo payment.Repository,
	paymentService *payment.Service,
) *ExpirePaymentsUseCase {
	return &ExpirePaymentsUseCase{
		paymentRepo:    paymentRepo,
		paymentService: paymentService,
	}
}

// Execute finds and expires all pending/processing payments that have passed their expiration time.
func (uc *ExpirePaymentsUseCase) Execute(ctx context.Context, req ExpirePaymentsRequest) (ExpirePaymentsResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "payment", "expire")

	startTime := time.Now()
	logger.Info("starting payment expiry job")

	// Get all pending payments
	pendingPayments, err := uc.paymentRepo.ListByStatus(ctx, payment.StatusPending)
	if err != nil {
		logger.Error("failed to list pending payments", zap.Error(err))
		return ExpirePaymentsResponse{
			FailedCount: 1,
			Errors:      []string{err.Error()},
		}, err
	}

	// Get all processing payments
	processingPayments, err := uc.paymentRepo.ListByStatus(ctx, payment.StatusProcessing)
	if err != nil {
		logger.Error("failed to list processing payments", zap.Error(err))
		return ExpirePaymentsResponse{
			FailedCount: 1,
			Errors:      []string{err.Error()},
		}, err
	}

	// Combine all payments that could potentially be expired
	allPayments := append(pendingPayments, processingPayments...)

	logger.Info("checking payments for expiration",
		zap.Int("total_pending", len(pendingPayments)),
		zap.Int("total_processing", len(processingPayments)),
		zap.Int("total_to_check", len(allPayments)),
	)

	expiredCount := 0
	failedCount := 0
	var errors []string

	for _, paymentEntity := range allPayments {
		// Check if payment has expired
		if !uc.paymentService.IsExpired(paymentEntity) {
			continue
		}

		logger.Info("expiring payment",
			zap.String("payment_id", paymentEntity.ID),
			zap.String("invoice_id", paymentEntity.InvoiceID),
			zap.String("status", string(paymentEntity.Status)),
			zap.Time("expires_at", paymentEntity.ExpiresAt),
			zap.Duration("expired_for", time.Since(paymentEntity.ExpiresAt)),
		)

		// Update status to failed
		if err := uc.paymentRepo.UpdateStatus(ctx, paymentEntity.ID, payment.StatusFailed); err != nil {
			logger.Error("failed to expire payment",
				zap.String("payment_id", paymentEntity.ID),
				zap.Error(err),
			)
			failedCount++
			errors = append(errors, err.Error())
			continue
		}

		expiredCount++

		// Apply batch size limit if specified
		if req.BatchSize > 0 && expiredCount >= req.BatchSize {
			logger.Info("batch size limit reached", zap.Int("batch_size", req.BatchSize))
			break
		}
	}

	duration := time.Since(startTime)

	logger.Info("payment expiry job completed",
		zap.Int("expired", expiredCount),
		zap.Int("failed", failedCount),
		zap.Duration("duration", duration),
	)

	return ExpirePaymentsResponse{
		ExpiredCount: expiredCount,
		FailedCount:  failedCount,
		Errors:       errors,
	}, nil
}
