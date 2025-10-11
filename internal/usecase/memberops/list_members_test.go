package memberops

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"library-service/internal/adapters/repository/mocks"
	"library-service/internal/domain/member"
	"library-service/pkg/errors"
	"library-service/test/builders"
	"library-service/test/helpers"
)

func TestListMembersUseCase_Execute(t *testing.T) {
	tests := []struct {
		name         string
		setupMocks   func(*mocks.MockMemberRepository)
		expectError  bool
		errorCheck   func(*testing.T, error)
		validateFunc func(*testing.T, ListMembersResponse)
	}{
		{
			name: "successful list with multiple members",
			setupMocks: func(repo *mocks.MockMemberRepository) {
				now := time.Now()

				member1 := builders.Member().Build()
				member1.ID = "member-1"
				member1.Email = "alice@example.com"
				member1.FullName = strPtr("Alice Johnson")

				member2 := builders.Member().Build()
				member2.ID = "member-2"
				member2.Email = "bob@example.com"
				member2.FullName = strPtr("Bob Smith")

				member3 := builders.Member().Build()
				member3.ID = "member-3"
				member3.Email = "charlie@example.com"
				member3.FullName = strPtr("Charlie Brown")
				member3.Role = member.RoleAdmin
				member3.Books = []string{"book-1", "book-2"}
				loginTime := now.Add(-24 * time.Hour)
				member3.LastLoginAt = &loginTime

				// Repository returns 3 members
				repo.On("List", mock.Anything).
					Return([]member.Member{member1, member2, member3}, nil).
					Once()
			},
			expectError: false,
			validateFunc: func(t *testing.T, resp ListMembersResponse) {
				helpers.AssertEqual(t, 3, resp.Total)
				helpers.AssertEqual(t, 3, len(resp.Members))
				helpers.AssertEqual(t, "member-1", resp.Members[0].ID)
				helpers.AssertEqual(t, "member-2", resp.Members[1].ID)
				helpers.AssertEqual(t, "member-3", resp.Members[2].ID)
				helpers.AssertEqual(t, "alice@example.com", resp.Members[0].Email)
				helpers.AssertEqual(t, "Bob Smith", *resp.Members[1].FullName)
				helpers.AssertEqual(t, member.RoleAdmin, resp.Members[2].Role)
				helpers.AssertEqual(t, 2, len(resp.Members[2].Books))
			},
		},
		{
			name: "empty list (no members)",
			setupMocks: func(repo *mocks.MockMemberRepository) {
				// Repository returns empty slice
				repo.On("List", mock.Anything).
					Return([]member.Member{}, nil).
					Once()
			},
			expectError: false,
			validateFunc: func(t *testing.T, resp ListMembersResponse) {
				helpers.AssertEqual(t, 0, resp.Total)
				helpers.AssertEqual(t, 0, len(resp.Members))
			},
		},
		{
			name: "repository error during list",
			setupMocks: func(repo *mocks.MockMemberRepository) {
				// Repository returns error with empty slice
				repo.On("List", mock.Anything).
					Return([]member.Member{}, errors.ErrDatabase).
					Once()
			},
			expectError: true,
			errorCheck: func(t *testing.T, err error) {
				helpers.AssertError(t, err)
			},
		},
		{
			name: "single member in list",
			setupMocks: func(repo *mocks.MockMemberRepository) {
				member1 := builders.Member().Build()
				member1.ID = "member-single"
				member1.Email = "single@example.com"

				// Repository returns 1 member
				repo.On("List", mock.Anything).
					Return([]member.Member{member1}, nil).
					Once()
			},
			expectError: false,
			validateFunc: func(t *testing.T, resp ListMembersResponse) {
				helpers.AssertEqual(t, 1, resp.Total)
				helpers.AssertEqual(t, 1, len(resp.Members))
				helpers.AssertEqual(t, "member-single", resp.Members[0].ID)
				helpers.AssertEqual(t, "single@example.com", resp.Members[0].Email)
			},
		},
		{
			name: "list with members having different roles",
			setupMocks: func(repo *mocks.MockMemberRepository) {
				admin := builders.Member().Build()
				admin.ID = "admin-1"
				admin.Email = "admin@library.com"
				admin.Role = member.RoleAdmin

				user1 := builders.Member().Build()
				user1.ID = "user-1"
				user1.Email = "user1@example.com"
				user1.Role = member.RoleUser

				user2 := builders.Member().Build()
				user2.ID = "user-2"
				user2.Email = "user2@example.com"
				user2.Role = member.RoleUser

				// Repository returns members with different roles
				repo.On("List", mock.Anything).
					Return([]member.Member{admin, user1, user2}, nil).
					Once()
			},
			expectError: false,
			validateFunc: func(t *testing.T, resp ListMembersResponse) {
				helpers.AssertEqual(t, 3, resp.Total)
				helpers.AssertEqual(t, member.RoleAdmin, resp.Members[0].Role)
				helpers.AssertEqual(t, member.RoleUser, resp.Members[1].Role)
				helpers.AssertEqual(t, member.RoleUser, resp.Members[2].Role)
			},
		},
		{
			name: "list with members having borrowed books",
			setupMocks: func(repo *mocks.MockMemberRepository) {
				member1 := builders.Member().Build()
				member1.ID = "member-with-books"
				member1.Books = []string{"book-1", "book-2", "book-3"}

				member2 := builders.Member().Build()
				member2.ID = "member-no-books"
				member2.Books = []string{}

				// Repository returns members with different book counts
				repo.On("List", mock.Anything).
					Return([]member.Member{member1, member2}, nil).
					Once()
			},
			expectError: false,
			validateFunc: func(t *testing.T, resp ListMembersResponse) {
				helpers.AssertEqual(t, 2, resp.Total)
				helpers.AssertEqual(t, 3, len(resp.Members[0].Books))
				helpers.AssertEqual(t, 0, len(resp.Members[1].Books))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := new(mocks.MockMemberRepository)
			tt.setupMocks(mockRepo)

			// Create use case
			uc := NewListMembersUseCase(mockRepo)

			// Execute
			ctx := helpers.TestContext(t)
			result, err := uc.Execute(ctx, ListMembersRequest{})

			// Assert
			if tt.expectError {
				helpers.AssertError(t, err)
				if tt.errorCheck != nil {
					tt.errorCheck(t, err)
				}
			} else {
				helpers.AssertNoError(t, err)
				if tt.validateFunc != nil {
					tt.validateFunc(t, result)
				}
			}

			// Verify mock expectations
			mockRepo.AssertExpectations(t)
		})
	}
}
