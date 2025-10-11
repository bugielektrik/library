package handlers

import (
	"fmt"

	"library-service/internal/adapters/http/middleware"
)

// ValidatorAdapter adapts the middleware.Validator to work with generic handlers
type ValidatorAdapter struct {
	validator *middleware.Validator
}

// NewValidatorAdapter creates a new validator adapter
func NewValidatorAdapter(v *middleware.Validator) *ValidatorAdapter {
	return &ValidatorAdapter{validator: v}
}

// CreateValidator creates a validation function for a specific type
func CreateValidator[T any](v *ValidatorAdapter) func(T) error {
	return func(req T) error {
		if v == nil || v.validator == nil {
			return nil
		}

		// Validate the struct
		if err := v.validator.Validate(req); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}

		return nil
	}
}

// NoValidation returns nil validator for handlers that don't need validation
func NoValidation[T any]() func(T) error {
	return nil
}
