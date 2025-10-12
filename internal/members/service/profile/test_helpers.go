package profile

// This file contains test helpers for member service.
// The mock repository has been moved to internal/adapters/repository/mocks/MockMemberRepository

import (
	"library-service/test/helpers"
)

// Common test constants
const (
	TestMemberID = helpers.TestMemberID
	TestEmail    = helpers.TestUserEmail
)

// Helper to create string pointers (for backward compatibility)
func strPtr(s string) *string {
	return &s
}
