//go:build integration
// +build integration

// Package mocks provides mock implementations for integration testing.
//
// This package contains manual mock implementations for external service
// and interfaces that are used in integration tests. These mocks allow
// integration tests to run without requiring actual external service.
//
// For repository mocks, use the auto-generated mocks in each bounded context:
//   - Books: internal/books/repository/mocks/
//   - Members: internal/members/repository/mocks/
//   - Payments: internal/payments/repository/mocks/
//   - Reservations: internal/reservations/repository/mocks/
//
// Usage:
//
//	//go:build integration
//	// +build integration
//
//	package mytest
//
//	import "library-service/test/mocks"
//
//	func TestWithMock(t *testing.T) {
//	    provider := mocks.NewPaymentGateway()
//	    // Use provider in test
//	}
package mocks
