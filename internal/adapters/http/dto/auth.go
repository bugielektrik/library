package dto

// AuthRequest represents authentication request payload
type AuthRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// RegisterRequest represents registration request payload
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	FullName string `json:"full_name" validate:"required"`
}

// AuthResponse represents authentication response payload
type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// RefreshRequest represents token refresh request payload
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// MemberWithTokenResponse combines member data with authentication tokens
type MemberWithTokenResponse struct {
	Member MemberResponse `json:"member"`
	Tokens AuthResponse   `json:"tokens"`
}