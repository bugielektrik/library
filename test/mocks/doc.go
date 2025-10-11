//go:build integration
// +build integration

// Package mocks provides mock implementations for integration testing.
//
// This package contains manual mock implementations for external services
// and interfaces that are used in integration tests. These mocks allow
// integration tests to run without requiring actual external services.
//
// For repository and cache mocks, use the auto-generated mocks in
// internal/adapters/repository/mocks/ instead.
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
//	    gateway := mocks.NewPaymentGateway()
//	    // Use gateway in test
//	}
package mocks
