package auth

import (
	"context"
	errors2 "library-service/internal/pkg/errors"
	"library-service/internal/pkg/logutil"

	"go.uber.org/zap"

	infraauth "library-service/internal/infrastructure/auth"
	"library-service/internal/members/domain"
)

// RefreshTokenUseCase handles token refresh
type RefreshTokenUseCase struct {
	memberRepo domain.Repository
	jwtService *infraauth.JWTService
}

// NewRefreshTokenUseCase creates a new refresh token use case
func NewRefreshTokenUseCase(
	memberRepo domain.Repository,
	jwtService *infraauth.JWTService,
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
func (uc *RefreshTokenUseCase) Execute(ctx context.Context, req RefreshTokenRequest) (RefreshTokenResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "auth", "refresh")

	// Validate refresh token
	refreshClaims, err := uc.jwtService.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		logger.Warn("token validation failed", zap.Error(err))
		return RefreshTokenResponse{}, errors2.ErrInvalidToken.WithDetails("error", err.Error())
	}

	// Add member ID to logger after successful validation
	logger = logger.With(zap.String("member_id", refreshClaims.MemberID))

	// Get member from repository to ensure they still exist and get current data
	memberEntity, err := uc.memberRepo.Get(ctx, refreshClaims.MemberID)
	if err != nil {
		logger.Warn("member not found for token refresh", zap.Error(err))
		return RefreshTokenResponse{}, errors2.NotFound("member")
	}

	// Generate new access token with current member data
	accessToken, err := uc.jwtService.GenerateAccessToken(
		memberEntity.ID,
		memberEntity.Email,
		string(memberEntity.Role),
	)
	if err != nil {
		logger.Error("failed to generate new access token", zap.Error(err))
		return RefreshTokenResponse{}, errors2.ErrInternal.
			WithDetails("operation", "generate_access_token").
			WithDetails("member_id", memberEntity.ID).
			Wrap(err)
	}

	// Calculate expiry time in seconds (24 hours)
	expiresIn := int64(86400) // 24 * 60 * 60 seconds

	logger.Info("token refreshed successfully")

	return RefreshTokenResponse{
		AccessToken: accessToken,
		ExpiresIn:   expiresIn,
	}, nil
}
