package auth

import (
	"context"
	"fmt"
	"time"

	"library-service/internal/domain/member"
	"library-service/internal/infrastructure/auth"
	"library-service/pkg/errors"
)

// RegisterUseCase handles member registration
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
	Member       member.Response    `json:"member"`
	TokenPair    *auth.TokenPair    `json:"tokens"`
}

// Execute performs the member registration
func (uc *RegisterUseCase) Execute(ctx context.Context, req RegisterRequest) (*RegisterResponse, error) {
	// Validate email format
	if err := auth.ValidateEmail(req.Email); err != nil {
		return nil, errors.ErrInvalidInput.WithDetails("email", err.Error())
	}

	// Validate password strength
	if err := uc.passwordService.ValidatePassword(req.Password); err != nil {
		return nil, errors.ErrInvalidInput.WithDetails("password", err.Error())
	}

	// Check if email already exists
	exists, err := uc.memberRepo.EmailExists(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return nil, errors.ErrAlreadyExists.WithDetails("field", "email").
			WithDetails("value", req.Email)
	}

	// Hash the password
	passwordHash, err := uc.passwordService.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
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
	if err := uc.memberService.ValidateMember(newMember); err != nil {
		return nil, err
	}

	// Save member to repository
	memberID, err := uc.memberRepo.Add(ctx, newMember)
	if err != nil {
		return nil, fmt.Errorf("failed to create member: %w", err)
	}
	newMember.ID = memberID

	// Generate JWT tokens
	tokenPair, err := uc.jwtService.GenerateTokenPair(memberID, req.Email, member.RoleUser)
	if err != nil {
		return nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	// Prepare response
	response := &RegisterResponse{
		Member:    member.ParseFromMember(newMember),
		TokenPair: tokenPair,
	}

	return response, nil
}