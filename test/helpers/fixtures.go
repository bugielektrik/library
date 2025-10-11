package helpers

import (
	"time"

	"library-service/internal/payments/domain"
)

// Common test IDs
const (
	TestMemberID      = "test-member-123"
	TestAdminID       = "test-admin-123"
	TestBookID        = "test-book-123"
	TestPaymentID     = "test-payment-123"
	TestReservationID = "test-reservation-123"
	TestAuthorID      = "test-author-123"
)

// Common test emails
const (
	TestUserEmail  = "test@example.com"
	TestAdminEmail = "admin@example.com"
)

// ValidISBN returns a valid ISBN-13 for testing
func ValidISBN() string {
	return "978-0-306-40615-7"
}

// ValidISBN10 returns a valid ISBN-10 for testing
func ValidISBN10() string {
	return "0-306-40615-7"
}

// InvalidISBN returns an invalid ISBN for testing
func InvalidISBN() string {
	return "123-invalid-isbn"
}

// ValidPasswordHash returns a bcrypt hash for testing
func ValidPasswordHash() string {
	// Hash of "Password123!"
	return "$2a$10$N9qo8uLOickgx2ZMRZoMye1GZ5dRrmXW9ToRGxhNkFRmQmBfXfXXX"
}

// FutureTime returns a time in the future
func FutureTime(days int) time.Time {
	return time.Now().Add(time.Duration(days) * 24 * time.Hour)
}

// PastTime returns a time in the past
func PastTime(days int) time.Time {
	return time.Now().Add(-time.Duration(days) * 24 * time.Hour)
}

// ValidPaymentStatuses returns all valid payment statuses
func ValidPaymentStatuses() []domain.Status {
	return []domain.Status{
		domain.StatusPending,
		domain.StatusProcessing,
		domain.StatusCompleted,
		domain.StatusFailed,
		domain.StatusCancelled,
		domain.StatusRefunded,
	}
}

// ValidCurrencies returns all valid currencies
func ValidCurrencies() []string {
	return []string{
		domain.CurrencyKZT,
		domain.CurrencyUSD,
		domain.CurrencyEUR,
		domain.CurrencyRUB,
	}
}

// TestAmount returns a standard test payment amount
func TestAmount() int64 {
	return 10000 // 100.00 in smallest unit
}

// StringPtr returns a pointer to a string
func StringPtr(s string) *string {
	return &s
}

// IntPtr returns a pointer to an int
func IntPtr(i int) *int {
	return &i
}

// Int64Ptr returns a pointer to an int64
func Int64Ptr(i int64) *int64 {
	return &i
}

// TimePtr returns a pointer to a time
func TimePtr(t time.Time) *time.Time {
	return &t
}

// BoolPtr returns a pointer to a bool
func BoolPtr(b bool) *bool {
	return &b
}
