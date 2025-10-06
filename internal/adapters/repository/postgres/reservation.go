package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"library-service/internal/domain/reservation"
	"library-service/internal/infrastructure/store"
)

// ReservationRepository handles CRUD operations for reservations in a PostgreSQL store.
type ReservationRepository struct {
	db *sqlx.DB
}

// NewReservationRepository creates a new ReservationRepository.
func NewReservationRepository(db *sqlx.DB) *ReservationRepository {
	return &ReservationRepository{db: db}
}

// Create inserts a new reservation into the store.
func (r *ReservationRepository) Create(ctx context.Context, data reservation.Reservation) (string, error) {
	query := `
		INSERT INTO reservations (book_id, member_id, status, created_at, expires_at, fulfilled_at, cancelled_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	args := []interface{}{
		data.BookID,
		data.MemberID,
		data.Status,
		data.CreatedAt,
		data.ExpiresAt,
		data.FulfilledAt,
		data.CancelledAt,
	}

	var id string
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&id)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return "", store.ErrorNotFound
	}
	return id, err
}

// GetByID retrieves a reservation by ID from the store.
func (r *ReservationRepository) GetByID(ctx context.Context, id string) (reservation.Reservation, error) {
	query := `
		SELECT id, book_id, member_id, status, created_at, expires_at, fulfilled_at, cancelled_at
		FROM reservations
		WHERE id=$1
	`
	var res reservation.Reservation
	err := r.db.GetContext(ctx, &res, query, id)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return res, store.ErrorNotFound
	}
	return res, err
}

// GetByMemberID retrieves all reservations for a specific member.
func (r *ReservationRepository) GetByMemberID(ctx context.Context, memberID string) ([]reservation.Reservation, error) {
	query := `
		SELECT id, book_id, member_id, status, created_at, expires_at, fulfilled_at, cancelled_at
		FROM reservations
		WHERE member_id=$1
		ORDER BY created_at DESC
	`
	var reservations []reservation.Reservation
	err := r.db.SelectContext(ctx, &reservations, query, memberID)
	if err != nil {
		return nil, err
	}
	return reservations, nil
}

// GetByBookID retrieves all reservations for a specific book.
func (r *ReservationRepository) GetByBookID(ctx context.Context, bookID string) ([]reservation.Reservation, error) {
	query := `
		SELECT id, book_id, member_id, status, created_at, expires_at, fulfilled_at, cancelled_at
		FROM reservations
		WHERE book_id=$1
		ORDER BY created_at ASC
	`
	var reservations []reservation.Reservation
	err := r.db.SelectContext(ctx, &reservations, query, bookID)
	if err != nil {
		return nil, err
	}
	return reservations, nil
}

// GetActiveByMemberAndBook retrieves active reservations for a member and book combination.
func (r *ReservationRepository) GetActiveByMemberAndBook(ctx context.Context, memberID, bookID string) ([]reservation.Reservation, error) {
	query := `
		SELECT id, book_id, member_id, status, created_at, expires_at, fulfilled_at, cancelled_at
		FROM reservations
		WHERE member_id=$1 AND book_id=$2 AND status IN ('pending', 'fulfilled')
		ORDER BY created_at ASC
	`
	var reservations []reservation.Reservation
	err := r.db.SelectContext(ctx, &reservations, query, memberID, bookID)
	if err != nil {
		return nil, err
	}
	return reservations, nil
}

// Update modifies an existing reservation in the store.
func (r *ReservationRepository) Update(ctx context.Context, data reservation.Reservation) error {
	query := `
		UPDATE reservations
		SET status=$1, fulfilled_at=$2, cancelled_at=$3, updated_at=CURRENT_TIMESTAMP
		WHERE id=$4
		RETURNING id
	`
	args := []interface{}{
		data.Status,
		data.FulfilledAt,
		data.CancelledAt,
		data.ID,
	}

	var id string
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&id)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return store.ErrorNotFound
	}
	return err
}

// Delete removes a reservation by ID from the store.
func (r *ReservationRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM reservations WHERE id=$1 RETURNING id`
	var resultID string
	err := r.db.QueryRowContext(ctx, query, id).Scan(&resultID)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		return store.ErrorNotFound
	}
	return err
}

// ListPending retrieves all pending reservations.
func (r *ReservationRepository) ListPending(ctx context.Context) ([]reservation.Reservation, error) {
	query := `
		SELECT id, book_id, member_id, status, created_at, expires_at, fulfilled_at, cancelled_at
		FROM reservations
		WHERE status='pending'
		ORDER BY created_at ASC
	`
	var reservations []reservation.Reservation
	err := r.db.SelectContext(ctx, &reservations, query)
	if err != nil {
		return nil, err
	}
	return reservations, nil
}

// ListExpired retrieves all expired reservations.
func (r *ReservationRepository) ListExpired(ctx context.Context) ([]reservation.Reservation, error) {
	query := `
		SELECT id, book_id, member_id, status, created_at, expires_at, fulfilled_at, cancelled_at
		FROM reservations
		WHERE status='pending' AND expires_at < CURRENT_TIMESTAMP
		ORDER BY created_at ASC
	`
	var reservations []reservation.Reservation
	err := r.db.SelectContext(ctx, &reservations, query)
	if err != nil {
		return nil, err
	}
	return reservations, nil
}
