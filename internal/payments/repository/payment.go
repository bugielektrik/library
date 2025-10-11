package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"library-service/internal/adapters/repository/postgres"
	"library-service/internal/infrastructure/store"
	"library-service/internal/payments/domain"
	"library-service/pkg/errors"
)

// PaymentRepository implements domain.Repository interface for PostgreSQL.
type PaymentRepository struct {
	postgres.BaseRepository[domain.Payment]
}

// NewPaymentRepository creates a new PostgreSQL payment repository.
func NewPaymentRepository(db *sqlx.DB) *PaymentRepository {
	return &PaymentRepository{
		BaseRepository: postgres.NewBaseRepository[domain.Payment](db, "payments"),
	}
}

// Create inserts a new payment and returns its ID.
func (r *PaymentRepository) Create(ctx context.Context, payment domain.Payment) (string, error) {
	query := `
		INSERT INTO payments (
			invoice_id, member_id, amount, currency, status, payment_method, payment_type,
			related_entity_id, created_at, updated_at, expires_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
		) RETURNING id
	`

	var id string
	err := r.GetDB().QueryRowContext(
		ctx,
		query,
		payment.InvoiceID,
		payment.MemberID,
		payment.Amount,
		payment.Currency,
		payment.Status,
		payment.PaymentMethod,
		payment.PaymentType,
		payment.RelatedEntityID,
		payment.CreatedAt,
		payment.UpdatedAt,
		payment.ExpiresAt,
	).Scan(&id)

	if err != nil {
		return "", fmt.Errorf("failed to create payment: %w", err)
	}

	return id, nil
}

// GetByID retrieves a payment by its ID.
func (r *PaymentRepository) GetByID(ctx context.Context, id string) (domain.Payment, error) {
	query := `
		SELECT
			id, invoice_id, member_id, amount, currency, status, payment_method, payment_type,
			related_entity_id, gateway_transaction_id, gateway_response, card_mask, approval_code,
			error_code, error_message, created_at, updated_at, completed_at, expires_at
		FROM payments
		WHERE id = $1
	`

	var p domain.Payment
	err := r.GetDB().GetContext(ctx, &p, query, id)
	if err != nil {
		err = postgres.HandleSQLError(err)
		if err == store.ErrorNotFound {
			return domain.Payment{}, errors.ErrNotFound.WithDetails("payment_id", id)
		}
		return domain.Payment{}, fmt.Errorf("failed to get payment: %w", err)
	}

	return p, nil
}

// GetByInvoiceID retrieves a payment by its invoice ID.
func (r *PaymentRepository) GetByInvoiceID(ctx context.Context, invoiceID string) (domain.Payment, error) {
	query := `
		SELECT
			id, invoice_id, member_id, amount, currency, status, payment_method, payment_type,
			related_entity_id, gateway_transaction_id, gateway_response, card_mask, approval_code,
			error_code, error_message, created_at, updated_at, completed_at, expires_at
		FROM payments
		WHERE invoice_id = $1
	`

	var p domain.Payment
	err := r.GetDB().GetContext(ctx, &p, query, invoiceID)
	if err != nil {
		err = postgres.HandleSQLError(err)
		if err == store.ErrorNotFound {
			return domain.Payment{}, errors.ErrNotFound.WithDetails("invoice_id", invoiceID)
		}
		return domain.Payment{}, fmt.Errorf("failed to get payment by invoice ID: %w", err)
	}

	return p, nil
}

// Update modifies an existing payment by its ID.
func (r *PaymentRepository) Update(ctx context.Context, id string, payment domain.Payment) error {
	query := `
		UPDATE payments SET
			status = $2,
			payment_method = $3,
			gateway_transaction_id = $4,
			gateway_response = $5,
			card_mask = $6,
			approval_code = $7,
			error_code = $8,
			error_message = $9,
			updated_at = $10,
			completed_at = $11
		WHERE id = $1
	`

	result, err := r.GetDB().ExecContext(
		ctx,
		query,
		id,
		payment.Status,
		payment.PaymentMethod,
		payment.GatewayTransactionID,
		payment.GatewayResponse,
		payment.CardMask,
		payment.ApprovalCode,
		payment.ErrorCode,
		payment.ErrorMessage,
		payment.UpdatedAt,
		payment.CompletedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrNotFound.WithDetails("payment_id", id)
	}

	return nil
}

// ListByMemberID retrieves all payments for a specific member.
func (r *PaymentRepository) ListByMemberID(ctx context.Context, memberID string) ([]domain.Payment, error) {
	query := `
		SELECT
			id, invoice_id, member_id, amount, currency, status, payment_method, payment_type,
			related_entity_id, gateway_transaction_id, gateway_response, card_mask, approval_code,
			error_code, error_message, created_at, updated_at, completed_at, expires_at
		FROM payments
		WHERE member_id = $1
		ORDER BY created_at DESC
	`

	var payments []domain.Payment
	err := r.GetDB().SelectContext(ctx, &payments, query, memberID)
	if err != nil {
		return nil, fmt.Errorf("failed to list payments by member ID: %w", err)
	}

	return payments, nil
}

// ListByStatus retrieves all payments with a specific status.
func (r *PaymentRepository) ListByStatus(ctx context.Context, status domain.Status) ([]domain.Payment, error) {
	query := `
		SELECT
			id, invoice_id, member_id, amount, currency, status, payment_method, payment_type,
			related_entity_id, gateway_transaction_id, gateway_response, card_mask, approval_code,
			error_code, error_message, created_at, updated_at, completed_at, expires_at
		FROM payments
		WHERE status = $1
		ORDER BY created_at DESC
	`

	var payments []domain.Payment
	err := r.GetDB().SelectContext(ctx, &payments, query, status)
	if err != nil {
		return nil, fmt.Errorf("failed to list payments by status: %w", err)
	}

	return payments, nil
}

// UpdateStatus updates the status of a domain.
func (r *PaymentRepository) UpdateStatus(ctx context.Context, id string, status domain.Status) error {
	query := `
		UPDATE payments SET
			status = $2,
			updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.GetDB().ExecContext(ctx, query, id, status)
	if err != nil {
		return fmt.Errorf("failed to update payment status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.ErrNotFound.WithDetails("payment_id", id)
	}

	return nil
}

// ListExpired lists expired pending payments
func (r *PaymentRepository) ListExpired(ctx context.Context) ([]domain.Payment, error) {
	query := `
		SELECT
			id, invoice_id, member_id, amount, currency, status, payment_method, payment_type,
			related_entity_id, gateway_transaction_id, gateway_response, card_mask, approval_code,
			error_code, error_message, created_at, updated_at, completed_at, expires_at
		FROM payments
		WHERE status IN ('pending', 'processing') AND expires_at < NOW()
		ORDER BY created_at DESC
	`

	var payments []domain.Payment
	err := r.GetDB().SelectContext(ctx, &payments, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list expired payments: %w", err)
	}

	return payments, nil
}

// ListPendingByMemberID lists pending payments for a member
func (r *PaymentRepository) ListPendingByMemberID(ctx context.Context, memberID string) ([]domain.Payment, error) {
	query := `
		SELECT
			id, invoice_id, member_id, amount, currency, status, payment_method, payment_type,
			related_entity_id, gateway_transaction_id, gateway_response, card_mask, approval_code,
			error_code, error_message, created_at, updated_at, completed_at, expires_at
		FROM payments
		WHERE member_id = $1 AND status = 'pending'
		ORDER BY created_at DESC
	`

	var payments []domain.Payment
	err := r.GetDB().SelectContext(ctx, &payments, query, memberID)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending payments for member: %w", err)
	}

	return payments, nil
}
