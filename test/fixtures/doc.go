// Package fixtures provides test data fixtures for all domain entities.
//
// This package contains factory functions that create sample data for testing purposes.
// Each entity has multiple fixture functions to cover different test scenarios:
//
// - Plural functions (Books(), Authors(), etc.) return collections of sample data
// - Singular functions (Book(), Author(), etc.) return a single default sample
// - Specialized functions (PendingReservation(), CompletedPayment(), etc.) return specific states
// - ForCreate functions return entities ready for creation (without IDs)
// - Update functions return partial data for update operations
//
// Usage:
//
//	import "library-service/test/fixtures"
//
//	// Get a single book
//	book := fixtures.Book()
//
//	// Get a collection of authors
//	authors := fixtures.Authors()
//
//	// Get a pending reservation
//	reservation := fixtures.PendingReservation()
//
//	// Create custom book data
//	book := fixtures.BookForCreate()
package fixtures
