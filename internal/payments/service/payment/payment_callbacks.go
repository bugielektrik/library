package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"

	"library-service/internal/payments/domain"
	errors2 "library-service/internal/pkg/errors"
	"library-service/internal/pkg/logutil"
)

// ================================================================================
// Handle Callback Use Case
// ================================================================================

// PaymentCallbackRequest represents the callback data from payment provider.
type PaymentCallbackRequest struct {
	InvoiceID       string
	TransactionID   string
	Amount          int64
	Currency        string
	Status          string // "success", "failed", "cancelled"
	CardMask        *string
	ApprovalCode    *string
	ErrorCode       *string
	ErrorMessage    *string
	GatewayResponse map[string]interface{}
}

// HandleCallbackResponse represents the output of processing a payment callback.
type HandleCallbackResponse struct {
	PaymentID string
	Status    domain.Status
	Processed bool
}

// HandleCallbackUseCase handles callbacks from the payment provider.
type HandleCallbackUseCase struct {
	paymentRepo    domain.Repository
	paymentService *domain.Service
}

// NewHandleCallbackUseCase creates a new instance of HandleCallbackUseCase.
func NewHandleCallbackUseCase(
	paymentRepo domain.Repository,
	paymentService *domain.Service,
) *HandleCallbackUseCase {
	return &HandleCallbackUseCase{
		paymentRepo:    paymentRepo,
		paymentService: paymentService,
	}
}

// Execute processes a callback from the payment provider.
func (uc *HandleCallbackUseCase) Execute(ctx context.Context, req PaymentCallbackRequest) (HandleCallbackResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "payment", "handle_callback")

	logger.Info("processing payment callback")

	// Get payment by invoice ID
	paymentEntity, err := uc.paymentRepo.GetByInvoiceID(ctx, req.InvoiceID)
	if err != nil {
		logger.Error("failed to get payment by invoice ID", zap.Error(err))
		return HandleCallbackResponse{}, errors2.ErrNotFound.WithDetails("invoice_id", req.InvoiceID)
	}

	// Security check: Validate amount matches
	if req.Amount != paymentEntity.Amount {
		logger.Error("callback amount mismatch",
			zap.Int64("expected_amount", paymentEntity.Amount),
			zap.Int64("callback_amount", req.Amount),
		)
		return HandleCallbackResponse{}, errors2.NewError(errors2.CodeValidation).
			WithDetail("field", "amount").
			WithDetail("reason", "callback amount does not match payment amount").
			Build()
	}

	// Security check: Validate currency matches
	if req.Currency != paymentEntity.Currency {
		logger.Error("callback currency mismatch",
			zap.String("expected_currency", paymentEntity.Currency),
			zap.String("callback_currency", req.Currency),
		)
		return HandleCallbackResponse{}, errors2.NewError(errors2.CodeValidation).
			WithDetail("field", "currency").
			WithDetail("reason", "callback currency does not match payment currency").
			Build()
	}

	// Idempotency check: If payment is already in a final state, don't process again
	if uc.paymentService.IsFinalStatus(paymentEntity.Status) {
		logger.Warn("payment already in final state, skipping callback",
			zap.String("payment_id", paymentEntity.ID),
			zap.String("current_status", string(paymentEntity.Status)),
		)
		return HandleCallbackResponse{
			PaymentID: paymentEntity.ID,
			Status:    paymentEntity.Status,
			Processed: false, // Not processed, already final
		}, nil
	}

	// Determine new status based on provider response
	newStatus := uc.paymentService.MapGatewayStatus(req.Status)

	// Validate status transition
	if err := uc.paymentService.ValidateStatusTransition(paymentEntity.Status, newStatus); err != nil {
		logger.Warn("invalid status transition", zap.Error(err),
			zap.String("current_status", string(paymentEntity.Status)),
			zap.String("new_status", string(newStatus)),
		)
		return HandleCallbackResponse{}, err
	}

	// Store provider response as JSON
	var gatewayResponseStr *string
	if req.GatewayResponse != nil {
		responseJSON, err := json.Marshal(req.GatewayResponse)
		if err != nil {
			logger.Warn("failed to marshal provider response", zap.Error(err))
		} else {
			str := string(responseJSON)
			gatewayResponseStr = &str
		}
	}

	// Update payment entity using domain service
	uc.paymentService.UpdateStatusFromCallback(&paymentEntity, domain.CallbackData{
		TransactionID:   req.TransactionID,
		CardMask:        req.CardMask,
		ApprovalCode:    req.ApprovalCode,
		ErrorCode:       req.ErrorCode,
		ErrorMessage:    req.ErrorMessage,
		GatewayResponse: gatewayResponseStr,
		NewStatus:       newStatus,
	})

	// Update payment in repository
	if err := uc.paymentRepo.Update(ctx, paymentEntity.ID, paymentEntity); err != nil {
		logger.Error("failed to update payment in repository", zap.Error(err))
		return HandleCallbackResponse{}, errors2.Database("database operation", err)
	}

	logger.Info("payment callback processed successfully",
		zap.String("payment_id", paymentEntity.ID),
		zap.String("new_status", string(newStatus)),
	)

	return HandleCallbackResponse{
		PaymentID: paymentEntity.ID,
		Status:    newStatus,
		Processed: true,
	}, nil
}

// ValidateCallbackSecurity performs additional security checks on callbacks.
// Recommended additional security measures for production:
// 1. IP Whitelisting: Only accept callbacks from known payment provider IPs
// 2. HTTPS Only: Ensure callback endpoint only accepts HTTPS requests
// 3. Secret Token: Generate and validate a unique secret token per payment
// 4. Request Signing: Implement HMAC signature verification if provider supports it
// 5. Replay Attack Protection: Store processed callback IDs with timestamps
// 6. Rate Limiting: Limit callback requests per IP/payment to prevent abuse
//
// Current implementation includes:
// - InvoiceID validation (payment must exist in database)
// - Amount and currency validation
// - Status transition validation
// - Idempotency protection (don't process final states twice)
// - Detailed security event logging
func (uc *HandleCallbackUseCase) ValidateCallbackSecurity(ctx context.Context, invoiceID string, amount int64, currency string) error {
	// This is a placeholder for additional security validations
	// In production, implement IP whitelisting, signature verification, etc.
	return nil
}

// ================================================================================
// Process Callback Retries Use Case
// ================================================================================

// ProcessCallbackRetriesRequest represents the input for processing callback retries.
type ProcessCallbackRetriesRequest struct {
	BatchSize int // Number of retries to process in one batch
}

// ProcessCallbackRetriesResponse represents the output of processing callback retries.
type ProcessCallbackRetriesResponse struct {
	ProcessedCount int
	SuccessCount   int
	FailedCount    int
	Errors         []string
}

// ProcessCallbackRetriesUseCase handles processing of pending callback retries.
type ProcessCallbackRetriesUseCase struct {
	callbackRetryRepo domain.CallbackRetryRepository
	handleCallbackUC  *HandleCallbackUseCase
}

// NewProcessCallbackRetriesUseCase creates a new instance of ProcessCallbackRetriesUseCase.
func NewProcessCallbackRetriesUseCase(
	callbackRetryRepo domain.CallbackRetryRepository,
	handleCallbackUC *HandleCallbackUseCase,
) *ProcessCallbackRetriesUseCase {
	return &ProcessCallbackRetriesUseCase{
		callbackRetryRepo: callbackRetryRepo,
		handleCallbackUC:  handleCallbackUC,
	}
}

// Execute processes pending callback retries.
func (uc *ProcessCallbackRetriesUseCase) Execute(ctx context.Context, req ProcessCallbackRetriesRequest) (ProcessCallbackRetriesResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "payment", "process_callback_retries")

	// Default batch size
	batchSize := req.BatchSize
	if batchSize <= 0 {
		batchSize = 50 // Default to 50 retries per batch
	}

	// Get pending retries
	retries, err := uc.callbackRetryRepo.GetPendingRetries(batchSize)
	if err != nil {
		logger.Error("failed to get pending retries", zap.Error(err))
		return ProcessCallbackRetriesResponse{}, errors2.Database("database operation", err)
	}

	if len(retries) == 0 {
		logger.Info("no pending retries to process")
		return ProcessCallbackRetriesResponse{}, nil
	}

	logger.Info("processing callback retries", zap.Int("retry_count", len(retries)))

	var (
		processedCount int
		successCount   int
		failedCount    int
		errors         []string
	)
	for _, retry := range retries {
		processedCount++

		// Mark as processing
		retry.MarkProcessing()
		if err := uc.callbackRetryRepo.Update(retry); err != nil {
			logger.Error("failed to mark retry as processing",
				zap.String("retry_id", retry.ID),
				zap.Error(err),
			)
			errors = append(errors, fmt.Sprintf("retry %s: failed to mark as processing: %v", retry.ID, err))
			continue
		}

		// Parse callback data
		var callbackReq PaymentCallbackRequest
		if err := json.Unmarshal(retry.CallbackData, &callbackReq); err != nil {
			logger.Error("failed to parse callback data",
				zap.String("retry_id", retry.ID),
				zap.Error(err),
			)

			// This is a permanent error - mark as failed
			retry.Status = domain.CallbackRetryStatusFailed
			retry.LastError = fmt.Sprintf("invalid callback data: %v", err)
			if updateErr := uc.callbackRetryRepo.Update(retry); updateErr != nil {
				logger.Error("failed to update retry status", zap.Error(updateErr))
			}

			errors = append(errors, fmt.Sprintf("retry %s: invalid callback data", retry.ID))
			failedCount++
			continue
		}

		// Attempt to process callback
		_, err = uc.handleCallbackUC.Execute(ctx, callbackReq)
		if err != nil {
			logger.Warn("callback retry failed",
				zap.String("retry_id", retry.ID),
				zap.String("payment_id", retry.PaymentID),
				zap.Int("attempt", retry.RetryCount+1),
				zap.Error(err),
			)

			// Increment retry count and update
			retry.IncrementRetry(err.Error())
			if updateErr := uc.callbackRetryRepo.Update(retry); updateErr != nil {
				logger.Error("failed to update retry after failure", zap.Error(updateErr))
			}

			errors = append(errors, fmt.Sprintf("retry %s: %v", retry.ID, err))
			failedCount++
			continue
		}

		// Success - mark as completed
		retry.MarkCompleted()
		if err := uc.callbackRetryRepo.Update(retry); err != nil {
			logger.Error("failed to mark retry as completed",
				zap.String("retry_id", retry.ID),
				zap.Error(err),
			)
			// Don't count this as a failure since the callback actually succeeded
		}

		logger.Info("callback retry succeeded",
			zap.String("retry_id", retry.ID),
			zap.String("payment_id", retry.PaymentID),
		)
		successCount++
	}

	logger.Info("callback retry processing completed")

	return ProcessCallbackRetriesResponse{
		ProcessedCount: processedCount,
		SuccessCount:   successCount,
		FailedCount:    failedCount,
		Errors:         errors,
	}, nil
}

// ================================================================================
// Expire Payments Use Case
// ================================================================================

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
	paymentRepo    domain.Repository
	paymentService *domain.Service
}

// NewExpirePaymentsUseCase creates a new instance of ExpirePaymentsUseCase.
func NewExpirePaymentsUseCase(
	paymentRepo domain.Repository,
	paymentService *domain.Service,
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
	pendingPayments, err := uc.paymentRepo.ListByStatus(ctx, domain.StatusPending)
	if err != nil {
		logger.Error("failed to list pending payments", zap.Error(err))
		return ExpirePaymentsResponse{
			FailedCount: 1,
			Errors:      []string{err.Error()},
		}, err
	}

	// Get all processing payments
	processingPayments, err := uc.paymentRepo.ListByStatus(ctx, domain.StatusProcessing)
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
		if err := uc.paymentRepo.UpdateStatus(ctx, paymentEntity.ID, domain.StatusFailed); err != nil {
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
