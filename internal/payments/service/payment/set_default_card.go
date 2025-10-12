package payment

import (
	"context"
	errors2 "library-service/internal/pkg/errors"
	"library-service/internal/pkg/logutil"

	"go.uber.org/zap"

	"library-service/internal/payments/domain"
)

// SetDefaultCardRequest represents the input for setting a default card.
type SetDefaultCardRequest struct {
	CardID   string
	MemberID string
}

// SetDefaultCardResponse represents the output of setting a default card.
type SetDefaultCardResponse struct {
	Success bool
}

// SetDefaultCardUseCase handles setting a card as the default.
type SetDefaultCardUseCase struct {
	savedCardRepo domain.SavedCardRepository
}

// NewSetDefaultCardUseCase creates a new instance of SetDefaultCardUseCase.
func NewSetDefaultCardUseCase(savedCardRepo domain.SavedCardRepository) *SetDefaultCardUseCase {
	return &SetDefaultCardUseCase{
		savedCardRepo: savedCardRepo,
	}
}

// Execute sets a card as the default for a member.
func (uc *SetDefaultCardUseCase) Execute(ctx context.Context, req SetDefaultCardRequest) (SetDefaultCardResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "payment", "set_default_card")

	// Verify card belongs to member
	card, err := uc.savedCardRepo.GetByID(ctx, req.CardID)
	if err != nil {
		logger.Error("failed to retrieve card", zap.Error(err))
		return SetDefaultCardResponse{}, errors2.NotFoundWithID("card", req.CardID)
	}

	if card.MemberID != req.MemberID {
		logger.Warn("unauthorized set default attempt")
		return SetDefaultCardResponse{}, errors2.NotFoundWithID("card", req.CardID)
	}

	// Verify card can be used
	if !card.CanBeUsed() {
		logger.Warn("card cannot be set as default", zap.Bool("is_active", card.IsActive), zap.Bool("is_expired", card.IsExpired()))
		return SetDefaultCardResponse{}, errors2.NewError(errors2.CodeValidation).
			WithDetail("field", "card_id").
			WithDetail("reason", "card is inactive or expired").
			Build()
	}

	// Set as default
	if err := uc.savedCardRepo.SetAsDefault(ctx, req.MemberID, req.CardID); err != nil {
		logger.Error("failed to set card as default", zap.Error(err))
		return SetDefaultCardResponse{}, errors2.Database("database operation", err)
	}

	logger.Info("card set as default successfully")

	return SetDefaultCardResponse{
		Success: true,
	}, nil
}
