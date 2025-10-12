package memory

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/google/uuid"

	"library-service/internal/reservations/domain"
)

// ReservationRepository handles CRUD operations for reservations in an in-memory store.
type ReservationRepository struct {
	db map[string]domain.Reservation
	sync.RWMutex
}

// Compile-time check that ReservationRepository implements domain.Repository
var _ domain.Repository = (*ReservationRepository)(nil)

// NewReservationRepository creates a new in-memory ReservationRepository.
func NewReservationRepository() *ReservationRepository {
	return &ReservationRepository{db: make(map[string]domain.Reservation)}
}

// Create inserts a new reservation into the in-memory store.
func (r *ReservationRepository) Create(ctx context.Context, reservation domain.Reservation) (string, error) {
	r.Lock()
	defer r.Unlock()

	id := uuid.New().String()
	reservation.ID = id
	r.db[id] = reservation
	return id, nil
}

// GetByID retrieves a reservation by ID from the in-memory store.
func (r *ReservationRepository) GetByID(ctx context.Context, id string) (domain.Reservation, error) {
	r.RLock()
	defer r.RUnlock()

	reservation, ok := r.db[id]
	if !ok {
		return domain.Reservation{}, sql.ErrNoRows
	}
	return reservation, nil
}

// GetByMemberID retrieves all reservations for a specific member.
func (r *ReservationRepository) GetByMemberID(ctx context.Context, memberID string) ([]domain.Reservation, error) {
	r.RLock()
	defer r.RUnlock()

	var reservations []domain.Reservation
	for _, reservation := range r.db {
		if reservation.MemberID == memberID {
			reservations = append(reservations, reservation)
		}
	}
	return reservations, nil
}

// GetByBookID retrieves all reservations for a specific book.
func (r *ReservationRepository) GetByBookID(ctx context.Context, bookID string) ([]domain.Reservation, error) {
	r.RLock()
	defer r.RUnlock()

	var reservations []domain.Reservation
	for _, reservation := range r.db {
		if reservation.BookID == bookID {
			reservations = append(reservations, reservation)
		}
	}
	return reservations, nil
}

// GetActiveByMemberAndBook retrieves active reservations for a member and book combination.
func (r *ReservationRepository) GetActiveByMemberAndBook(ctx context.Context, memberID, bookID string) ([]domain.Reservation, error) {
	r.RLock()
	defer r.RUnlock()

	var reservations []domain.Reservation
	for _, reservation := range r.db {
		if reservation.MemberID == memberID && reservation.BookID == bookID {
			// Active means pending or fulfilled status
			if reservation.Status == domain.StatusPending || reservation.Status == domain.StatusFulfilled {
				reservations = append(reservations, reservation)
			}
		}
	}
	return reservations, nil
}

// Update modifies an existing reservation in the in-memory store.
func (r *ReservationRepository) Update(ctx context.Context, reservation domain.Reservation) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.db[reservation.ID]; !ok {
		return sql.ErrNoRows
	}
	r.db[reservation.ID] = reservation
	return nil
}

// Delete removes a reservation by ID from the in-memory store.
func (r *ReservationRepository) Delete(ctx context.Context, id string) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.db[id]; !ok {
		return sql.ErrNoRows
	}
	delete(r.db, id)
	return nil
}

// ListPending retrieves all pending reservations.
func (r *ReservationRepository) ListPending(ctx context.Context) ([]domain.Reservation, error) {
	r.RLock()
	defer r.RUnlock()

	var reservations []domain.Reservation
	for _, reservation := range r.db {
		if reservation.Status == domain.StatusPending {
			reservations = append(reservations, reservation)
		}
	}
	return reservations, nil
}

// ListExpired retrieves all expired reservations.
func (r *ReservationRepository) ListExpired(ctx context.Context) ([]domain.Reservation, error) {
	r.RLock()
	defer r.RUnlock()

	now := time.Now()
	var reservations []domain.Reservation
	for _, reservation := range r.db {
		// Expired means pending status and past expiration time
		if reservation.Status == domain.StatusPending && reservation.ExpiresAt.Before(now) {
			reservations = append(reservations, reservation)
		}
	}
	return reservations, nil
}
