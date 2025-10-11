package receipt

import (
	"context"

	"go.uber.org/zap"

	"library-service/internal/payments/domain"
	"library-service/pkg/errors"
	"library-service/pkg/logutil"
)

// GetReceiptRequest represents the input for getting a receipt.
type GetReceiptRequest struct {
	ReceiptID string
	MemberID  string // For authorization
}

// GetReceiptResponse represents the output of getting a receipt.
type GetReceiptResponse struct {
	Receipt domain.Receipt
}

// GetReceiptUseCase handles retrieving a receipt by ID.
type GetReceiptUseCase struct {
	receiptRepo domain.ReceiptRepository
}

// NewGetReceiptUseCase creates a new instance of GetReceiptUseCase.
func NewGetReceiptUseCase(receiptRepo domain.ReceiptRepository) *GetReceiptUseCase {
	return &GetReceiptUseCase{
		receiptRepo: receiptRepo,
	}
}

// Execute retrieves a receipt by ID.
func (uc *GetReceiptUseCase) Execute(ctx context.Context, req GetReceiptRequest) (GetReceiptResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "payment", "get_receipt")

	logger.Info("retrieving receipt")

	// Get receipt
	receipt, err := uc.receiptRepo.GetByID(req.ReceiptID)
	if err != nil {
		logger.Error("failed to get receipt", zap.Error(err))
		return GetReceiptResponse{}, errors.NotFoundWithID("receipt", req.ReceiptID)
	}

	// Verify receipt belongs to member
	if receipt.MemberID != req.MemberID {
		logger.Warn("unauthorized receipt access attempt",
			zap.String("receipt_member_id", receipt.MemberID),
			zap.String("requesting_member_id", req.MemberID),
		)
		return GetReceiptResponse{}, errors.Unauthorized("invalid credentials")
	}

	logger.Info("receipt retrieved successfully",
		zap.String("receipt_id", req.ReceiptID),
	)

	return GetReceiptResponse{
		Receipt: receipt,
	}, nil
}
