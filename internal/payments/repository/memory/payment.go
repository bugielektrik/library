package memory

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/google/uuid"

	"library-service/internal/payments/domain"
)

// PaymentRepository handles CRUD operations for payments in an in-memory store.
type PaymentRepository struct {
	db map[string]domain.Payment
	sync.RWMutex
}

// Compile-time check that PaymentRepository implements domain.Repository
var _ domain.Repository = (*PaymentRepository)(nil)

// NewPaymentRepository creates a new in-memory PaymentRepository.
func NewPaymentRepository() *PaymentRepository {
	return &PaymentRepository{db: make(map[string]domain.Payment)}
}

// Create inserts a new payment into the in-memory store.
func (r *PaymentRepository) Create(ctx context.Context, payment domain.Payment) (string, error) {
	r.Lock()
	defer r.Unlock()

	id := uuid.New().String()
	payment.ID = id
	r.db[id] = payment
	return id, nil
}

// GetByID retrieves a payment by ID from the in-memory store.
func (r *PaymentRepository) GetByID(ctx context.Context, id string) (domain.Payment, error) {
	r.RLock()
	defer r.RUnlock()

	payment, ok := r.db[id]
	if !ok {
		return domain.Payment{}, sql.ErrNoRows
	}
	return payment, nil
}

// GetByInvoiceID retrieves a payment by invoice ID from the in-memory store.
func (r *PaymentRepository) GetByInvoiceID(ctx context.Context, invoiceID string) (domain.Payment, error) {
	r.RLock()
	defer r.RUnlock()

	for _, payment := range r.db {
		if payment.InvoiceID == invoiceID {
			return payment, nil
		}
	}
	return domain.Payment{}, sql.ErrNoRows
}

// Update modifies an existing payment in the in-memory store.
func (r *PaymentRepository) Update(ctx context.Context, id string, payment domain.Payment) error {
	r.Lock()
	defer r.Unlock()

	if _, ok := r.db[id]; !ok {
		return sql.ErrNoRows
	}
	payment.ID = id // Ensure ID remains unchanged
	r.db[id] = payment
	return nil
}

// ListByMemberID retrieves all payments for a specific member.
func (r *PaymentRepository) ListByMemberID(ctx context.Context, memberID string) ([]domain.Payment, error) {
	r.RLock()
	defer r.RUnlock()

	var payments []domain.Payment
	for _, payment := range r.db {
		if payment.MemberID == memberID {
			payments = append(payments, payment)
		}
	}
	return payments, nil
}

// ListByStatus retrieves all payments with a specific status.
func (r *PaymentRepository) ListByStatus(ctx context.Context, status domain.Status) ([]domain.Payment, error) {
	r.RLock()
	defer r.RUnlock()

	var payments []domain.Payment
	for _, payment := range r.db {
		if payment.Status == status {
			payments = append(payments, payment)
		}
	}
	return payments, nil
}

// UpdateStatus updates the status of a payment.
func (r *PaymentRepository) UpdateStatus(ctx context.Context, id string, status domain.Status) error {
	r.Lock()
	defer r.Unlock()

	payment, ok := r.db[id]
	if !ok {
		return sql.ErrNoRows
	}
	payment.Status = status
	r.db[id] = payment
	return nil
}

// ListExpired lists expired pending payments.
func (r *PaymentRepository) ListExpired(ctx context.Context) ([]domain.Payment, error) {
	r.RLock()
	defer r.RUnlock()

	now := time.Now()
	var payments []domain.Payment
	for _, payment := range r.db {
		// Check if payment is pending/processing and expired
		if (payment.Status == domain.StatusPending || payment.Status == domain.StatusProcessing) &&
			payment.ExpiresAt.Before(now) {
			payments = append(payments, payment)
		}
	}
	return payments, nil
}

// ListPendingByMemberID lists pending payments for a member.
func (r *PaymentRepository) ListPendingByMemberID(ctx context.Context, memberID string) ([]domain.Payment, error) {
	r.RLock()
	defer r.RUnlock()

	var payments []domain.Payment
	for _, payment := range r.db {
		if payment.MemberID == memberID && payment.Status == domain.StatusPending {
			payments = append(payments, payment)
		}
	}
	return payments, nil
}
