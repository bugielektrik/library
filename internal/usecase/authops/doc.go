// Package authops implements authentication and authorization use cases.
//
// This package handles member authentication workflows including registration,
// login, token refresh, and token validation. It coordinates JWT token
// generation, password hashing, and member repository operations.
//
// Use cases implemented:
//   - RegisterMemberUseCase: Creates new member account with password hashing
//   - LoginMemberUseCase: Authenticates member and issues JWT tokens
//   - RefreshTokenUseCase: Issues new access token from valid refresh token
//   - ValidateTokenUseCase: Validates JWT token and retrieves member information
//
// Dependencies:
//   - member.Repository: For member account persistence
//   - auth.TokenService: For JWT token generation and validation
//   - auth.PasswordService: For secure password hashing and verification
//
// Example usage:
//
//	loginUC := authops.NewLoginMemberUseCase(memberRepo, tokenService, passwordService)
//	response, err := loginUC.Execute(ctx, authops.LoginRequest{
//	    Email:    "user@example.com",
//	    Password: "SecurePassword123",
//	})
//	// response contains: AccessToken, RefreshToken, Member info
//
// Security features:
//   - Passwords hashed with bcrypt (cost factor 10)
//   - JWT tokens with configurable expiry (default: access 24h, refresh 7d)
//   - Email uniqueness enforced at repository level
//   - Failed login attempts logged for security monitoring
//
// Token structure:
//   - Access token: Short-lived (24h), used for API authorization
//   - Refresh token: Long-lived (7d), used to obtain new access tokens
//   - Both tokens contain: member_id, email, role claims
//
// Architecture:
//   - Package name uses "ops" suffix to avoid naming conflicts
//   - Token services injected from infrastructure layer
//   - Password validation rules enforced by domain member.Service
package authops
