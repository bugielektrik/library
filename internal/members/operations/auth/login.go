package auth

import (
	"context"
	"time"

	"go.uber.org/zap"

	infraauth "library-service/internal/infrastructure/auth"
	"library-service/internal/members/domain"
	"library-service/pkg/errors"
	"library-service/pkg/logutil"
)

// LoginUseCase handles member authentication
type LoginUseCase struct {
	memberRepo      domain.Repository
	passwordService *infraauth.PasswordService
	jwtService      *infraauth.JWTService
}

// NewLoginUseCase creates a new login use case
func NewLoginUseCase(
	memberRepo domain.Repository,
	passwordService *infraauth.PasswordService,
	jwtService *infraauth.JWTService,
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
	Member    domain.Response      `json:"member"`
	TokenPair *infraauth.TokenPair `json:"tokens"`
}

// Execute performs member authentication
func (uc *LoginUseCase) Execute(ctx context.Context, req LoginRequest) (LoginResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "auth", "login")

	// Validate email format
	if err := infraauth.ValidateEmail(req.Email); err != nil {
		logger.Warn("email validation failed", zap.Error(err))
		return LoginResponse{}, errors.NewError(errors.CodeValidation).WithField("email", "invalid format").WithDetail("details", err.Error()).Build()
	}

	// Find member by email
	memberEntity, err := uc.memberRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		// Don't reveal if email exists or not for security
		logger.Warn("authentication failed", zap.String("reason", "member_not_found"))
		return LoginResponse{}, errors.Unauthorized("invalid credentials")
	}

	// Verify password
	if !uc.passwordService.CheckPasswordHash(req.Password, memberEntity.PasswordHash) {
		logger.Warn("authentication failed", zap.String("reason", "invalid_password"))
		return LoginResponse{}, errors.Unauthorized("invalid credentials")
	}

	// Generate JWT tokens
	tokenPair, err := uc.jwtService.GenerateTokenPair(
		memberEntity.ID,
		memberEntity.Email,
		memberEntity.Role,
	)
	if err != nil {
		logger.Error("failed to generate JWT tokens", zap.Error(err))
		return LoginResponse{}, errors.ErrInternal.
			WithDetails("operation", "generate_tokens").
			WithDetails("member_id", memberEntity.ID).
			Wrap(err)
	}

	// Update last login timestamp
	now := time.Now()
	if err := uc.memberRepo.UpdateLastLogin(ctx, memberEntity.ID, now); err != nil {
		// Log the error but don't fail the login
		// This is a non-critical operation
		logger.Warn("failed to update last login",
			zap.String("member_id", memberEntity.ID),
			zap.Error(err),
		)
	}

	logger.Info("member logged in successfully", zap.String("member_id", memberEntity.ID))

	return LoginResponse{
		Member:    domain.ParseFromMember(memberEntity),
		TokenPair: tokenPair,
	}, nil
}
