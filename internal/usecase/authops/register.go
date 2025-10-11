package authops

import (
	"context"
	"time"

	"go.uber.org/zap"

	"library-service/internal/domain/member"
	"library-service/internal/infrastructure/auth"
	"library-service/pkg/errors"
	"library-service/pkg/logutil"
)

// RegisterUseCase handles member registration.
//
// Architecture Pattern: Authentication use case with infrastructure services.
// Demonstrates usage of JWTService and PasswordService (external dependencies).
//
// See Also:
//   - Similar pattern: internal/usecase/authops/login.go (authentication flow)
//   - Infrastructure: internal/infrastructure/auth/jwt.go (token generation)
//   - Infrastructure: internal/infrastructure/auth/password.go (hashing)
//   - Domain service: internal/domain/member/service.go (validation)
//   - HTTP handler: internal/adapters/http/handlers/auth/register.go
//   - ADR: .claude/adr/003-domain-services-vs-infrastructure.md (JWT is infrastructure)
//   - Test: internal/usecase/authops/register_test.go
type RegisterUseCase struct {
	memberRepo      member.Repository
	passwordService *auth.PasswordService
	jwtService      *auth.JWTService
	memberService   *member.Service
}

// NewRegisterUseCase creates a new register use case
func NewRegisterUseCase(
	memberRepo member.Repository,
	passwordService *auth.PasswordService,
	jwtService *auth.JWTService,
	memberService *member.Service,
) *RegisterUseCase {
	return &RegisterUseCase{
		memberRepo:      memberRepo,
		passwordService: passwordService,
		jwtService:      jwtService,
		memberService:   memberService,
	}
}

// RegisterRequest represents the registration request
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	FullName string `json:"full_name" validate:"required"`
}

// RegisterResponse represents the registration response
type RegisterResponse struct {
	Member    member.Response `json:"member"`
	TokenPair *auth.TokenPair `json:"tokens"`
}

// Execute performs the member registration
func (uc *RegisterUseCase) Execute(ctx context.Context, req RegisterRequest) (RegisterResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "auth", "register")

	// Validate email format
	if err := auth.ValidateEmail(req.Email); err != nil {
		logger.Warn("email validation failed", zap.Error(err))
		return RegisterResponse{}, errors.NewError(errors.CodeValidation).WithField("email", "invalid format").WithDetail("details", err.Error()).Build()
	}

	// Validate password strength
	if err := uc.passwordService.ValidatePassword(req.Password); err != nil {
		logger.Warn("password validation failed", zap.Error(err))
		return RegisterResponse{}, errors.NewError(errors.CodeValidation).WithField("password", "invalid format").WithDetail("details", err.Error()).Build()
	}

	// Check if email already exists
	exists, err := uc.memberRepo.EmailExists(ctx, req.Email)
	if err != nil {
		logger.Error("failed to check email existence", zap.Error(err))
		return RegisterResponse{}, errors.ErrDatabase.
			WithDetails("operation", "email_exists_check").
			WithDetails("email", req.Email).
			Wrap(err)
	}
	if exists {
		logger.Warn("email already exists", zap.String("email", req.Email))
		return RegisterResponse{}, errors.ErrAlreadyExists.WithDetails("field", "email").
			WithDetails("value", req.Email)
	}

	// Hash the password
	passwordHash, err := uc.passwordService.HashPassword(req.Password)
	if err != nil {
		logger.Error("failed to hash password", zap.Error(err))
		return RegisterResponse{}, errors.ErrInternal.
			WithDetails("operation", "hash_password").
			Wrap(err)
	}

	// Create new member
	now := time.Now()
	newMember := member.Member{
		Email:        req.Email,
		PasswordHash: passwordHash,
		FullName:     &req.FullName,
		Role:         member.RoleUser,
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
		return RegisterResponse{}, errors.ErrDatabase.
			WithDetails("operation", "create_member").
			WithDetails("email", req.Email).
			Wrap(err)
	}
	newMember.ID = memberID

	// Generate JWT tokens
	tokenPair, err := uc.jwtService.GenerateTokenPair(memberID, req.Email, member.RoleUser)
	if err != nil {
		logger.Error("failed to generate JWT tokens", zap.Error(err))
		return RegisterResponse{}, errors.ErrInternal.
			WithDetails("operation", "generate_tokens").
			WithDetails("member_id", memberID).
			Wrap(err)
	}

	logger.Info("member registered successfully", zap.String("member_id", memberID))

	return RegisterResponse{
		Member:    member.ParseFromMember(newMember),
		TokenPair: tokenPair,
	}, nil
}
