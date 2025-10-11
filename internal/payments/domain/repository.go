package domain

import (
	"context"
)

// Repository defines the interface for payment persistence
type Repository interface {
	// Create creates a new payment
	Create(ctx context.Context, payment Payment) (string, error)

	// GetByID retrieves a payment by ID
	GetByID(ctx context.Context, id string) (Payment, error)

	// GetByInvoiceID retrieves a payment by invoice ID
	GetByInvoiceID(ctx context.Context, invoiceID string) (Payment, error)

	// Update updates a payment
	Update(ctx context.Context, id string, payment Payment) error

	// UpdateStatus updates payment status
	UpdateStatus(ctx context.Context, id string, status Status) error

	// ListByMemberID lists payments for a member
	ListByMemberID(ctx context.Context, memberID string) ([]Payment, error)

	// ListByStatus lists payments by status
	ListByStatus(ctx context.Context, status Status) ([]Payment, error)

	// ListExpired lists expired pending payments
	ListExpired(ctx context.Context) ([]Payment, error)

	// ListPendingByMemberID lists pending payments for a member
	ListPendingByMemberID(ctx context.Context, memberID string) ([]Payment, error)
}
