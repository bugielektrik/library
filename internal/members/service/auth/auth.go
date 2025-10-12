package auth

import (
	"context"
	"time"

	"go.uber.org/zap"

	infraauth "library-service/internal/infrastructure/auth"
	"library-service/internal/infrastructure/store"
	"library-service/internal/members/domain"
	errors2 "library-service/internal/pkg/errors"
	"library-service/internal/pkg/logutil"
)

// ================================================================================
// Login Use Case
// ================================================================================

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

// Execute performs member authentication
func (uc *LoginUseCase) Execute(ctx context.Context, req LoginRequest) (LoginResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "auth", "login")

	// Validate email format
	if err := infraauth.ValidateEmail(req.Email); err != nil {
		logger.Warn("email validation failed", zap.Error(err))
		return LoginResponse{}, errors2.NewError(errors2.CodeValidation).WithField("email", "invalid format").WithDetail("details", err.Error()).Build()
	}

	// Find member by email
	memberEntity, err := uc.memberRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		// Don't reveal if email exists or not for security
		logger.Warn("authentication failed", zap.String("reason", "member_not_found"))
		return LoginResponse{}, errors2.Unauthorized("invalid credentials")
	}

	// Verify password
	if !uc.passwordService.CheckPasswordHash(req.Password, memberEntity.PasswordHash) {
		logger.Warn("authentication failed", zap.String("reason", "invalid_password"))
		return LoginResponse{}, errors2.Unauthorized("invalid credentials")
	}

	// Generate JWT tokens
	tokenPair, err := uc.jwtService.GenerateTokenPair(
		memberEntity.ID,
		memberEntity.Email,
		string(memberEntity.Role),
	)
	if err != nil {
		logger.Error("failed to generate JWT tokens", zap.Error(err))
		return LoginResponse{}, errors2.ErrInternal.
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

// ================================================================================
// Register Use Case
// ================================================================================

// RegisterRequest represents the registration request
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	FullName string `json:"full_name" validate:"required"`
}

// RegisterResponse represents the registration response
type RegisterResponse struct {
	Member    domain.Response      `json:"member"`
	TokenPair *infraauth.TokenPair `json:"tokens"`
}

// RegisterUseCase handles member registration.
//
// Architecture Pattern: Authentication use case with infrastructure service.
// Demonstrates usage of JWTService and PasswordService (external dependencies).
//
// See Also:
//   - Similar pattern: internal/usecase/authops/login.go (authentication flow)
//   - Infrastructure: internal/infrastructure/auth/jwt.go (token generation)
//   - Infrastructure: internal/infrastructure/auth/password.go (hashing)
//   - Domain service: internal/domain/member/service.go (validation)
//   - HTTP handler: internal/adapters/http/handler/auth/register.go
//   - ADR: .claude/adr/003-domain-service-vs-infrastructure.md (JWT is infrastructure)
//   - Test: internal/usecase/authops/register_test.go
type RegisterUseCase struct {
	memberRepo      domain.Repository
	passwordService *infraauth.PasswordService
	jwtService      *infraauth.JWTService
	memberService   *domain.Service
}

// NewRegisterUseCase creates a new register use case
func NewRegisterUseCase(
	memberRepo domain.Repository,
	passwordService *infraauth.PasswordService,
	jwtService *infraauth.JWTService,
	memberService *domain.Service,
) *RegisterUseCase {
	return &RegisterUseCase{
		memberRepo:      memberRepo,
		passwordService: passwordService,
		jwtService:      jwtService,
		memberService:   memberService,
	}
}

// Execute performs the member registration
func (uc *RegisterUseCase) Execute(ctx context.Context, req RegisterRequest) (RegisterResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "auth", "register")

	// Validate email format
	if err := infraauth.ValidateEmail(req.Email); err != nil {
		logger.Warn("email validation failed", zap.Error(err))
		return RegisterResponse{}, errors2.NewError(errors2.CodeValidation).WithField("email", "invalid format").WithDetail("details", err.Error()).Build()
	}

	// Validate password strength
	if err := uc.passwordService.ValidatePassword(req.Password); err != nil {
		logger.Warn("password validation failed", zap.Error(err))
		return RegisterResponse{}, errors2.NewError(errors2.CodeValidation).WithField("password", "invalid format").WithDetail("details", err.Error()).Build()
	}

	// Check if email already exists
	exists, err := uc.memberRepo.EmailExists(ctx, req.Email)
	if err != nil {
		logger.Error("failed to check email existence", zap.Error(err))
		return RegisterResponse{}, errors2.ErrDatabase.
			WithDetails("operation", "email_exists_check").
			WithDetails("email", req.Email).
			Wrap(err)
	}
	if exists {
		logger.Warn("email already exists", zap.String("email", req.Email))
		return RegisterResponse{}, errors2.ErrAlreadyExists.WithDetails("field", "email").
			WithDetails("value", req.Email)
	}

	// Hash the password
	passwordHash, err := uc.passwordService.HashPassword(req.Password)
	if err != nil {
		logger.Error("failed to hash password", zap.Error(err))
		return RegisterResponse{}, errors2.ErrInternal.
			WithDetails("operation", "hash_password").
			Wrap(err)
	}

	// Create new member
	now := time.Now()
	newMember := domain.Member{
		Email:        req.Email,
		PasswordHash: passwordHash,
		FullName:     &req.FullName,
		Role:         domain.RoleUser,
		Books:        []string{},
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Validate member using domain service
	if err := uc.memberService.Validate(newMember); err != nil {
		logger.Warn("member validation failed", zap.Error(err))
		return RegisterResponse{}, err
	}

	// Save member to repository
	memberID, err := uc.memberRepo.Add(ctx, newMember)
	if err != nil {
		logger.Error("failed to create member in repository", zap.Error(err))
		return RegisterResponse{}, errors2.ErrDatabase.
			WithDetails("operation", "create_member").
			WithDetails("email", req.Email).
			Wrap(err)
	}
	newMember.ID = memberID

	// Generate JWT tokens
	tokenPair, err := uc.jwtService.GenerateTokenPair(memberID, req.Email, string(domain.RoleUser))
	if err != nil {
		logger.Error("failed to generate JWT tokens", zap.Error(err))
		return RegisterResponse{}, errors2.ErrInternal.
			WithDetails("operation", "generate_tokens").
			WithDetails("member_id", memberID).
			Wrap(err)
	}

	logger.Info("member registered successfully", zap.String("member_id", memberID))

	return RegisterResponse{
		Member:    domain.ParseFromMember(newMember),
		TokenPair: tokenPair,
	}, nil
}

// ================================================================================
// Refresh Token Use Case
// ================================================================================

// RefreshTokenRequest represents the refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// RefreshTokenResponse represents the refresh token response
type RefreshTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

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

// ================================================================================
// Validate Token Use Case
// ================================================================================

// ValidateTokenRequest represents the validate token request
type ValidateTokenRequest struct {
	AccessToken string `json:"access_token" validate:"required"`
}

// ValidateTokenResponse represents the validate token response
type ValidateTokenResponse struct {
	Member domain.Response   `json:"member"`
	Claims *infraauth.Claims `json:"claims"`
}

// ValidateTokenUseCase handles token validation and returns member info
type ValidateTokenUseCase struct {
	memberRepo domain.Repository
	jwtService *infraauth.JWTService
}

// NewValidateTokenUseCase creates a new validate token use case
func NewValidateTokenUseCase(
	memberRepo domain.Repository,
	jwtService *infraauth.JWTService,
) *ValidateTokenUseCase {
	return &ValidateTokenUseCase{
		memberRepo: memberRepo,
		jwtService: jwtService,
	}
}

// Execute validates token and returns member information
func (uc *ValidateTokenUseCase) Execute(ctx context.Context, req ValidateTokenRequest) (ValidateTokenResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "auth", "validate")

	// Validate access token
	claims, err := uc.jwtService.ValidateToken(req.AccessToken)
	if err != nil {
		logger.Warn("token validation failed", zap.Error(err))
		return ValidateTokenResponse{}, errors2.ErrInvalidToken.WithDetails("error", err.Error())
	}

	// Add member ID to logger after successful validation
	logger = logger.With(zap.String("member_id", claims.MemberID))

	// Get member from repository to ensure they still exist
	memberEntity, err := uc.memberRepo.Get(ctx, claims.MemberID)
	if err != nil {
		logger.Warn("member not found for token validation", zap.Error(err))
		return ValidateTokenResponse{}, errors2.NotFound("member")
	}

	// Verify the email and role haven't changed
	if memberEntity.Email != claims.Email || string(memberEntity.Role) != claims.Role {
		logger.Warn("token data mismatch",
			zap.String("token_email", claims.Email),
			zap.String("member_email", memberEntity.Email),
			zap.String("token_role", claims.Role),
			zap.String("member_role", string(memberEntity.Role)),
		)
		return ValidateTokenResponse{}, errors2.ErrInvalidToken.WithDetails("reason", "token data mismatch")
	}

	logger.Info("token validated successfully")

	return ValidateTokenResponse{
		Member: domain.ParseFromMember(memberEntity),
		Claims: claims,
	}, nil
}

// GetCurrentMember is a simplified version that takes a token and returns the member
func (uc *ValidateTokenUseCase) GetCurrentMember(ctx context.Context, token string) (*domain.Member, error) {
	logger := logutil.UseCaseLogger(ctx, "auth", "get_current")

	// Validate access token
	claims, err := uc.jwtService.ValidateToken(token)
	if err != nil {
		logger.Warn("token validation failed", zap.Error(err))
		return nil, errors2.ErrInvalidToken.Wrap(err)
	}

	// Add member ID to logger after successful validation
	logger = logger.With(zap.String("member_id", claims.MemberID))

	// Get member from repository
	memberEntity, err := uc.memberRepo.Get(ctx, claims.MemberID)
	if err != nil {
		logger.Warn("member not found", zap.Error(err))
		if errors2.Is(err, store.ErrorNotFound) {
			return nil, errors2.NotFoundWithID("member", claims.MemberID)
		}
		return nil, errors2.Database("database operation", err)
	}

	logger.Info("current member retrieved successfully")

	return &memberEntity, nil
}
