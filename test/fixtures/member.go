package fixtures

import (
	"library-service/internal/pkg/strutil"
	"time"

	"library-service/internal/members/domain"
)

// ValidMember returns a valid member entity for testing
func ValidMember() domain.Member {
	now := time.Now()
	return domain.Member{
		ID:           "b4101570-0a35-4dd3-b8f7-745d56013263",
		Email:        "john.doe@example.com",
		PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy",
		FullName:     strutil.SafeStringPtr("John Doe"),
		Role:         domain.RoleUser,
		Books:        []string{},
		CreatedAt:    now,
		UpdatedAt:    now,
		LastLoginAt:  &now,
	}
}

// AdminMember returns a member with admin role
func AdminMember() domain.Member {
	now := time.Now()
	return domain.Member{
		ID:           "a4101570-0a35-4dd3-b8f7-745d56013264",
		Email:        "admin@example.com",
		PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy",
		FullName:     strutil.SafeStringPtr("Admin User"),
		Role:         domain.RoleAdmin,
		Books:        []string{},
		CreatedAt:    now,
		UpdatedAt:    now,
		LastLoginAt:  &now,
	}
}

// MemberWithBooks returns a member with borrowed books
func MemberWithBooks() domain.Member {
	now := time.Now()
	return domain.Member{
		ID:           "c4101570-0a35-4dd3-b8f7-745d56013265",
		Email:        "reader@example.com",
		PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy",
		FullName:     strutil.SafeStringPtr("Active Reader"),
		Role:         domain.RoleUser,
		Books: []string{
			"550e8400-e29b-41d4-a716-446655440000",
			"550e8400-e29b-41d4-a716-446655440002",
		},
		CreatedAt:   now,
		UpdatedAt:   now,
		LastLoginAt: &now,
	}
}

// NewMember returns a member without last login (recently created)
func NewMember() domain.Member {
	now := time.Now()
	return domain.Member{
		ID:           "d4101570-0a35-4dd3-b8f7-745d56013266",
		Email:        "new.member@example.com",
		PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy",
		FullName:     strutil.SafeStringPtr("New Member"),
		Role:         domain.RoleUser,
		Books:        []string{},
		CreatedAt:    now,
		UpdatedAt:    now,
		LastLoginAt:  nil,
	}
}

// MemberResponse returns a valid member response
func MemberResponse() domain.Response {
	return domain.Response{
		ID:       "b4101570-0a35-4dd3-b8f7-745d56013263",
		Email:    "john.doe@example.com",
		FullName: "John Doe",
		Role:     "user",
	}
}

// MemberResponses returns a slice of member responses for testing list operations
func MemberResponses() []domain.Response {
	return []domain.Response{
		{
			ID:       "b4101570-0a35-4dd3-b8f7-745d56013263",
			Email:    "john.doe@example.com",
			FullName: "John Doe",
			Role:     "user",
		},
		{
			ID:       "a4101570-0a35-4dd3-b8f7-745d56013264",
			Email:    "admin@example.com",
			FullName: "Admin User",
			Role:     "admin",
		},
		{
			ID:       "c4101570-0a35-4dd3-b8f7-745d56013265",
			Email:    "reader@example.com",
			FullName: "Active Reader",
			Role:     "user",
		},
	}
}

// MemberForCreate returns a member entity suitable for repository creation (no ID)
func MemberForCreate() domain.Member {
	return domain.Member{
		Email:        "newmember@example.com",
		PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy",
		FullName:     strutil.SafeStringPtr("New Member"),
		Role:         domain.RoleUser,
		Books:        []string{},
	}
}

// MemberUpdate returns partial member data for update operations
func MemberUpdate() domain.Member {
	return domain.Member{
		FullName: strutil.SafeStringPtr("Updated Member Name"),
		Role:     domain.RoleAdmin,
	}
}

// Members returns a collection of sample members for batch testing
func Members() []domain.Member {
	return []domain.Member{
		{
			Email:        "member1@example.com",
			PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy",
			FullName:     strutil.SafeStringPtr("Member One"),
			Role:         domain.RoleUser,
			Books:        []string{},
		},
		{
			Email:        "member2@example.com",
			PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy",
			FullName:     strutil.SafeStringPtr("Member Two"),
			Role:         domain.RoleUser,
			Books:        []string{},
		},
		{
			Email:        "member3@example.com",
			PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy",
			FullName:     strutil.SafeStringPtr("Member Three"),
			Role:         domain.RoleUser,
			Books:        []string{},
		},
	}
}
