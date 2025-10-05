package authops

import (
	"context"
	"fmt"

	"library-service/internal/domain/member"
	"library-service/internal/infrastructure/auth"
	"library-service/pkg/errors"
)

// ValidateTokenUseCase handles token validation and returns member info
type ValidateTokenUseCase struct {
	memberRepo member.Repository
	jwtService *auth.JWTService
}

// NewValidateTokenUseCase creates a new validate token use case
func NewValidateTokenUseCase(
	memberRepo member.Repository,
	jwtService *auth.JWTService,
) *ValidateTokenUseCase {
	return &ValidateTokenUseCase{
		memberRepo: memberRepo,
		jwtService: jwtService,
	}
}

// ValidateTokenRequest represents the validate token request
type ValidateTokenRequest struct {
	AccessToken string `json:"access_token" validate:"required"`
}

// ValidateTokenResponse represents the validate token response
type ValidateTokenResponse struct {
	Member member.Response `json:"member"`
	Claims *auth.Claims    `json:"claims"`
}

// Execute validates token and returns member information
func (uc *ValidateTokenUseCase) Execute(ctx context.Context, req ValidateTokenRequest) (*ValidateTokenResponse, error) {
	// Validate access token
	claims, err := uc.jwtService.ValidateToken(req.AccessToken)
	if err != nil {
		return nil, errors.ErrInvalidToken.WithDetails("error", err.Error())
	}

	// Get member from repository to ensure they still exist
	memberEntity, err := uc.memberRepo.Get(ctx, claims.MemberID)
	if err != nil {
		return nil, errors.ErrNotFound.WithDetails("entity", "member")
	}

	// Verify the email and role haven't changed
	if memberEntity.Email != claims.Email || string(memberEntity.Role) != claims.Role {
		return nil, errors.ErrInvalidToken.WithDetails("reason", "token data mismatch")
	}

	return &ValidateTokenResponse{
		Member: member.ParseFromMember(memberEntity),
		Claims: claims,
	}, nil
}

// GetCurrentMember is a simplified version that takes a token and returns the member
func (uc *ValidateTokenUseCase) GetCurrentMember(ctx context.Context, token string) (*member.Member, error) {
	// Validate access token
	claims, err := uc.jwtService.ValidateToken(token)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	// Get member from repository
	memberEntity, err := uc.memberRepo.Get(ctx, claims.MemberID)
	if err != nil {
		return nil, fmt.Errorf("member not found: %w", err)
	}

	return &memberEntity, nil
}