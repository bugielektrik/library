package errors

import (
	"encoding/json"
	"fmt"
)

// ErrorCode represents the type of error
type ErrorCode string

// Error codes
const (
	CodeNotFound      ErrorCode = "NOT_FOUND"
	CodeAlreadyExists ErrorCode = "ALREADY_EXISTS"
	CodeValidation    ErrorCode = "VALIDATION_ERROR"
	CodeUnauthorized  ErrorCode = "UNAUTHORIZED"
	CodeForbidden     ErrorCode = "FORBIDDEN"
	CodeDatabase      ErrorCode = "DATABASE_ERROR"
	CodeExternal      ErrorCode = "EXTERNAL_ERROR"
	CodeInternal      ErrorCode = "INTERNAL_ERROR"
	CodeTimeout       ErrorCode = "TIMEOUT"
	CodeRateLimit     ErrorCode = "RATE_LIMIT"
	CodeInvalidToken  ErrorCode = "INVALID_TOKEN"
	CodeExpiredToken  ErrorCode = "EXPIRED_TOKEN"
	CodePaymentFailed ErrorCode = "PAYMENT_FAILED"
)

// DomainError represents a domain-specific error
type DomainError struct {
	Code    ErrorCode              `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
	Cause   error                  `json:"-"`
}

// Error implements the error interface
func (e *DomainError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return string(e.Code)
}

// Is implements error comparison for errors.Is
func (e *DomainError) Is(target error) bool {
	t, ok := target.(*DomainError)
	if !ok {
		return false
	}
	return e.Code == t.Code
}

// Unwrap returns the wrapped error
func (e *DomainError) Unwrap() error {
	return e.Cause
}

// MarshalJSON customizes JSON serialization
func (e *DomainError) MarshalJSON() ([]byte, error) {
	type Alias DomainError
	return json.Marshal(&struct {
		*Alias
		Cause string `json:"cause,omitempty"`
	}{
		Alias: (*Alias)(e),
		Cause: getCauseMessage(e.Cause),
	})
}

// getCauseMessage safely extracts the cause message
func getCauseMessage(err error) string {
	if err != nil {
		return err.Error()
	}
	return ""
}

// HTTPStatus returns the appropriate HTTP status code for the error
func (e *DomainError) HTTPStatus() int {
	switch e.Code {
	case CodeNotFound:
		return 404
	case CodeAlreadyExists:
		return 409
	case CodeValidation:
		return 400
	case CodeUnauthorized, CodeInvalidToken, CodeExpiredToken:
		return 401
	case CodeForbidden:
		return 403
	case CodeTimeout:
		return 408
	case CodeRateLimit:
		return 429
	case CodeDatabase, CodeExternal, CodeInternal, CodePaymentFailed:
		return 500
	default:
		return 500
	}
}

// WithDetails adds details to an existing error
func (e *DomainError) WithDetails(key string, value interface{}) *DomainError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// GetDetail retrieves a detail value
func (e *DomainError) GetDetail(key string) (interface{}, bool) {
	if e.Details == nil {
		return nil, false
	}
	val, ok := e.Details[key]
	return val, ok
}

// GetRequestID retrieves the request ID if present
func (e *DomainError) GetRequestID() string {
	if e.Details == nil {
		return ""
	}
	if id, ok := e.Details["request_id"].(string); ok {
		return id
	}
	return ""
}

// Format formats the error for logging
func (e *DomainError) Format() string {
	return fmt.Sprintf("[%s] %s (details: %+v)", e.Code, e.Message, e.Details)
}
