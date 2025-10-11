// Package validation provides common validation helpers for use cases.
package validation

import (
	"library-service/pkg/errors"
	"regexp"
	"strings"
)

// RequiredString validates that a string field is not empty
func RequiredString(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return errors.ErrValidation.
			WithDetails("field", fieldName).
			WithDetails("reason", "required")
	}
	return nil
}

// RequiredSlice validates that a slice is not empty
func RequiredSlice[T any](slice []T, fieldName string) error {
	if len(slice) == 0 {
		return errors.ErrValidation.
			WithDetails("field", fieldName).
			WithDetails("reason", "at least one item required")
	}
	return nil
}

// ValidateStringLength validates that a string is within min/max length bounds
func ValidateStringLength(value, fieldName string, min, max int) error {
	length := len(value)
	if length < min {
		return errors.ErrValidation.
			WithDetails("field", fieldName).
			WithDetails("reason", "too short").
			WithDetails("min_length", min).
			WithDetails("actual_length", length)
	}
	if max > 0 && length > max {
		return errors.ErrValidation.
			WithDetails("field", fieldName).
			WithDetails("reason", "too long").
			WithDetails("max_length", max).
			WithDetails("actual_length", length)
	}
	return nil
}

// ValidateEmail validates an email address format
func ValidateEmail(email string) error {
	// Basic email regex pattern
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, _ := regexp.MatchString(emailRegex, email)

	if !matched {
		return errors.ErrValidation.
			WithDetails("field", "email").
			WithDetails("reason", "invalid email format")
	}
	return nil
}

// ValidateRange validates that a numeric value is within a range
func ValidateRange[T int | int64 | float64](value T, fieldName string, min, max T) error {
	if value < min {
		return errors.ErrValidation.
			WithDetails("field", fieldName).
			WithDetails("reason", "below minimum").
			WithDetails("minimum", min).
			WithDetails("actual", value)
	}
	if max > 0 && value > max {
		return errors.ErrValidation.
			WithDetails("field", fieldName).
			WithDetails("reason", "exceeds maximum").
			WithDetails("maximum", max).
			WithDetails("actual", value)
	}
	return nil
}

// ValidateEnum validates that a value is one of the allowed values
func ValidateEnum[T comparable](value T, fieldName string, allowed []T) error {
	for _, v := range allowed {
		if v == value {
			return nil
		}
	}
	return errors.ErrValidation.
		WithDetails("field", fieldName).
		WithDetails("reason", "invalid value").
		WithDetails("allowed_values", allowed).
		WithDetails("actual_value", value)
}

// ValidateSliceItems validates each item in a slice
func ValidateSliceItems[T any](slice []T, fieldName string, validator func(T, int) error) error {
	for i, item := range slice {
		if err := validator(item, i); err != nil {
			return err
		}
	}
	return nil
}

// ValidateConditional validates a field only if a condition is met
func ValidateConditional(condition bool, validator func() error) error {
	if condition {
		return validator()
	}
	return nil
}
