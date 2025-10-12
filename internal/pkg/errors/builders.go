package errors

import (
	"fmt"
	"runtime"
	"strings"
)

// ErrorBuilder provides a fluent interface for building errors
type ErrorBuilder struct {
	err        *DomainError
	stackTrace string
}

// NewError creates a new error builder with a code
func NewError(code ErrorCode) *ErrorBuilder {
	// Capture stack trace at error creation point
	_, file, line, _ := runtime.Caller(1)
	stackTrace := fmt.Sprintf("%s:%d", file, line)

	return &ErrorBuilder{
		err: &DomainError{
			Code:    code,
			Details: make(map[string]interface{}),
		},
		stackTrace: stackTrace,
	}
}

// WithMessage sets the error message
func (b *ErrorBuilder) WithMessage(msg string) *ErrorBuilder {
	b.err.Message = msg
	return b
}

// WithMessagef sets a formatted error message
func (b *ErrorBuilder) WithMessagef(format string, args ...interface{}) *ErrorBuilder {
	b.err.Message = fmt.Sprintf(format, args...)
	return b
}

// WithDetail adds a single detail to the error
func (b *ErrorBuilder) WithDetail(key string, value interface{}) *ErrorBuilder {
	b.err.Details[key] = value
	return b
}

// WithDetails adds multiple details to the error
func (b *ErrorBuilder) WithDetails(details map[string]interface{}) *ErrorBuilder {
	for k, v := range details {
		b.err.Details[k] = v
	}
	return b
}

// WithCause wraps another error
func (b *ErrorBuilder) WithCause(cause error) *ErrorBuilder {
	if cause != nil {
		b.err.Details["cause"] = cause.Error()
		b.err.Cause = cause
	}
	return b
}

// WithStack adds the stack trace to error details
func (b *ErrorBuilder) WithStack() *ErrorBuilder {
	b.err.Details["stack"] = b.stackTrace
	return b
}

// WithRequestID adds request ID for correlation
func (b *ErrorBuilder) WithRequestID(requestID string) *ErrorBuilder {
	if requestID != "" {
		b.err.Details["request_id"] = requestID
	}
	return b
}

// WithField adds field-specific error information
func (b *ErrorBuilder) WithField(fieldName, reason string) *ErrorBuilder {
	b.err.Details["field"] = fieldName
	b.err.Details["reason"] = reason
	return b
}

// Build returns the constructed error
func (b *ErrorBuilder) Build() error {
	// Set default message if not provided
	if b.err.Message == "" {
		b.err.Message = getDefaultMessage(b.err.Code)
	}
	return b.err
}

// Common error constructors for convenience

// NotFound creates a not found error
func NotFound(entity string) error {
	return NewError(CodeNotFound).
		WithMessagef("%s not found", strings.Title(entity)).
		WithDetail("entity", entity).
		Build()
}

// NotFoundWithID creates a not found error with ID
func NotFoundWithID(entity, id string) error {
	return NewError(CodeNotFound).
		WithMessagef("%s with ID '%s' not found", strings.Title(entity), id).
		WithDetail("entity", entity).
		WithDetail("id", id).
		Build()
}

// AlreadyExists creates an already exists error
func AlreadyExists(entity, field, value string) error {
	return NewError(CodeAlreadyExists).
		WithMessagef("%s with %s '%s' already exists", strings.Title(entity), field, value).
		WithDetail("entity", entity).
		WithDetail("field", field).
		WithDetail("value", value).
		Build()
}

// Validation creates a validation error
func Validation(field, reason string) error {
	return NewError(CodeValidation).
		WithMessagef("Validation failed for field '%s': %s", field, reason).
		WithField(field, reason).
		Build()
}

// ValidationRequired creates a required field validation error
func ValidationRequired(field string) error {
	return Validation(field, "required")
}

// ValidationInvalid creates an invalid value validation error
func ValidationInvalid(field string, value interface{}) error {
	return NewError(CodeValidation).
		WithMessagef("Invalid value for field '%s'", field).
		WithDetail("field", field).
		WithDetail("value", value).
		WithDetail("reason", "invalid").
		Build()
}

// ValidationRange creates a range validation error
func ValidationRange(field string, min, max interface{}) error {
	return NewError(CodeValidation).
		WithMessagef("Value for field '%s' must be between %v and %v", field, min, max).
		WithDetail("field", field).
		WithDetail("min", min).
		WithDetail("max", max).
		WithDetail("reason", "out_of_range").
		Build()
}

// Unauthorized creates an unauthorized error
func Unauthorized(reason string) error {
	return NewError(CodeUnauthorized).
		WithMessage("Unauthorized").
		WithDetail("reason", reason).
		Build()
}

// Forbidden creates a forbidden error
func Forbidden(action, resource string) error {
	return NewError(CodeForbidden).
		WithMessagef("Forbidden: cannot %s %s", action, resource).
		WithDetail("action", action).
		WithDetail("resource", resource).
		Build()
}

// Database creates a database error
func Database(operation string, cause error) error {
	return NewError(CodeDatabase).
		WithMessagef("Database operation failed: %s", operation).
		WithDetail("operation", operation).
		WithCause(cause).
		Build()
}

// External creates an external service error
func External(service string, cause error) error {
	return NewError(CodeExternal).
		WithMessagef("External service error: %s", service).
		WithDetail("service", service).
		WithCause(cause).
		Build()
}

// Internal creates an internal server error
func Internal(message string, cause error) error {
	return NewError(CodeInternal).
		WithMessage(message).
		WithCause(cause).
		WithStack().
		Build()
}

// getDefaultMessage returns a default message for an error code
func getDefaultMessage(code ErrorCode) string {
	switch code {
	case CodeNotFound:
		return "Resource not found"
	case CodeAlreadyExists:
		return "Resource already exists"
	case CodeValidation:
		return "Validation failed"
	case CodeUnauthorized:
		return "Unauthorized"
	case CodeForbidden:
		return "Forbidden"
	case CodeDatabase:
		return "Database error"
	case CodeExternal:
		return "External service error"
	case CodeInternal:
		return "Internal server error"
	default:
		return "An error occurred"
	}
}
