package payment

import (
	"encoding/json"
	"time"
)

// CallbackRetryStatus represents the status of a callback retry
type CallbackRetryStatus string

const (
	CallbackRetryStatusPending    CallbackRetryStatus = "pending"
	CallbackRetryStatusProcessing CallbackRetryStatus = "processing"
	CallbackRetryStatusCompleted  CallbackRetryStatus = "completed"
	CallbackRetryStatusFailed     CallbackRetryStatus = "failed"
)

// CallbackRetry represents a webhook callback retry entry
type CallbackRetry struct {
	ID           string
	PaymentID    string
	CallbackData json.RawMessage
	RetryCount   int
	MaxRetries   int
	LastError    string
	NextRetryAt  *time.Time
	Status       CallbackRetryStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// CallbackRetryRepository defines the interface for callback retry persistence
type CallbackRetryRepository interface {
	Create(callbackRetry *CallbackRetry) error
	GetByID(id string) (*CallbackRetry, error)
	GetPendingRetries(limit int) ([]*CallbackRetry, error)
	Update(callbackRetry *CallbackRetry) error
	Delete(id string) error
}

// CalculateNextRetryTime calculates the next retry time using exponential backoff
// Retry schedule: 1min, 5min, 15min, 1hour, 6hours
func CalculateNextRetryTime(retryCount int) time.Time {
	var delay time.Duration

	switch retryCount {
	case 0:
		delay = 1 * time.Minute
	case 1:
		delay = 5 * time.Minute
	case 2:
		delay = 15 * time.Minute
	case 3:
		delay = 1 * time.Hour
	case 4:
		delay = 6 * time.Hour
	default:
		delay = 24 * time.Hour // Fallback for any additional retries
	}

	return time.Now().Add(delay)
}

// ShouldRetry checks if a callback retry should be attempted
func (c *CallbackRetry) ShouldRetry() bool {
	if c.Status != CallbackRetryStatusPending {
		return false
	}

	if c.RetryCount >= c.MaxRetries {
		return false
	}

	if c.NextRetryAt != nil && time.Now().Before(*c.NextRetryAt) {
		return false
	}

	return true
}

// IncrementRetry increments the retry count and calculates next retry time
func (c *CallbackRetry) IncrementRetry(errorMsg string) {
	c.RetryCount++
	c.LastError = errorMsg
	c.UpdatedAt = time.Now()

	if c.RetryCount >= c.MaxRetries {
		c.Status = CallbackRetryStatusFailed
		c.NextRetryAt = nil
	} else {
		nextRetry := CalculateNextRetryTime(c.RetryCount)
		c.NextRetryAt = &nextRetry
		c.Status = CallbackRetryStatusPending
	}
}

// MarkCompleted marks the callback retry as completed
func (c *CallbackRetry) MarkCompleted() {
	c.Status = CallbackRetryStatusCompleted
	c.UpdatedAt = time.Now()
	c.NextRetryAt = nil
}

// MarkProcessing marks the callback retry as currently processing
func (c *CallbackRetry) MarkProcessing() {
	c.Status = CallbackRetryStatusProcessing
	c.UpdatedAt = time.Now()
}
