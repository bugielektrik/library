package errors_test

import (
	"fmt"
	errors2 "library-service/internal/pkg/errors"
)

// Example demonstrates basic error usage
func Example() {
	// Create a simple error
	err := errors2.ErrNotFound
	fmt.Println(err.Code)
	fmt.Println(err.Message)
	// Output:
	// NOT_FOUND
	// Resource not found
}

// Example_withDetails demonstrates adding context to errors
func Example_withDetails() {
	// Add single detail
	err := errors2.ErrNotFound.WithDetails("book_id", "123")
	fmt.Println(err.Code)
	fmt.Println(err.Details["book_id"])
	// Output:
	// NOT_FOUND
	// 123
}

// Example_chainedDetails demonstrates chaining multiple details
func Example_chainedDetails() {
	// Chain multiple details
	err := errors2.ErrValidation.
		WithDetails("field", "email").
		WithDetails("reason", "invalid format").
		WithDetails("value", "not-an-email")

	fmt.Println(err.Code)
	fmt.Println(err.Details["field"])
	fmt.Println(err.Details["reason"])
	// Output:
	// VALIDATION_ERROR
	// email
	// invalid format
}

// Example_wrap demonstrates wrapping underlying errors
func Example_wrap() {
	// Wrap a standard error
	dbErr := fmt.Errorf("connection timeout")
	err := errors2.ErrDatabase.Wrap(dbErr)

	fmt.Println(err.Code)
	fmt.Printf("%v\n", err.Error())
	// Output:
	// DATABASE_ERROR
	// Database operation failed: connection timeout
}

// Example_validationError demonstrates validation error with context
func Example_validationError() {
	// Typical validation error
	err := errors2.ErrValidation.
		WithDetails("field", "amount").
		WithDetails("reason", "below minimum").
		WithDetails("minimum", 100).
		WithDetails("actual", 50)

	fmt.Println(err.Code)
	fmt.Println(err.Details["field"])
	fmt.Println(err.Details["minimum"])
	// Output:
	// VALIDATION_ERROR
	// amount
	// 100
}

// Example_notFoundError demonstrates not found error with ID
func Example_notFoundError() {
	bookID := "book-123"
	err := errors2.ErrNotFound.WithDetails("book_id", bookID)

	fmt.Println(err.Code)
	fmt.Println(err.Details["book_id"])
	// Output:
	// NOT_FOUND
	// book-123
}

// Example_businessRuleError demonstrates business rule violation
func Example_businessRuleError() {
	err := errors2.ErrBusinessRule.
		WithDetails("rule", "max_active_reservations").
		WithDetails("limit", 5).
		WithDetails("current", 5)

	fmt.Println(err.Code)
	fmt.Println(err.Details["rule"])
	// Output:
	// BUSINESS_RULE_VIOLATION
	// max_active_reservations
}

// Example_unauthorizedError demonstrates authorization error
func Example_unauthorizedError() {
	err := errors2.ErrForbidden.
		WithDetails("resource", "payment").
		WithDetails("required_role", "admin").
		WithDetails("actual_role", "user")

	fmt.Println(err.Code)
	fmt.Println(err.Details["resource"])
	// Output:
	// FORBIDDEN
	// payment
}

// Example_httpStatus demonstrates extracting HTTP status from error
func Example_httpStatus() {
	err := errors2.ErrNotFound
	status := errors2.GetHTTPStatus(err)
	fmt.Println(status)

	err2 := errors2.ErrValidation
	status2 := errors2.GetHTTPStatus(err2)
	fmt.Println(status2)

	// Output:
	// 404
	// 400
}

// Example_errorComparison demonstrates error comparison with errors.Is
func Example_errorComparison() {
	// Create an error with details
	err := errors2.ErrValidation.WithDetails("field", "email")

	// Compare with base error
	isValidation := errors2.Is(err, errors2.ErrValidation)
	isNotFound := errors2.Is(err, errors2.ErrNotFound)

	fmt.Println(isValidation)
	fmt.Println(isNotFound)
	// Output:
	// true
	// false
}

// Example_domainSpecificError demonstrates using domain-specific errors
func Example_domainSpecificError() {
	// Book not found
	err := errors2.ErrBookNotFound.WithDetails("book_id", "book-456")
	fmt.Println(err.Code)
	fmt.Println(err.Message)

	// Invalid ISBN
	err2 := errors2.ErrInvalidISBN.
		WithDetails("isbn", "invalid-isbn").
		WithDetails("reason", "invalid checksum")
	fmt.Println(err2.Code)
	fmt.Println(err2.Details["isbn"])

	// Output:
	// BOOK_NOT_FOUND
	// Book not found
	// INVALID_ISBN
	// invalid-isbn
}

// Example_multipleDetails demonstrates comprehensive error context
func Example_multipleDetails() {
	err := errors2.ErrValidation.
		WithDetails("field", "saved_card_id").
		WithDetails("reason", "card is inactive or expired").
		WithDetails("is_active", false).
		WithDetails("is_expired", true).
		WithDetails("expiry_month", 12).
		WithDetails("expiry_year", 2023)

	fmt.Println(err.Code)
	fmt.Println(err.Details["field"])
	fmt.Println(err.Details["reason"])
	fmt.Println(err.Details["is_expired"])
	// Output:
	// VALIDATION_ERROR
	// saved_card_id
	// card is inactive or expired
	// true
}
