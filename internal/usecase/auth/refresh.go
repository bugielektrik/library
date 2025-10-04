package auth

import (
	"context"
	"fmt"

	"library-service/internal/domain/member"
	"library-service/internal/infrastructure/auth"
	"library-service/pkg/errors"
)

// RefreshTokenUseCase handles token refresh
type RefreshTokenUseCase struct {
	memberRepo member.Repository
	jwtService *auth.JWTService
}

// NewRefreshTokenUseCase creates a new refresh token use case
func NewRefreshTokenUseCase(
	memberRepo member.Repository,
	jwtService *auth.JWTService,
) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{
		memberRepo: memberRepo,
		jwtService: jwtService,
	}
}

// RefreshTokenRequest represents the refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshTokenResponse represents the refresh token response
type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

// Execute performs token refresh
func (uc *RefreshTokenUseCase) Execute(ctx context.Context, req RefreshTokenRequest) (*RefreshTokenResponse, error) {
	// Validate refresh token
	refreshClaims, err := uc.jwtService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, errors.ErrInvalidToken.WithDetails("error", err.Error())
	}

	// Get member from repository to ensure they still exist and get current data
	memberEntity, err := uc.memberRepo.Get(ctx, refreshClaims.MemberID)
	if err != nil {
		return nil, errors.ErrNotFound.WithDetails("entity", "member")
	}

	// Generate new access token with current member data
	accessToken, err := uc.jwtService.GenerateAccessToken(
		memberEntity.ID,
		memberEntity.Email,
		memberEntity.Role,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate new access token: %w", err)
	}

	// Calculate expiry time in seconds
	expiresIn := int64(86400) // 24 hours default, should come from config

	return &RefreshTokenResponse{
		AccessToken: accessToken,
		ExpiresIn:   expiresIn,
	}, nil
}