package payment

import (
	"context"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"

	"library-service/internal/payments/domain"
	"library-service/pkg/errors"
	"library-service/pkg/logutil"
)

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
		return ProcessCallbackRetriesResponse{}, errors.Database("database operation", err)
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
