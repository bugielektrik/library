package auth

import (
	"library-service/internal/adapters/repository/mocks"
	"library-service/test/helpers"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	infraauth "library-service/internal/infrastructure/auth"
	"library-service/internal/infrastructure/store"
	"library-service/internal/members/domain"
	"library-service/pkg/errors"
)

func TestValidateTokenUseCase_Execute(t *testing.T) {
	// Create a JWT service for generating valid tokens
	jwtService := infraauth.NewJWTService("test-secret-key", 24*time.Hour, 7*24*time.Hour, "test-issuer")

	// Generate a valid token for use in tests
	validTokenPair, _ := jwtService.GenerateTokenPair("member-123", "user@example.com", domain.RoleUser)

	// Create a JWT service with different secret for invalid token test
	differentSecretService := infraauth.NewJWTService("different-secret", 24*time.Hour, 7*24*time.Hour, "test-issuer")
	invalidTokenPair, _ := differentSecretService.GenerateTokenPair("member-123", "user@example.com", domain.RoleUser)

	tests := []struct {
		name          string
		request       ValidateTokenRequest
		setupMocks    func(*mocks.MockMemberRepository)
		expectError   bool
		errorContains string
		validateFunc  func(*testing.T, ValidateTokenResponse)
	}{
		{
			name: "successful validation with valid token",
			request: ValidateTokenRequest{
				AccessToken: validTokenPair.AccessToken,
			},
			setupMocks: func(repo *mocks.MockMemberRepository) {
				// Member exists with matching data
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
			validateFunc: func(t *testing.T, resp ValidateTokenResponse) {
				helpers.AssertEqual(t, "member-123", resp.Member.ID)
				helpers.AssertEqual(t, "user@example.com", resp.Member.Email)
				helpers.AssertNotNil(t, resp.Claims)
				helpers.AssertEqual(t, "member-123", resp.Claims.MemberID)
			},
		},
		{
			name: "invalid token (wrong signature)",
			request: ValidateTokenRequest{
				AccessToken: invalidTokenPair.AccessToken,
			},
			setupMocks: func(repo *mocks.MockMemberRepository) {
				// No mocks should be called - validation fails before repo access
			},
			expectError:   true,
			errorContains: "Invalid or expired token",
		},
		{
			name: "invalid token format",
			request: ValidateTokenRequest{
				AccessToken: "invalid.token.format",
			},
			setupMocks: func(repo *mocks.MockMemberRepository) {
				// No mocks should be called
			},
			expectError:   true,
			errorContains: "Invalid or expired token",
		},
		{
			name: "member not found (deleted after token was issued)",
			request: ValidateTokenRequest{
				AccessToken: validTokenPair.AccessToken,
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
			name: "email changed (token data mismatch)",
			request: ValidateTokenRequest{
				AccessToken: validTokenPair.AccessToken,
			},
			setupMocks: func(repo *mocks.MockMemberRepository) {
				// Member exists but email has changed
				memberEntity := domain.Member{
					ID:           "member-123",
					Email:        "newemail@example.com", // Different from token
					PasswordHash: "hash",
					FullName:     strPtr("John Doe"),
					Role:         domain.RoleUser,
				}

				repo.On("Get", mock.Anything, "member-123").
					Return(memberEntity, nil).
					Once()
			},
			expectError:   true,
			errorContains: "Invalid or expired token",
		},
		{
			name: "role changed (token data mismatch)",
			request: ValidateTokenRequest{
				AccessToken: validTokenPair.AccessToken,
			},
			setupMocks: func(repo *mocks.MockMemberRepository) {
				// Member exists but role has changed
				memberEntity := domain.Member{
					ID:           "member-123",
					Email:        "user@example.com",
					PasswordHash: "hash",
					FullName:     strPtr("John Doe"),
					Role:         domain.RoleAdmin, // Different from token
				}

				repo.On("Get", mock.Anything, "member-123").
					Return(memberEntity, nil).
					Once()
			},
			expectError:   true,
			errorContains: "Invalid or expired token",
		},
		{
			name: "repository error when fetching member",
			request: ValidateTokenRequest{
				AccessToken: validTokenPair.AccessToken,
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
			uc := NewValidateTokenUseCase(mockRepo, jwtService)

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

func TestValidateTokenUseCase_GetCurrentMember(t *testing.T) {
	// Create a JWT service for generating valid tokens
	jwtService := infraauth.NewJWTService("test-secret-key", 24*time.Hour, 7*24*time.Hour, "test-issuer")

	// Generate a valid token
	validTokenPair, _ := jwtService.GenerateTokenPair("member-123", "user@example.com", domain.RoleUser)

	tests := []struct {
		name          string
		token         string
		setupMocks    func(*mocks.MockMemberRepository)
		expectError   bool
		errorContains string
		validateFunc  func(*testing.T, *domain.Member)
	}{
		{
			name:  "successful get current member",
			token: validTokenPair.AccessToken,
			setupMocks: func(repo *mocks.MockMemberRepository) {
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
			validateFunc: func(t *testing.T, m *domain.Member) {
				helpers.AssertNotNil(t, m)
				helpers.AssertEqual(t, "member-123", m.ID)
				helpers.AssertEqual(t, "user@example.com", m.Email)
			},
		},
		{
			name:  "invalid token",
			token: "invalid.token",
			setupMocks: func(repo *mocks.MockMemberRepository) {
				// No mocks should be called
			},
			expectError:   true,
			errorContains: "Invalid or expired token",
		},
		{
			name:  "member not found",
			token: validTokenPair.AccessToken,
			setupMocks: func(repo *mocks.MockMemberRepository) {
				repo.On("Get", mock.Anything, "member-123").
					Return(domain.Member{}, store.ErrorNotFound).
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
			uc := NewValidateTokenUseCase(mockRepo, jwtService)

			// Execute
			ctx := helpers.TestContext(t)
			result, err := uc.GetCurrentMember(ctx, tt.token)

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
