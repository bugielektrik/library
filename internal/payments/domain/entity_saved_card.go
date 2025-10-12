package domain

import (
	"context"
	"time"
)

// SavedCard represents a saved payment card in the system.
type SavedCard struct {
	// ID is the unique identifier for the saved card.
	ID string `db:"id" bson:"_id"`

	// MemberID is the ID of the member who owns this card.
	MemberID string `db:"member_id" bson:"member_id"`

	// CardToken is the tokenized card reference from payment provider.
	CardToken string `db:"card_token" bson:"card_token"`

	// CardMask is the masked card number (e.g., "****1234").
	CardMask string `db:"card_mask" bson:"card_mask"`

	// CardType is the card brand (Visa, MasterCard, etc.).
	CardType string `db:"card_type" bson:"card_type"`

	// ExpiryMonth is the card expiration month (1-12).
	ExpiryMonth int `db:"expiry_month" bson:"expiry_month"`

	// ExpiryYear is the card expiration year (e.g., 2025).
	ExpiryYear int `db:"expiry_year" bson:"expiry_year"`

	// IsDefault indicates if this is the member's default payment card.
	IsDefault bool `db:"is_default" bson:"is_default"`

	// IsActive indicates if the card is active and can be used.
	IsActive bool `db:"is_active" bson:"is_active"`

	// CreatedAt is the timestamp when the card was saved.
	CreatedAt time.Time `db:"created_at" bson:"created_at"`

	// UpdatedAt is the timestamp when the card was last updated.
	UpdatedAt time.Time `db:"updated_at" bson:"updated_at"`

	// LastUsedAt is the timestamp when the card was last used for payment.
	LastUsedAt *time.Time `db:"last_used_at" bson:"last_used_at"`
}

// NewSavedCard creates a new SavedCard instance.
func NewSavedCard(memberID, cardToken, cardMask, cardType string, expiryMonth, expiryYear int) SavedCard {
	now := time.Now()

	return SavedCard{
		MemberID:    memberID,
		CardToken:   cardToken,
		CardMask:    cardMask,
		CardType:    cardType,
		ExpiryMonth: expiryMonth,
		ExpiryYear:  expiryYear,
		IsDefault:   false,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// IsExpired checks if the card has expired.
func (c SavedCard) IsExpired() bool {
	now := time.Now()
	currentYear := now.Year()
	currentMonth := int(now.Month())

	if c.ExpiryYear < currentYear {
		return true
	}

	if c.ExpiryYear == currentYear && c.ExpiryMonth < currentMonth {
		return true
	}

	return false
}

// CanBeUsed returns true if the card can be used for payments.
func (c SavedCard) CanBeUsed() bool {
	return c.IsActive && !c.IsExpired()
}

// SavedCardRepository defines the interface for saved card repository service.
type SavedCardRepository interface {
	// Create inserts a new saved card and returns its ID.
	Create(ctx context.Context, card SavedCard) (string, error)

	// GetByID retrieves a saved card by its ID.
	GetByID(ctx context.Context, id string) (SavedCard, error)

	// GetByCardToken retrieves a saved card by its token.
	GetByCardToken(ctx context.Context, cardToken string) (SavedCard, error)

	// ListByMemberID retrieves all saved cards for a member.
	ListByMemberID(ctx context.Context, memberID string) ([]SavedCard, error)

	// GetDefaultCard retrieves the default card for a member.
	GetDefaultCard(ctx context.Context, memberID string) (SavedCard, error)

	// Update modifies an existing saved card.
	Update(ctx context.Context, id string, card SavedCard) error

	// Delete removes a saved card by its ID.
	Delete(ctx context.Context, id string) error

	// SetAsDefault sets a card as the default for a member.
	SetAsDefault(ctx context.Context, memberID, cardID string) error

	// Deactivate marks a card as inactive.
	Deactivate(ctx context.Context, id string) error
}
