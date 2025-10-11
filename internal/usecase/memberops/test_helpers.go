package memberops

// This file contains test helpers for member operations.
// The mock repository has been moved to internal/adapters/repository/mocks/MockMemberRepository

import (
	"library-service/test/builders"
	"library-service/test/helpers"
)

// Common test constants
const (
	TestMemberID = helpers.TestMemberID
	TestEmail    = helpers.TestUserEmail
)

// CreateTestMember creates a test member using the builder
var CreateTestMember = builders.Member

// Helper to create string pointers (for backward compatibility)
func strPtr(s string) *string {
	return &s
}
