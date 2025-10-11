package fixtures

import (
	"time"

	"library-service/internal/domain/member"
	"library-service/pkg/strutil"
)

// ValidMember returns a valid member entity for testing
func ValidMember() member.Member {
	now := time.Now()
	return member.Member{
		ID:           "b4101570-0a35-4dd3-b8f7-745d56013263",
		Email:        "john.doe@example.com",
		PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy",
		FullName:     strutil.SafeStringPtr("John Doe"),
		Role:         member.RoleUser,
		Books:        []string{},
		CreatedAt:    now,
		UpdatedAt:    now,
		LastLoginAt:  &now,
	}
}

// AdminMember returns a member with admin role
func AdminMember() member.Member {
	now := time.Now()
	return member.Member{
		ID:           "a4101570-0a35-4dd3-b8f7-745d56013264",
		Email:        "admin@example.com",
		PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy",
		FullName:     strutil.SafeStringPtr("Admin User"),
		Role:         member.RoleAdmin,
		Books:        []string{},
		CreatedAt:    now,
		UpdatedAt:    now,
		LastLoginAt:  &now,
	}
}

// MemberWithBooks returns a member with borrowed books
func MemberWithBooks() member.Member {
	now := time.Now()
	return member.Member{
		ID:           "c4101570-0a35-4dd3-b8f7-745d56013265",
		Email:        "reader@example.com",
		PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy",
		FullName:     strutil.SafeStringPtr("Active Reader"),
		Role:         member.RoleUser,
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
func NewMember() member.Member {
	now := time.Now()
	return member.Member{
		ID:           "d4101570-0a35-4dd3-b8f7-745d56013266",
		Email:        "new.member@example.com",
		PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy",
		FullName:     strutil.SafeStringPtr("New Member"),
		Role:         member.RoleUser,
		Books:        []string{},
		CreatedAt:    now,
		UpdatedAt:    now,
		LastLoginAt:  nil,
	}
}

// MemberResponse returns a valid member response
func MemberResponse() member.Response {
	return member.Response{
		ID:       "b4101570-0a35-4dd3-b8f7-745d56013263",
		Email:    "john.doe@example.com",
		FullName: "John Doe",
		Role:     "user",
	}
}

// MemberResponses returns a slice of member responses for testing list operations
func MemberResponses() []member.Response {
	return []member.Response{
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
func MemberForCreate() member.Member {
	return member.Member{
		Email:        "newmember@example.com",
		PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy",
		FullName:     strutil.SafeStringPtr("New Member"),
		Role:         member.RoleUser,
		Books:        []string{},
	}
}

// MemberUpdate returns partial member data for update operations
func MemberUpdate() member.Member {
	return member.Member{
		FullName: strutil.SafeStringPtr("Updated Member Name"),
		Role:     member.RoleAdmin,
	}
}

// Members returns a collection of sample members for batch testing
func Members() []member.Member {
	return []member.Member{
		{
			Email:        "member1@example.com",
			PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy",
			FullName:     strutil.SafeStringPtr("Member One"),
			Role:         member.RoleUser,
			Books:        []string{},
		},
		{
			Email:        "member2@example.com",
			PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy",
			FullName:     strutil.SafeStringPtr("Member Two"),
			Role:         member.RoleUser,
			Books:        []string{},
		},
		{
			Email:        "member3@example.com",
			PasswordHash: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy",
			FullName:     strutil.SafeStringPtr("Member Three"),
			Role:         member.RoleUser,
			Books:        []string{},
		},
	}
}
