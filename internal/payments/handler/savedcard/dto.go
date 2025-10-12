package savedcard

import (
	"library-service/internal/payments/domain"
)

// SaveCardRequest represents the request for saving a payment card.
type SaveCardRequest struct {
	CardToken   string `json:"card_token" validate:"required"`
	CardMask    string `json:"card_mask" validate:"required"`
	CardType    string `json:"card_type" validate:"required"`
	ExpiryMonth int    `json:"expiry_month" validate:"required,min=1,max=12"`
	ExpiryYear  int    `json:"expiry_year" validate:"required,min=2000"`
}

// ListSavedCardsResponse represents the response for listing saved cards.
type ListSavedCardsResponse struct {
	Cards []domain.SavedCardResponse `json:"cards"`
}

// DeleteSavedCardResponse represents the response for deleting a saved card.
type DeleteSavedCardResponse struct {
	Success bool `json:"success"`
}

// SetDefaultCardResponse represents the response for setting a default card.
type SetDefaultCardResponse struct {
	Success bool `json:"success"`
}
