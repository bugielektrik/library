package profile

import (
	"library-service/internal/pkg/errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/mock"

	"library-service/internal/members/domain"
	"library-service/internal/members/repository/mocks"
	"library-service/test/helpers"
)

func TestGetMemberProfileUseCase_Execute(t *testing.T) {
	tests := []struct {
		name         string
		request      GetMemberProfileRequest
		setupMocks   func(*mocks.MockMemberRepository)
		expectError  bool
		errorCheck   func(*testing.T, error)
		validateFunc func(*testing.T, GetMemberProfileResponse)
	}{
		{
			name: "successful retrieval of member profile",
			request: GetMemberProfileRequest{
				MemberID: "member-123",
			},
			setupMocks: func(repo *mocks.MockMemberRepository) {
				memberEntity := domain.Member{
					ID:           "member-123",
					Email:        "test@example.com",
					PasswordHash: "hash",
					FullName:     strPtr("Test User"),
					Role:         domain.RoleUser,
				}

				repo.On("Get", mock.Anything, "member-123").
					Return(memberEntity, nil).
					Once()
			},
			expectError: false,
			validateFunc: func(t *testing.T, resp GetMemberProfileResponse) {
				assert.Equal(t, "member-123", resp.Member.ID)
				assert.Equal(t, "test@example.com", resp.Member.Email)
				assert.Equal(t, "Test User", *resp.Member.FullName)
			},
		},
		{
			name: "successful retrieval of member with borrowed books",
			request: GetMemberProfileRequest{
				MemberID: "member-456",
			},
			setupMocks: func(repo *mocks.MockMemberRepository) {
				memberEntity := domain.Member{
					ID:           "member-456",
					Email:        "test@example.com",
					PasswordHash: "hash",
					FullName:     strPtr("Test User"),
					Role:         domain.RoleUser,
					Books:        []string{"book-1", "book-2", "book-3"},
				}

				repo.On("Get", mock.Anything, "member-456").
					Return(memberEntity, nil).
					Once()
			},
			expectError: false,
			validateFunc: func(t *testing.T, resp GetMemberProfileResponse) {
				assert.Equal(t, "member-456", resp.Member.ID)
				assert.Len(t, resp.Member.Books, 3)
			},
		},
		{
			name: "member not found",
			request: GetMemberProfileRequest{
				MemberID: "non-existent-member",
			},
			setupMocks: func(repo *mocks.MockMemberRepository) {
				repo.On("Get", mock.Anything, "non-existent-member").
					Return(domain.Member{}, errors.ErrNotFound).
					Once()
			},
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.ErrorContains(t, err, "not found")
			},
		},
		{
			name: "repository error during retrieval",
			request: GetMemberProfileRequest{
				MemberID: "member-error",
			},
			setupMocks: func(repo *mocks.MockMemberRepository) {
				repo.On("Get", mock.Anything, "member-error").
					Return(domain.Member{}, errors.ErrDatabase).
					Once()
			},
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				assert.Error(t, err)
				assert.ErrorContains(t, err, "not found")
			},
		},
		{
			name: "member with admin role",
			request: GetMemberProfileRequest{
				MemberID: "admin-user",
			},
			setupMocks: func(repo *mocks.MockMemberRepository) {
				memberEntity := domain.Member{
					ID:           "admin-user",
					Email:        "admin@library.com",
					PasswordHash: "hash",
					FullName:     strPtr("Admin User"),
					Role:         domain.RoleAdmin,
				}

				repo.On("Get", mock.Anything, "admin-user").
					Return(memberEntity, nil).
					Once()
			},
			expectError: false,
			validateFunc: func(t *testing.T, resp GetMemberProfileResponse) {
				assert.Equal(t, "admin-user", resp.Member.ID)
				assert.Equal(t, domain.RoleAdmin, resp.Member.Role)
			},
		},
		{
			name: "member with no full name (nullable field)",
			request: GetMemberProfileRequest{
				MemberID: "member-no-name",
			},
			setupMocks: func(repo *mocks.MockMemberRepository) {
				memberEntity := domain.Member{
					ID:           "member-no-name",
					Email:        "test@example.com",
					PasswordHash: "hash",
					Role:         domain.RoleUser,
				}
				memberEntity.FullName = nil // Explicitly set to nil

				repo.On("Get", mock.Anything, "member-no-name").
					Return(memberEntity, nil).
					Once()
			},
			expectError: false,
			validateFunc: func(t *testing.T, resp GetMemberProfileResponse) {
				assert.Equal(t, "member-no-name", resp.Member.ID)
				assert.Nil(t, resp.Member.FullName)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := new(mocks.MockMemberRepository)
			tt.setupMocks(mockRepo)

			// Create use case
			uc := NewGetMemberProfileUseCase(mockRepo)

			// Execute with test context
			ctx := helpers.TestContext(t)
			result, err := uc.Execute(ctx, tt.request)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				require.NoError(t, err)
				if tt.validateFunc != nil {
					tt.validateFunc(t, result)
				}
			}

			// Verify mock expectations
			mockRepo.AssertExpectations(t)
		})
	}
}
