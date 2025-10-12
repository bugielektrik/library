// Package pkg contains shared utility packages that can be reused across the application.
//
// This directory contains reusable libraries and utilities that are independent
// of the application's business logic. These packages could potentially be
// extracted into separate modules.
//
// Subpackages:
//   - errors: Custom error types and error handling utilities
//   - validator: Input validation with custom rules
//   - pagination: Pagination helpers for list operations
//   - crypto: Cryptography utilities (hashing, encryption)
//   - timeutil: Time manipulation and formatting helpers
//
// Design principles:
//   - Self-contained and reusable
//   - No dependencies on application-specific code
//   - Well-tested and documented
//   - Follow single responsibility principle
package pkg
