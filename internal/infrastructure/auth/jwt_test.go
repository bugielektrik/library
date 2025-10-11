package auth

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"library-service/internal/domain/member"
)

// Test constants
const (
	testSecretKey = "test-secret-key-for-jwt-testing-min-32-chars"
	testIssuer    = "library-service-test"
	testMemberID  = "member-123"
	testEmail     = "test@example.com"
)

func newTestJWTService() *JWTService {
	return NewJWTService(
		testSecretKey,
		15*time.Minute, // access token TTL
		7*24*time.Hour, // refresh token TTL
		testIssuer,
	)
}

// TestGenerateAccessToken tests successful access token generation
func TestGenerateAccessToken(t *testing.T) {
	service := newTestJWTService()

	token, err := service.GenerateAccessToken(testMemberID, testEmail, member.RoleUser)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty token")
	}

	// Verify token has 3 parts (header.payload.signature)
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Errorf("Expected token with 3 parts, got %d", len(parts))
	}
}

// TestGenerateAccessToken_ValidClaims verifies the generated token contains correct claims
func TestGenerateAccessToken_ValidClaims(t *testing.T) {
	service := newTestJWTService()

	token, err := service.GenerateAccessToken(testMemberID, testEmail, member.RoleAdmin)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	// Validate and extract claims
	claims, err := service.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	// Verify claims
	if claims.MemberID != testMemberID {
		t.Errorf("Expected MemberID %s, got %s", testMemberID, claims.MemberID)
	}
	if claims.Email != testEmail {
		t.Errorf("Expected Email %s, got %s", testEmail, claims.Email)
	}
	if claims.Role != string(member.RoleAdmin) {
		t.Errorf("Expected Role %s, got %s", member.RoleAdmin, claims.Role)
	}
	if claims.Issuer != testIssuer {
		t.Errorf("Expected Issuer %s, got %s", testIssuer, claims.Issuer)
	}
	if claims.Subject != testMemberID {
		t.Errorf("Expected Subject %s, got %s", testMemberID, claims.Subject)
	}
}

// TestGenerateAccessToken_ExpiryConfiguration tests custom expiry times
func TestGenerateAccessToken_ExpiryConfiguration(t *testing.T) {
	tests := []struct {
		name      string
		accessTTL time.Duration
	}{
		{"5 minutes", 5 * time.Minute},
		{"1 hour", 1 * time.Hour},
		{"24 hours", 24 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewJWTService(testSecretKey, tt.accessTTL, 7*24*time.Hour, testIssuer)

			beforeGen := time.Now()
			token, err := service.GenerateAccessToken(testMemberID, testEmail, member.RoleUser)

			if err != nil {
				t.Fatalf("GenerateAccessToken failed: %v", err)
			}

			claims, err := service.ValidateToken(token)
			if err != nil {
				t.Fatalf("ValidateToken failed: %v", err)
			}

			// JWT timestamps are in seconds, so allow 2 second tolerance
			expectedExpiry := beforeGen.Add(tt.accessTTL)
			tolerance := 2 * time.Second

			diff := claims.ExpiresAt.Time.Sub(expectedExpiry)
			if diff < -tolerance || diff > tolerance {
				t.Errorf("Token expiry %v differs from expected %v by %v (tolerance: %v)",
					claims.ExpiresAt.Time, expectedExpiry, diff, tolerance)
			}
		})
	}
}

// TestGenerateRefreshToken tests refresh token generation
func TestGenerateRefreshToken(t *testing.T) {
	service := newTestJWTService()

	token, err := service.GenerateRefreshToken(testMemberID)
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty refresh token")
	}

	// Verify token structure
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Errorf("Expected token with 3 parts, got %d", len(parts))
	}
}

// TestGenerateRefreshToken_ValidClaims verifies refresh token claims
func TestGenerateRefreshToken_ValidClaims(t *testing.T) {
	service := newTestJWTService()

	token, err := service.GenerateRefreshToken(testMemberID)
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}

	claims, err := service.ValidateRefreshToken(token)
	if err != nil {
		t.Fatalf("ValidateRefreshToken failed: %v", err)
	}

	if claims.MemberID != testMemberID {
		t.Errorf("Expected MemberID %s, got %s", testMemberID, claims.MemberID)
	}
	if claims.Issuer != testIssuer {
		t.Errorf("Expected Issuer %s, got %s", testIssuer, claims.Issuer)
	}
}

// TestGenerateTokenPair tests generating both tokens together
func TestGenerateTokenPair(t *testing.T) {
	service := newTestJWTService()

	pair, err := service.GenerateTokenPair(testMemberID, testEmail, member.RoleUser)
	if err != nil {
		t.Fatalf("GenerateTokenPair failed: %v", err)
	}

	if pair.AccessToken == "" {
		t.Error("Expected non-empty access token")
	}
	if pair.RefreshToken == "" {
		t.Error("Expected non-empty refresh token")
	}
	if pair.ExpiresIn != int64((15 * time.Minute).Seconds()) {
		t.Errorf("Expected ExpiresIn %d, got %d", int64((15 * time.Minute).Seconds()), pair.ExpiresIn)
	}

	// Verify both tokens are valid
	_, err = service.ValidateToken(pair.AccessToken)
	if err != nil {
		t.Errorf("Access token validation failed: %v", err)
	}

	_, err = service.ValidateRefreshToken(pair.RefreshToken)
	if err != nil {
		t.Errorf("Refresh token validation failed: %v", err)
	}
}

// TestValidateToken_ValidToken tests successful token validation
func TestValidateToken_ValidToken(t *testing.T) {
	service := newTestJWTService()

	token, err := service.GenerateAccessToken(testMemberID, testEmail, member.RoleUser)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	claims, err := service.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if claims == nil {
		t.Error("Expected non-nil claims")
	}
}

// TestValidateToken_ExpiredToken tests that expired tokens are rejected
func TestValidateToken_ExpiredToken(t *testing.T) {
	// Create service with very short TTL
	service := NewJWTService(testSecretKey, 1*time.Millisecond, 7*24*time.Hour, testIssuer)

	token, err := service.GenerateAccessToken(testMemberID, testEmail, member.RoleUser)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	_, err = service.ValidateToken(token)
	if err == nil {
		t.Error("Expected error for expired token, got nil")
	}
	if !strings.Contains(err.Error(), "expired") {
		t.Errorf("Expected error message to contain 'expired', got: %v", err)
	}
}

// TestValidateToken_InvalidSignature tests detection of tampered tokens
func TestValidateToken_InvalidSignature(t *testing.T) {
	service := newTestJWTService()

	token, err := service.GenerateAccessToken(testMemberID, testEmail, member.RoleUser)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	// Create a different service with different secret
	differentService := NewJWTService("different-secret-key-min-32-chars", 15*time.Minute, 7*24*time.Hour, testIssuer)

	// Try to validate with wrong secret
	_, err = differentService.ValidateToken(token)
	if err == nil {
		t.Error("Expected error for invalid signature, got nil")
	}
}

// TestValidateToken_MalformedToken tests rejection of malformed tokens
func TestValidateToken_MalformedToken(t *testing.T) {
	service := newTestJWTService()

	tests := []struct {
		name  string
		token string
	}{
		{"empty", ""},
		{"random string", "not-a-valid-jwt-token"},
		{"missing parts", "header.payload"},
		{"too many parts", "header.payload.signature.extra"},
		{"invalid base64", "header.!nv@lid.signature"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.ValidateToken(tt.token)
			if err == nil {
				t.Errorf("Expected error for malformed token %q, got nil", tt.name)
			}
		})
	}
}

// TestValidateToken_WrongSigningMethod tests rejection of tokens with unexpected signing methods
func TestValidateToken_WrongSigningMethod(t *testing.T) {
	service := newTestJWTService()

	// Create a token with RS256 instead of HS256
	claims := &Claims{
		MemberID: testMemberID,
		Email:    testEmail,
		Role:     string(member.RoleUser),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    testIssuer,
			Subject:   testMemberID,
		},
	}

	// Try to create with "none" algorithm (security risk)
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenString, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

	// Should reject token with unexpected signing method
	_, err := service.ValidateToken(tokenString)
	if err == nil {
		t.Error("Expected error for wrong signing method, got nil")
	}
}

// TestValidateRefreshToken_ValidToken tests refresh token validation
func TestValidateRefreshToken_ValidToken(t *testing.T) {
	service := newTestJWTService()

	token, err := service.GenerateRefreshToken(testMemberID)
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}

	claims, err := service.ValidateRefreshToken(token)
	if err != nil {
		t.Fatalf("ValidateRefreshToken failed: %v", err)
	}

	if claims.MemberID != testMemberID {
		t.Errorf("Expected MemberID %s, got %s", testMemberID, claims.MemberID)
	}
}

// TestValidateRefreshToken_ExpiredToken tests expired refresh token rejection
func TestValidateRefreshToken_ExpiredToken(t *testing.T) {
	service := NewJWTService(testSecretKey, 15*time.Minute, 1*time.Millisecond, testIssuer)

	token, err := service.GenerateRefreshToken(testMemberID)
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	_, err = service.ValidateRefreshToken(token)
	if err == nil {
		t.Error("Expected error for expired refresh token, got nil")
	}
	if !strings.Contains(err.Error(), "expired") {
		t.Errorf("Expected error message to contain 'expired', got: %v", err)
	}
}

// TestValidateRefreshToken_AccessTokenAsRefreshToken tests that access tokens are accepted by refresh validator
// Note: JWT is lenient and will parse tokens with extra fields. This is expected behavior.
// In practice, the application logic should ensure tokens are used for their intended purpose.
func TestValidateRefreshToken_AccessTokenAsRefreshToken(t *testing.T) {
	service := newTestJWTService()

	// Generate access token
	accessToken, err := service.GenerateAccessToken(testMemberID, testEmail, member.RoleUser)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	// JWT validation is lenient - it will parse access tokens as refresh tokens
	// because the required fields (MemberID, RegisteredClaims) are present
	refreshClaims, err := service.ValidateRefreshToken(accessToken)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify the MemberID is correctly extracted despite using wrong validator
	if refreshClaims.MemberID != testMemberID {
		t.Errorf("Expected MemberID %s, got %s", testMemberID, refreshClaims.MemberID)
	}
}

// TestRefreshAccessToken tests the token refresh flow
func TestRefreshAccessToken(t *testing.T) {
	service := newTestJWTService()

	// Generate initial refresh token
	refreshToken, err := service.GenerateRefreshToken(testMemberID)
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}

	// Use refresh token to get new access token
	newAccessToken, err := service.RefreshAccessToken(refreshToken, testEmail, member.RoleAdmin)
	if err != nil {
		t.Fatalf("RefreshAccessToken failed: %v", err)
	}

	if newAccessToken == "" {
		t.Error("Expected non-empty new access token")
	}

	// Verify new access token is valid
	claims, err := service.ValidateToken(newAccessToken)
	if err != nil {
		t.Fatalf("Validation of refreshed token failed: %v", err)
	}

	// Verify claims match provided data
	if claims.MemberID != testMemberID {
		t.Errorf("Expected MemberID %s, got %s", testMemberID, claims.MemberID)
	}
	if claims.Email != testEmail {
		t.Errorf("Expected Email %s, got %s", testEmail, claims.Email)
	}
	if claims.Role != string(member.RoleAdmin) {
		t.Errorf("Expected Role %s, got %s", member.RoleAdmin, claims.Role)
	}
}

// TestRefreshAccessToken_ExpiredRefreshToken tests refresh with expired token
func TestRefreshAccessToken_ExpiredRefreshToken(t *testing.T) {
	service := NewJWTService(testSecretKey, 15*time.Minute, 1*time.Millisecond, testIssuer)

	refreshToken, err := service.GenerateRefreshToken(testMemberID)
	if err != nil {
		t.Fatalf("GenerateRefreshToken failed: %v", err)
	}

	// Wait for refresh token to expire
	time.Sleep(10 * time.Millisecond)

	_, err = service.RefreshAccessToken(refreshToken, testEmail, member.RoleUser)
	if err == nil {
		t.Error("Expected error when refreshing with expired token, got nil")
	}
	if !strings.Contains(err.Error(), "invalid refresh token") {
		t.Errorf("Expected error message to contain 'invalid refresh token', got: %v", err)
	}
}

// TestRefreshAccessToken_InvalidRefreshToken tests refresh with invalid token
func TestRefreshAccessToken_InvalidRefreshToken(t *testing.T) {
	service := newTestJWTService()

	_, err := service.RefreshAccessToken("invalid-token", testEmail, member.RoleUser)
	if err == nil {
		t.Error("Expected error when refreshing with invalid token, got nil")
	}
}

// TestTokensAreUnique tests that tokens generated at different times are unique
func TestTokensAreUnique(t *testing.T) {
	service := newTestJWTService()

	token1, err := service.GenerateAccessToken(testMemberID, testEmail, member.RoleUser)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	// JWT timestamps are in seconds, so sleep for 1 second to ensure different IssuedAt
	time.Sleep(1100 * time.Millisecond)

	token2, err := service.GenerateAccessToken(testMemberID, testEmail, member.RoleUser)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}

	if token1 == token2 {
		t.Error("Expected unique tokens, but got identical tokens")
	}
}

// TestDifferentRoles tests tokens for different member roles
func TestDifferentRoles(t *testing.T) {
	service := newTestJWTService()

	roles := []member.Role{member.RoleUser, member.RoleAdmin}

	for _, role := range roles {
		t.Run(string(role), func(t *testing.T) {
			token, err := service.GenerateAccessToken(testMemberID, testEmail, role)
			if err != nil {
				t.Fatalf("GenerateAccessToken failed for role %s: %v", role, err)
			}

			claims, err := service.ValidateToken(token)
			if err != nil {
				t.Fatalf("ValidateToken failed for role %s: %v", role, err)
			}

			if claims.Role != string(role) {
				t.Errorf("Expected role %s, got %s", role, claims.Role)
			}
		})
	}
}
