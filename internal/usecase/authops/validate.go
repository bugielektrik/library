package authops

import (
	"context"

	"go.uber.org/zap"

	"library-service/internal/domain/member"
	"library-service/internal/infrastructure/auth"
	"library-service/internal/infrastructure/store"
	"library-service/pkg/errors"
	"library-service/pkg/logutil"
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
func (uc *ValidateTokenUseCase) Execute(ctx context.Context, req ValidateTokenRequest) (ValidateTokenResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "auth", "validate")

	// Validate access token
	claims, err := uc.jwtService.ValidateToken(req.AccessToken)
	if err != nil {
		logger.Warn("token validation failed", zap.Error(err))
		return ValidateTokenResponse{}, errors.ErrInvalidToken.WithDetails("error", err.Error())
	}

	// Add member ID to logger after successful validation
	logger = logger.With(zap.String("member_id", claims.MemberID))

	// Get member from repository to ensure they still exist
	memberEntity, err := uc.memberRepo.Get(ctx, claims.MemberID)
	if err != nil {
		logger.Warn("member not found for token validation", zap.Error(err))
		return ValidateTokenResponse{}, errors.NotFound("member")
	}

	// Verify the email and role haven't changed
	if memberEntity.Email != claims.Email || string(memberEntity.Role) != claims.Role {
		logger.Warn("token data mismatch",
			zap.String("token_email", claims.Email),
			zap.String("member_email", memberEntity.Email),
			zap.String("token_role", claims.Role),
			zap.String("member_role", string(memberEntity.Role)),
		)
		return ValidateTokenResponse{}, errors.ErrInvalidToken.WithDetails("reason", "token data mismatch")
	}

	logger.Info("token validated successfully")

	return ValidateTokenResponse{
		Member: member.ParseFromMember(memberEntity),
		Claims: claims,
	}, nil
}

// GetCurrentMember is a simplified version that takes a token and returns the member
func (uc *ValidateTokenUseCase) GetCurrentMember(ctx context.Context, token string) (*member.Member, error) {
	logger := logutil.UseCaseLogger(ctx, "auth", "get_current")

	// Validate access token
	claims, err := uc.jwtService.ValidateToken(token)
	if err != nil {
		logger.Warn("token validation failed", zap.Error(err))
		return nil, errors.ErrInvalidToken.Wrap(err)
	}

	// Add member ID to logger after successful validation
	logger = logger.With(zap.String("member_id", claims.MemberID))

	// Get member from repository
	memberEntity, err := uc.memberRepo.Get(ctx, claims.MemberID)
	if err != nil {
		logger.Warn("member not found", zap.Error(err))
		if errors.Is(err, store.ErrorNotFound) {
			return nil, errors.NotFoundWithID("member", claims.MemberID)
		}
		return nil, errors.Database("database operation", err)
	}

	logger.Info("current member retrieved successfully")

	return &memberEntity, nil
}
