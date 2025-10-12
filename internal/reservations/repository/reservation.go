package repository

import (
	"context"
	"fmt"
	"library-service/internal/pkg/repository/postgres"

	"github.com/jmoiron/sqlx"

	reservationdomain "library-service/internal/reservations/domain"
)

// ReservationRepository handles CRUD operations for reservations in a PostgreSQL store.
type ReservationRepository struct {
	postgres.BaseRepository[reservationdomain.Reservation]
}

// Compile-time check that ReservationRepository implements reservationdomain.Repository
var _ reservationdomain.Repository = (*ReservationRepository)(nil)

// NewReservationRepository creates a new ReservationRepository.
func NewReservationRepository(db *sqlx.DB) *ReservationRepository {
	return &ReservationRepository{
		BaseRepository: postgres.NewBaseRepository[reservationdomain.Reservation](db, "reservations"),
	}
}

// Create inserts a new reservation into the store.
func (r *ReservationRepository) Create(ctx context.Context, data reservationdomain.Reservation) (string, error) {
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
	err := r.GetDB().QueryRowContext(ctx, query, args...).Scan(&id)
	return id, postgres.HandleSQLError(err)
}

// GetByID retrieves a reservation by ID from the store (delegates to BaseRepository.Get).
func (r *ReservationRepository) GetByID(ctx context.Context, id string) (reservationdomain.Reservation, error) {
	return r.Get(ctx, id)
}

// GetByMemberID retrieves all reservations for a specific member.
func (r *ReservationRepository) GetByMemberID(ctx context.Context, memberID string) ([]reservationdomain.Reservation, error) {
	query := `
		SELECT id, book_id, member_id, status, created_at, expires_at, fulfilled_at, cancelled_at
		FROM reservations
		WHERE member_id=$1
		ORDER BY created_at DESC
	`
	var reservations []reservationdomain.Reservation
	err := r.GetDB().SelectContext(ctx, &reservations, query, memberID)
	if err != nil {
		return nil, fmt.Errorf("getting reservations by member ID: %w", err)
	}
	return reservations, nil
}

// GetByBookID retrieves all reservations for a specific book.
func (r *ReservationRepository) GetByBookID(ctx context.Context, bookID string) ([]reservationdomain.Reservation, error) {
	query := `
		SELECT id, book_id, member_id, status, created_at, expires_at, fulfilled_at, cancelled_at
		FROM reservations
		WHERE book_id=$1
		ORDER BY created_at ASC
	`
	var reservations []reservationdomain.Reservation
	err := r.GetDB().SelectContext(ctx, &reservations, query, bookID)
	if err != nil {
		return nil, fmt.Errorf("getting reservations by book ID: %w", err)
	}
	return reservations, nil
}

// GetActiveByMemberAndBook retrieves active reservations for a member and book combination.
func (r *ReservationRepository) GetActiveByMemberAndBook(ctx context.Context, memberID, bookID string) ([]reservationdomain.Reservation, error) {
	query := `
		SELECT id, book_id, member_id, status, created_at, expires_at, fulfilled_at, cancelled_at
		FROM reservations
		WHERE member_id=$1 AND book_id=$2 AND status IN ('pending', 'fulfilled')
		ORDER BY created_at ASC
	`
	var reservations []reservationdomain.Reservation
	err := r.GetDB().SelectContext(ctx, &reservations, query, memberID, bookID)
	if err != nil {
		return nil, fmt.Errorf("getting active reservations by member and book: %w", err)
	}
	return reservations, nil
}

// Update modifies an existing reservation in the store.
func (r *ReservationRepository) Update(ctx context.Context, data reservationdomain.Reservation) error {
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
	err := r.GetDB().QueryRowContext(ctx, query, args...).Scan(&id)
	return postgres.HandleSQLError(err)
}

// Delete is inherited from BaseRepository

// ListPending retrieves all pending reservations.
func (r *ReservationRepository) ListPending(ctx context.Context) ([]reservationdomain.Reservation, error) {
	query := `
		SELECT id, book_id, member_id, status, created_at, expires_at, fulfilled_at, cancelled_at
		FROM reservations
		WHERE status='pending'
		ORDER BY created_at ASC
	`
	var reservations []reservationdomain.Reservation
	err := r.GetDB().SelectContext(ctx, &reservations, query)
	if err != nil {
		return nil, fmt.Errorf("listing pending reservations: %w", err)
	}
	return reservations, nil
}

// ListExpired retrieves all expired reservations.
func (r *ReservationRepository) ListExpired(ctx context.Context) ([]reservationdomain.Reservation, error) {
	query := `
		SELECT id, book_id, member_id, status, created_at, expires_at, fulfilled_at, cancelled_at
		FROM reservations
		WHERE status='pending' AND expires_at < CURRENT_TIMESTAMP
		ORDER BY created_at ASC
	`
	var reservations []reservationdomain.Reservation
	err := r.GetDB().SelectContext(ctx, &reservations, query)
	if err != nil {
		return nil, fmt.Errorf("listing expired reservations: %w", err)
	}
	return reservations, nil
}
