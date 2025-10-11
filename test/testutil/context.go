package testutil

import (
	"context"
)

// NewContext creates a new context for testing
// Returns context.Background() which is suitable for most tests
func NewContext() context.Context {
	return context.Background()
}

// NewContextWithValue creates a new context with a key-value pair for testing
func NewContextWithValue(key, value interface{}) context.Context {
	return context.WithValue(context.Background(), key, value)
}
