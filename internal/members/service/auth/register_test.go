package auth

import (
	"context"
	"library-service/internal/pkg/errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	infraauth "library-service/internal/infrastructure/auth"
	"library-service/internal/members/domain"
	"library-service/internal/members/repository/mocks"
)

func TestRegisterUseCase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		request       RegisterRequest
		setupMocks    func(*mocks.MockMemberRepository)
		expectError   bool
		errorContains string
		validateFunc  func(*testing.T, RegisterResponse)
	}{
		{
			name: "successful registration",
			request: RegisterRequest{
				Email:    "newuser@example.com",
				Password: "SecureP@ss123",
				FullName: "John Doe",
			},
			setupMocks: func(repo *mocks.MockMemberRepository) {
				// Email doesn't exist
				repo.On("EmailExists", mock.Anything, "newuser@example.com").
					Return(false, nil).
					Once()

				// Member creation succeeds
				repo.On("Add", mock.Anything, mock.MatchedBy(func(m domain.Member) bool {
					return m.Email == "newuser@example.com" &&
						m.PasswordHash != "" &&
						m.Role == domain.RoleUser
				})).
					Return("member-123", nil).
					Once()
			},
			expectError: false,
			validateFunc: func(t *testing.T, resp RegisterResponse) {
				assert.Equal(t, "member-123", resp.Member.ID)
				assert.Equal(t, "newuser@example.com", resp.Member.Email)
				assert.Equal(t, "John Doe", resp.Member.FullName)
				assert.Equal(t, "user", resp.Member.Role)
				assert.NotNil(t, resp.TokenPair)
				assert.NotEmpty(t, resp.TokenPair.AccessToken)
				assert.NotEmpty(t, resp.TokenPair.RefreshToken)
			},
		},
		{
			name: "invalid email format",
			request: RegisterRequest{
				Email:    "invalid-email",
				Password: "SecureP@ss123",
				FullName: "John Doe",
			},
			setupMocks: func(repo *mocks.MockMemberRepository) {
				// No mock calls expected - validation fails first
			},
			expectError:   true,
			errorContains: "Validation failed",
		},
		{
			name: "weak password",
			request: RegisterRequest{
				Email:    "user@example.com",
				Password: "weak",
				FullName: "John Doe",
			},
			setupMocks: func(repo *mocks.MockMemberRepository) {
				// No mock calls expected - validation fails first
			},
			expectError:   true,
			errorContains: "Validation failed",
		},
		{
			name: "email already exists",
			request: RegisterRequest{
				Email:    "existing@example.com",
				Password: "SecureP@ss123",
				FullName: "John Doe",
			},
			setupMocks: func(repo *mocks.MockMemberRepository) {
				// Email exists
				repo.On("EmailExists", mock.Anything, "existing@example.com").
					Return(true, nil).
					Once()
			},
			expectError:   true,
			errorContains: "already exists",
		},
		{
			name: "repository error during creation",
			request: RegisterRequest{
				Email:    "user@example.com",
				Password: "SecureP@ss123",
				FullName: "John Doe",
			},
			setupMocks: func(repo *mocks.MockMemberRepository) {
				// Email doesn't exist
				repo.On("EmailExists", mock.Anything, "user@example.com").
					Return(false, nil).
					Once()

				// Repository error during Add
				repo.On("Add", mock.Anything, mock.Anything).
					Return("", errors.ErrDatabase).
					Once()
			},
			expectError:   true,
			errorContains: "Database operation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := new(mocks.MockMemberRepository)
			tt.setupMocks(mockRepo)

			// Create real auth service for testing
			jwtService := infraauth.NewJWTService("test-secret-key", 24*time.Hour, 7*24*time.Hour, "test-issuer")
			passwordService := infraauth.NewPasswordService()
			memberService := domain.NewService()

			// Create use case
			uc := NewRegisterUseCase(mockRepo, passwordService, jwtService, memberService)

			// Execute
			ctx := context.Background()
			result, err := uc.Execute(ctx, tt.request)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				require.NoError(t, err)
				if tt.validateFunc != nil {
					tt.validateFunc(t, result)
				}
			}

			// Verify all expectations were met
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestRegisterUseCase_PasswordHashing(t *testing.T) {
	// Setup
	mockRepo := new(mocks.MockMemberRepository)

	// Email doesn't exist
	mockRepo.On("EmailExists", mock.Anything, mock.Anything).
		Return(false, nil)

	// Capture the hashed password
	var capturedPasswordHash string
	mockRepo.On("Add", mock.Anything, mock.MatchedBy(func(m domain.Member) bool {
		capturedPasswordHash = m.PasswordHash
		return true
	})).
		Return("member-123", nil)

	jwtService := infraauth.NewJWTService("test-secret-key", 24*time.Hour, 7*24*time.Hour, "test-issuer")
	passwordService := infraauth.NewPasswordService()
	memberService := domain.NewService()

	uc := NewRegisterUseCase(mockRepo, passwordService, jwtService, memberService)

	// Execute
	req := RegisterRequest{
		Email:    "test@example.com",
		Password: "PlainPassword123!",
		FullName: "Test User",
	}

	ctx := context.Background()
	_, err := uc.Execute(ctx, req)
	require.NoError(t, err)

	// Verify password was hashed (not stored in plain text)
	assert.NotEmpty(t, capturedPasswordHash)
	assert.NotEqual(t, req.Password, capturedPasswordHash)
	assert.Greater(t, len(capturedPasswordHash), 20, "Bcrypt hashes are typically 60 chars")

	// Verify the hash can be validated
	isValid := passwordService.CheckPasswordHash(req.Password, capturedPasswordHash)
	assert.True(t, isValid)
}
