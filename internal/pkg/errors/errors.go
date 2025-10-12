package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Error represents a domain error with additional context
type Error struct {
	Code       string                 `json:"code"`
	Message    string                 `json:"message"`
	HTTPStatus int                    `json:"-"`
	Err        error                  `json:"-"`
	Details    map[string]interface{} `json:"details,omitempty"`
}

// Error implements the error interface
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap implements the unwrap interface for error chaining
func (e *Error) Unwrap() error {
	return e.Err
}

// Is implements error comparison for errors.Is
func (e *Error) Is(target error) bool {
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// WithDetails adds contextual details to the error
func (e *Error) WithDetails(key string, value interface{}) *Error {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// Wrap wraps an underlying error with this domain error
func (e *Error) Wrap(err error) *Error {
	return &Error{
		Code:       e.Code,
		Message:    e.Message,
		HTTPStatus: e.HTTPStatus,
		Err:        err,
		Details:    e.Details,
	}
}

// Common domain errors
var (
	// Validation errors
	ErrValidation = &Error{
		Code:       "VALIDATION_ERROR",
		Message:    "Validation failed",
		HTTPStatus: http.StatusBadRequest,
	}

	ErrInvalidInput = &Error{
		Code:       "INVALID_INPUT",
		Message:    "Invalid input provided",
		HTTPStatus: http.StatusBadRequest,
	}

	// Resource errors
	ErrNotFound = &Error{
		Code:       "NOT_FOUND",
		Message:    "Resource not found",
		HTTPStatus: http.StatusNotFound,
	}

	ErrAlreadyExists = &Error{
		Code:       "ALREADY_EXISTS",
		Message:    "Resource already exists",
		HTTPStatus: http.StatusConflict,
	}

	// Authorization errors
	ErrUnauthorized = &Error{
		Code:       "UNAUTHORIZED",
		Message:    "Authentication required",
		HTTPStatus: http.StatusUnauthorized,
	}

	ErrForbidden = &Error{
		Code:       "FORBIDDEN",
		Message:    "Access forbidden",
		HTTPStatus: http.StatusForbidden,
	}

	// System errors
	ErrInternal = &Error{
		Code:       "INTERNAL_ERROR",
		Message:    "Internal server error",
		HTTPStatus: http.StatusInternalServerError,
	}

	ErrDatabase = &Error{
		Code:       "DATABASE_ERROR",
		Message:    "Database operation failed",
		HTTPStatus: http.StatusInternalServerError,
	}

	ErrCache = &Error{
		Code:       "CACHE_ERROR",
		Message:    "Cache operation failed",
		HTTPStatus: http.StatusInternalServerError,
	}

	// Business logic errors
	ErrBusinessRule = &Error{
		Code:       "BUSINESS_RULE_VIOLATION",
		Message:    "Business rule violation",
		HTTPStatus: http.StatusUnprocessableEntity,
	}

	// Authentication errors
	ErrInvalidCredentials = &Error{
		Code:       "INVALID_CREDENTIALS",
		Message:    "Invalid email or password",
		HTTPStatus: http.StatusUnauthorized,
	}

	ErrInvalidToken = &Error{
		Code:       "INVALID_TOKEN",
		Message:    "Invalid or expired token",
		HTTPStatus: http.StatusUnauthorized,
	}
)

// New creates a new domain error
func New(code, message string, httpStatus int) *Error {
	return &Error{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// Is checks if the target error matches this error type
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As finds the first error in err's chain that matches target
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}

// GetHTTPStatus extracts HTTP status from error or returns 500
func GetHTTPStatus(err error) int {
	var domainErr *Error
	if errors.As(err, &domainErr) {
		return domainErr.HTTPStatus
	}
	return http.StatusInternalServerError
}
