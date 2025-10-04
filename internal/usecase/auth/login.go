package auth

import (
	"context"
	"fmt"
	"time"

	"library-service/internal/domain/member"
	"library-service/internal/infrastructure/auth"
	"library-service/pkg/errors"
)

// LoginUseCase handles member authentication
type LoginUseCase struct {
	memberRepo      member.Repository
	passwordService *auth.PasswordService
	jwtService      *auth.JWTService
}

// NewLoginUseCase creates a new login use case
func NewLoginUseCase(
	memberRepo member.Repository,
	passwordService *auth.PasswordService,
	jwtService *auth.JWTService,
) *LoginUseCase {
	return &LoginUseCase{
		memberRepo:      memberRepo,
		passwordService: passwordService,
		jwtService:      jwtService,
	}
}

// LoginRequest represents the login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Member    member.Response `json:"member"`
	TokenPair *auth.TokenPair `json:"tokens"`
}

// Execute performs member authentication
func (uc *LoginUseCase) Execute(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	// Validate email format
	if err := auth.ValidateEmail(req.Email); err != nil {
		return nil, errors.ErrInvalidInput.WithDetails("email", err.Error())
	}

	// Find member by email
	memberEntity, err := uc.memberRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		// Don't reveal if email exists or not for security
		return nil, errors.ErrInvalidCredentials
	}

	// Verify password
	if !uc.passwordService.CheckPasswordHash(req.Password, memberEntity.PasswordHash) {
		return nil, errors.ErrInvalidCredentials
	}

	// Generate JWT tokens
	tokenPair, err := uc.jwtService.GenerateTokenPair(
		memberEntity.ID,
		memberEntity.Email,
		memberEntity.Role,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Update last login timestamp
	now := time.Now()
	if err := uc.memberRepo.UpdateLastLogin(ctx, memberEntity.ID, now); err != nil {
		// Log the error but don't fail the login
		// This is a non-critical operation
		fmt.Printf("failed to update last login for member %s: %v\n", memberEntity.ID, err)
	}

	// Prepare response
	response := &LoginResponse{
		Member:    member.ParseFromMember(memberEntity),
		TokenPair: tokenPair,
	}

	return response, nil
}