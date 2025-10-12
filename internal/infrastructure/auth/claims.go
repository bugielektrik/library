package auth

import (
	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the JWT claims for access tokens
type Claims struct {
	MemberID string `json:"member_id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// RefreshClaims represents the JWT claims for refresh tokens
type RefreshClaims struct {
	MemberID string `json:"member_id"`
	jwt.RegisteredClaims
}

// TokenPair represents a pair of access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // Access token expiry in seconds
}
