package postgres

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"

	"library-service/internal/domain/payment"
)

// CallbackRetryRepository implements payment.CallbackRetryRepository for PostgreSQL
type CallbackRetryRepository struct {
	BaseRepository[payment.CallbackRetry]
}

// NewCallbackRetryRepository creates a new PostgreSQL callback retry repository
func NewCallbackRetryRepository(db *sqlx.DB) *CallbackRetryRepository {
	return &CallbackRetryRepository{
		BaseRepository: NewBaseRepository[payment.CallbackRetry](db, "callback_retries"),
	}
}

// Create inserts a new callback retry record
func (r *CallbackRetryRepository) Create(callbackRetry *payment.CallbackRetry) error {
	query := `
		INSERT INTO callback_retries (
			id, payment_id, callback_data, retry_count, max_retries,
			last_error, next_retry_at, status, created_at, updated_at
		) VALUES (
			:id, :payment_id, :callback_data, :retry_count, :max_retries,
			:last_error, :next_retry_at, :status, :created_at, :updated_at
		)
	`

	_, err := r.GetDB().NamedExec(query, map[string]interface{}{
		"id":            callbackRetry.ID,
		"payment_id":    callbackRetry.PaymentID,
		"callback_data": callbackRetry.CallbackData,
		"retry_count":   callbackRetry.RetryCount,
		"max_retries":   callbackRetry.MaxRetries,
		"last_error":    callbackRetry.LastError,
		"next_retry_at": callbackRetry.NextRetryAt,
		"status":        callbackRetry.Status,
		"created_at":    callbackRetry.CreatedAt,
		"updated_at":    callbackRetry.UpdatedAt,
	})

	if err != nil {
		return fmt.Errorf("failed to create callback retry: %w", err)
	}

	return nil
}

// GetByID retrieves a callback retry by ID
func (r *CallbackRetryRepository) GetByID(id string) (*payment.CallbackRetry, error) {
	query := `
		SELECT id, payment_id, callback_data, retry_count, max_retries,
		       last_error, next_retry_at, status, created_at, updated_at
		FROM callback_retries
		WHERE id = $1
	`

	var callbackRetry payment.CallbackRetry
	err := r.GetDB().Get(&callbackRetry, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("callback retry not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get callback retry: %w", err)
	}

	return &callbackRetry, nil
}

// GetPendingRetries retrieves pending callback retries that are due for retry
func (r *CallbackRetryRepository) GetPendingRetries(limit int) ([]*payment.CallbackRetry, error) {
	query := `
		SELECT id, payment_id, callback_data, retry_count, max_retries,
		       last_error, next_retry_at, status, created_at, updated_at
		FROM callback_retries
		WHERE status = $1
		  AND (next_retry_at IS NULL OR next_retry_at <= NOW())
		  AND retry_count < max_retries
		ORDER BY created_at ASC
		LIMIT $2
	`

	var callbackRetries []*payment.CallbackRetry
	err := r.GetDB().Select(&callbackRetries, query, payment.CallbackRetryStatusPending, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending retries: %w", err)
	}

	return callbackRetries, nil
}

// Update updates an existing callback retry record
func (r *CallbackRetryRepository) Update(callbackRetry *payment.CallbackRetry) error {
	query := `
		UPDATE callback_retries
		SET retry_count = :retry_count,
		    last_error = :last_error,
		    next_retry_at = :next_retry_at,
		    status = :status,
		    updated_at = :updated_at
		WHERE id = :id
	`

	result, err := r.GetDB().NamedExec(query, map[string]interface{}{
		"id":            callbackRetry.ID,
		"retry_count":   callbackRetry.RetryCount,
		"last_error":    callbackRetry.LastError,
		"next_retry_at": callbackRetry.NextRetryAt,
		"status":        callbackRetry.Status,
		"updated_at":    callbackRetry.UpdatedAt,
	})

	if err != nil {
		return fmt.Errorf("failed to update callback retry: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("callback retry not found: %s", callbackRetry.ID)
	}

	return nil
}

// Delete removes a callback retry record
func (r *CallbackRetryRepository) Delete(id string) error {
	query := `DELETE FROM callback_retries WHERE id = $1`

	result, err := r.GetDB().Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete callback retry: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("callback retry not found: %s", id)
	}

	return nil
}
