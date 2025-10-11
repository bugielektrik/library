package savedcard

import (
	"context"

	"go.uber.org/zap"

	"library-service/internal/payments/domain"
	"library-service/pkg/errors"
	"library-service/pkg/logutil"
)

// DeleteSavedCardRequest represents the input for deleting a saved card.
type DeleteSavedCardRequest struct {
	CardID   string
	MemberID string
}

// DeleteSavedCardResponse represents the output of deleting a saved card.
type DeleteSavedCardResponse struct {
	Success bool
}

// DeleteSavedCardUseCase handles deleting a saved card.
type DeleteSavedCardUseCase struct {
	savedCardRepo domain.SavedCardRepository
}

// NewDeleteSavedCardUseCase creates a new instance of DeleteSavedCardUseCase.
func NewDeleteSavedCardUseCase(savedCardRepo domain.SavedCardRepository) *DeleteSavedCardUseCase {
	return &DeleteSavedCardUseCase{
		savedCardRepo: savedCardRepo,
	}
}

// Execute deletes a saved card if it belongs to the member.
func (uc *DeleteSavedCardUseCase) Execute(ctx context.Context, req DeleteSavedCardRequest) (DeleteSavedCardResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "payment", "delete_saved_card")

	// Verify card belongs to member
	card, err := uc.savedCardRepo.GetByID(ctx, req.CardID)
	if err != nil {
		logger.Error("failed to retrieve card", zap.Error(err))
		return DeleteSavedCardResponse{}, errors.NotFoundWithID("card", req.CardID)
	}

	if card.MemberID != req.MemberID {
		logger.Warn("unauthorized delete attempt")
		return DeleteSavedCardResponse{}, errors.NotFoundWithID("card", req.CardID)
	}

	// Delete card
	if err := uc.savedCardRepo.Delete(ctx, req.CardID); err != nil {
		logger.Error("failed to delete card", zap.Error(err))
		return DeleteSavedCardResponse{}, errors.Database("database operation", err)
	}

	logger.Info("card deleted successfully")

	return DeleteSavedCardResponse{
		Success: true,
	}, nil
}
