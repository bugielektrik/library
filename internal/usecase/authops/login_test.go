package authops

import (
	"library-service/internal/adapters/repository/mocks"
	"library-service/test/helpers"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"library-service/internal/domain/member"
	"library-service/internal/infrastructure/auth"
	"library-service/pkg/errors"
)

// Helper function to create string pointers
func strPtr(s string) *string {
	return &s
}

func TestLoginUseCase_Execute(t *testing.T) {
	// Pre-hash a password for use in tests
	passwordService := auth.NewPasswordService()
	correctPasswordHash, _ := passwordService.HashPassword("ValidP@ssword123")

	tests := []struct {
		name          string
		request       LoginRequest
		setupMocks    func(*mocks.MockMemberRepository)
		expectError   bool
		errorContains string
		validateFunc  func(*testing.T, LoginResponse)
	}{
		{
			name: "successful login with valid credentials",
			request: LoginRequest{
				Email:    "user@example.com",
				Password: "ValidP@ssword123",
			},
			setupMocks: func(repo *mocks.MockMemberRepository) {
				// Member exists with correct password hash
				memberEntity := member.Member{
					ID:           "member-123",
					Email:        "user@example.com",
					PasswordHash: correctPasswordHash,
					FullName:     strPtr("John Doe"),
					Role:         member.RoleUser,
				}

				repo.On("GetByEmail", mock.Anything, "user@example.com").
					Return(memberEntity, nil).
					Once()

				// Last login update succeeds
				repo.On("UpdateLastLogin", mock.Anything, "member-123", mock.AnythingOfType("time.Time")).
					Return(nil).
					Once()
			},
			expectError: false,
			validateFunc: func(t *testing.T, resp LoginResponse) {
				helpers.AssertEqual(t, "member-123", resp.Member.ID)
				helpers.AssertEqual(t, "user@example.com", resp.Member.Email)
				helpers.AssertNotNil(t, resp.TokenPair)
				helpers.AssertTrue(t, resp.TokenPair.AccessToken != "")
				helpers.AssertTrue(t, resp.TokenPair.RefreshToken != "")
			},
		},
		{
			name: "invalid email format",
			request: LoginRequest{
				Email:    "invalid-email",
				Password: "ValidP@ssword123",
			},
			setupMocks: func(repo *mocks.MockMemberRepository) {
				// No mocks should be called
			},
			expectError:   true,
			errorContains: "Validation failed",
		},
		{
			name: "email not found (member doesn't exist)",
			request: LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "ValidP@ssword123",
			},
			setupMocks: func(repo *mocks.MockMemberRepository) {
				// Repository returns not found error
				repo.On("GetByEmail", mock.Anything, "nonexistent@example.com").
					Return(member.Member{}, errors.ErrNotFound).
					Once()
			},
			expectError:   true,
			errorContains: "Unauthorized",
		},
		{
			name: "wrong password",
			request: LoginRequest{
				Email:    "user@example.com",
				Password: "WrongPassword123",
			},
			setupMocks: func(repo *mocks.MockMemberRepository) {
				// Member exists but password doesn't match
				memberEntity := member.Member{
					ID:           "member-123",
					Email:        "user@example.com",
					PasswordHash: correctPasswordHash,
					FullName:     strPtr("John Doe"),
					Role:         member.RoleUser,
				}

				repo.On("GetByEmail", mock.Anything, "user@example.com").
					Return(memberEntity, nil).
					Once()
			},
			expectError:   true,
			errorContains: "Unauthorized",
		},
		{
			name: "repository error when fetching member",
			request: LoginRequest{
				Email:    "user@example.com",
				Password: "ValidP@ssword123",
			},
			setupMocks: func(repo *mocks.MockMemberRepository) {
				// Repository returns database error
				repo.On("GetByEmail", mock.Anything, "user@example.com").
					Return(member.Member{}, errors.ErrDatabase).
					Once()
			},
			expectError:   true,
			errorContains: "Unauthorized",
		},
		{
			name: "last login update failure (should still succeed)",
			request: LoginRequest{
				Email:    "user@example.com",
				Password: "ValidP@ssword123",
			},
			setupMocks: func(repo *mocks.MockMemberRepository) {
				// Member exists with correct password hash
				memberEntity := member.Member{
					ID:           "member-456",
					Email:        "user@example.com",
					PasswordHash: correctPasswordHash,
					FullName:     strPtr("Jane Doe"),
					Role:         member.RoleUser,
				}

				repo.On("GetByEmail", mock.Anything, "user@example.com").
					Return(memberEntity, nil).
					Once()

				// Last login update fails (non-critical)
				repo.On("UpdateLastLogin", mock.Anything, "member-456", mock.AnythingOfType("time.Time")).
					Return(errors.ErrDatabase).
					Once()
			},
			expectError: false, // Should succeed despite last login update failure
			validateFunc: func(t *testing.T, resp LoginResponse) {
				helpers.AssertEqual(t, "member-456", resp.Member.ID)
				helpers.AssertNotNil(t, resp.TokenPair)
				helpers.AssertTrue(t, resp.TokenPair.AccessToken != "")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := new(mocks.MockMemberRepository)
			tt.setupMocks(mockRepo)

			// Create services
			jwtService := auth.NewJWTService("test-secret-key", 24*time.Hour, 7*24*time.Hour, "test-issuer")
			passwordService := auth.NewPasswordService()

			// Create use case
			uc := NewLoginUseCase(mockRepo, passwordService, jwtService)

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
