// Package auth provides authentication and authorization services.
//
// This package contains infrastructure services for:
//   - JWT token generation and validation
//   - Password hashing and verification (bcrypt)
//   - Email validation
//   - Token claims management
//
// Key components:
//   - JWTService: Handles JWT token operations (access and refresh tokens)
//   - PasswordService: Securely hashes and verifies passwords
//   - Claims: JWT token payload structure
//
// These services are used by authentication use cases but are independent
// of business logic, focusing on security and cryptographic operations.
package auth
