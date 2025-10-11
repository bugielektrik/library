package auth

import (
	"library-service/internal/adapters/repository/mocks"
	"library-service/test/helpers"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	infraauth "library-service/internal/infrastructure/auth"
	"library-service/internal/members/domain"
	"library-service/pkg/errors"
)

func TestRefreshTokenUseCase_Execute(t *testing.T) {
	// Create a JWT service for generating valid tokens
	jwtService := infraauth.NewJWTService("test-secret-key", 24*time.Hour, 7*24*time.Hour, "test-issuer")

	// Generate a valid refresh token for use in tests
	validTokenPair, _ := jwtService.GenerateTokenPair("member-123", "user@example.com", domain.RoleUser)

	// Create a JWT service with different secret for invalid token test
	differentSecretService := infraauth.NewJWTService("different-secret", 24*time.Hour, 7*24*time.Hour, "test-issuer")
	invalidTokenPair, _ := differentSecretService.GenerateTokenPair("member-123", "user@example.com", domain.RoleUser)

	tests := []struct {
		name          string
		request       RefreshTokenRequest
		setupMocks    func(*mocks.MockMemberRepository)
		expectError   bool
		errorContains string
		validateFunc  func(*testing.T, RefreshTokenResponse)
	}{
		{
			name: "successful token refresh with valid refresh token",
			request: RefreshTokenRequest{
				RefreshToken: validTokenPair.RefreshToken,
			},
			setupMocks: func(repo *mocks.MockMemberRepository) {
				// Member exists
				memberEntity := domain.Member{
					ID:           "member-123",
					Email:        "user@example.com",
					PasswordHash: "hash",
					FullName:     strPtr("John Doe"),
					Role:         domain.RoleUser,
				}

				repo.On("Get", mock.Anything, "member-123").
					Return(memberEntity, nil).
					Once()
			},
			expectError: false,
			validateFunc: func(t *testing.T, resp RefreshTokenResponse) {
				helpers.AssertTrue(t, resp.AccessToken != "")
				helpers.AssertTrue(t, resp.ExpiresIn > 0)
			},
		},
		{
			name: "invalid refresh token (wrong signature)",
			request: RefreshTokenRequest{
				RefreshToken: invalidTokenPair.RefreshToken,
			},
			setupMocks: func(repo *mocks.MockMemberRepository) {
				// No mocks should be called - validation fails before repo access
			},
			expectError:   true,
			errorContains: "Invalid or expired token",
		},
		{
			name: "invalid refresh token format",
			request: RefreshTokenRequest{
				RefreshToken: "invalid.token.format",
			},
			setupMocks: func(repo *mocks.MockMemberRepository) {
				// No mocks should be called
			},
			expectError:   true,
			errorContains: "Invalid or expired token",
		},
		{
			name: "member not found (deleted after token was issued)",
			request: RefreshTokenRequest{
				RefreshToken: validTokenPair.RefreshToken,
			},
			setupMocks: func(repo *mocks.MockMemberRepository) {
				// Member doesn't exist
				repo.On("Get", mock.Anything, "member-123").
					Return(domain.Member{}, errors.ErrNotFound).
					Once()
			},
			expectError:   true,
			errorContains: "not found",
		},
		{
			name: "repository error when fetching member",
			request: RefreshTokenRequest{
				RefreshToken: validTokenPair.RefreshToken,
			},
			setupMocks: func(repo *mocks.MockMemberRepository) {
				// Repository returns database error
				repo.On("Get", mock.Anything, "member-123").
					Return(domain.Member{}, errors.ErrDatabase).
					Once()
			},
			expectError:   true,
			errorContains: "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := new(mocks.MockMemberRepository)
			tt.setupMocks(mockRepo)

			// Create use case
			uc := NewRefreshTokenUseCase(mockRepo, jwtService)

			// Execute
			ctx := helpers.TestContext(t)
			result, err := uc.Execute(ctx, tt.request)

			// Assert
			if tt.expectError {
				helpers.AssertError(t, err)
				if tt.errorContains != "" {
					helpers.AssertErrorContains(t, err, tt.errorContains)
				}
			} else {
				helpers.AssertNoError(t, err)
				if tt.validateFunc != nil {
					tt.validateFunc(t, result)
				}
			}

			// Verify all expectations were met
			mockRepo.AssertExpectations(t)
		})
	}
}
