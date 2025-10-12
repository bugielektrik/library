package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"library-service/internal/pkg/errors"
	"library-service/internal/pkg/repository/postgres"

	"github.com/jmoiron/sqlx"

	"library-service/internal/payments/domain"
)

// SavedCardRepository implements domain.SavedCardRepository interface for PostgreSQL.
type SavedCardRepository struct {
	postgres.BaseRepository[domain.SavedCard]
}

// Compile-time check that SavedCardRepository implements domain.SavedCardRepository
var _ domain.SavedCardRepository = (*SavedCardRepository)(nil)

// NewSavedCardRepository creates a new PostgreSQL saved card repository.
func NewSavedCardRepository(db *sqlx.DB) *SavedCardRepository {
	return &SavedCardRepository{
		BaseRepository: postgres.NewBaseRepository[domain.SavedCard](db, "saved_cards"),
	}
}

// Create inserts a new saved card and returns its ID.
func (r *SavedCardRepository) Create(ctx context.Context, card domain.SavedCard) (string, error) {
	query := `
		INSERT INTO saved_cards (
			member_id, card_token, card_mask, card_type, expiry_month, expiry_year,
			is_default, is_active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		) RETURNING id
	`

	var id string
	err := r.GetDB().QueryRowContext(
		ctx,
		query,
		card.MemberID,
		card.CardToken,
		card.CardMask,
		card.CardType,
		card.ExpiryMonth,
		card.ExpiryYear,
		card.IsDefault,
		card.IsActive,
		card.CreatedAt,
		card.UpdatedAt,
	).Scan(&id)

	if err != nil {
		return "", fmt.Errorf("failed to create saved card: %w", err)
	}

	return id, nil
}

// GetByID retrieves a saved card by its ID.
func (r *SavedCardRepository) GetByID(ctx context.Context, id string) (domain.SavedCard, error) {
	query := `
		SELECT
			id, member_id, card_token, card_mask, card_type, expiry_month, expiry_year,
			is_default, is_active, created_at, updated_at, last_used_at
		FROM saved_cards
		WHERE id = $1
	`

	var card domain.SavedCard
	err := r.GetDB().GetContext(ctx, &card, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.SavedCard{}, errors.ErrNotFound.WithDetails("card_id", id)
		}
		return domain.SavedCard{}, fmt.Errorf("failed to get saved card: %w", err)
	}

	return card, nil
}

// GetByCardToken retrieves a saved card by its token.
func (r *SavedCardRepository) GetByCardToken(ctx context.Context, cardToken string) (domain.SavedCard, error) {
	query := `
		SELECT
			id, member_id, card_token, card_mask, card_type, expiry_month, expiry_year,
			is_default, is_active, created_at, updated_at, last_used_at
		FROM saved_cards
		WHERE card_token = $1
	`

	var card domain.SavedCard
	err := r.GetDB().GetContext(ctx, &card, query, cardToken)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.SavedCard{}, errors.ErrNotFound.WithDetails("card_token", cardToken)
		}
		return domain.SavedCard{}, fmt.Errorf("failed to get saved card by token: %w", err)
	}

	return card, nil
}

// ListByMemberID retrieves all saved cards for a member.
func (r *SavedCardRepository) ListByMemberID(ctx context.Context, memberID string) ([]domain.SavedCard, error) {
	query := `
		SELECT
			id, member_id, card_token, card_mask, card_type, expiry_month, expiry_year,
			is_default, is_active, created_at, updated_at, last_used_at
		FROM saved_cards
		WHERE member_id = $1
		ORDER BY is_default DESC, created_at DESC
	`

	var cards []domain.SavedCard
	err := r.GetDB().SelectContext(ctx, &cards, query, memberID)
	if err != nil {
		return nil, fmt.Errorf("failed to list saved cards: %w", err)
	}

	return cards, nil
}

// GetDefaultCard retrieves the default card for a member.
func (r *SavedCardRepository) GetDefaultCard(ctx context.Context, memberID string) (domain.SavedCard, error) {
	query := `
		SELECT
			id, member_id, card_token, card_mask, card_type, expiry_month, expiry_year,
			is_default, is_active, created_at, updated_at, last_used_at
		FROM saved_cards
		WHERE member_id = $1 AND is_default = TRUE AND is_active = TRUE
		LIMIT 1
	`

	var card domain.SavedCard
	err := r.GetDB().GetContext(ctx, &card, query, memberID)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.SavedCard{}, errors.ErrNotFound.WithDetails("member_id", memberID)
		}
		return domain.SavedCard{}, fmt.Errorf("failed to get default card: %w", err)
	}

	return card, nil
}

// Update modifies an existing saved card.
func (r *SavedCardRepository) Update(ctx context.Context, id string, card domain.SavedCard) error {
	query := `
		UPDATE saved_cards SET
			card_mask = $2,
			card_type = $3,
			expiry_month = $4,
			expiry_year = $5,
			is_default = $6,
			is_active = $7,
			updated_at = $8,
			last_used_at = $9
		WHERE id = $1
	`

	result, err := r.GetDB().ExecContext(
		ctx,
		query,
		id,
		card.CardMask,
		card.CardType,
		card.ExpiryMonth,
		card.ExpiryYear,
		card.IsDefault,
		card.IsActive,
		card.UpdatedAt,
		card.LastUsedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update saved card: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrNotFound.WithDetails("card_id", id)
	}

	return nil
}

// Delete is inherited from BaseRepository

// SetAsDefault sets a card as the default for a member.
func (r *SavedCardRepository) SetAsDefault(ctx context.Context, memberID, cardID string) error {
	// Start transaction
	tx, err := r.GetDB().BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Unset current default
	unsetQuery := `
		UPDATE saved_cards
		SET is_default = FALSE, updated_at = CURRENT_TIMESTAMP
		WHERE member_id = $1 AND is_default = TRUE
	`
	_, err = tx.ExecContext(ctx, unsetQuery, memberID)
	if err != nil {
		return fmt.Errorf("failed to unset current default: %w", err)
	}

	// Set new default
	setQuery := `
		UPDATE saved_cards
		SET is_default = TRUE, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND member_id = $2 AND is_active = TRUE
	`
	result, err := tx.ExecContext(ctx, setQuery, cardID, memberID)
	if err != nil {
		return fmt.Errorf("failed to set new default: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrNotFound.WithDetails("card_id", cardID)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Deactivate marks a card as inactive.
func (r *SavedCardRepository) Deactivate(ctx context.Context, id string) error {
	query := `
		UPDATE saved_cards
		SET is_active = FALSE, is_default = FALSE, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	result, err := r.GetDB().ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to deactivate saved card: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrNotFound.WithDetails("card_id", id)
	}

	return nil
}
