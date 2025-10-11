package builders

import (
	"time"

	"library-service/internal/members/domain"
)

// MemberBuilder provides a fluent interface for building Member test fixtures.
type MemberBuilder struct {
	member domain.Member
}

// NewMember creates a MemberBuilder with sensible defaults.
func NewMember() *MemberBuilder {
	fullName := "Test User"
	return &MemberBuilder{
		member: domain.Member{
			ID:        "test-member-id",
			Email:     "test@example.com",
			FullName:  &fullName,
			Role:      domain.RoleUser,
			CreatedAt: time.Now(),
		},
	}
}

// Member is a convenience function that creates a MemberBuilder with sensible defaults.
// This is an alias for NewMember for more fluent usage.
func Member() *MemberBuilder {
	return NewMember()
}

// WithID sets the member ID.
func (b *MemberBuilder) WithID(id string) *MemberBuilder {
	b.member.ID = id
	return b
}

// WithEmail sets the email.
func (b *MemberBuilder) WithEmail(email string) *MemberBuilder {
	b.member.Email = email
	return b
}

// WithFullName sets the full name.
func (b *MemberBuilder) WithFullName(fullName string) *MemberBuilder {
	b.member.FullName = &fullName
	return b
}

// WithRole sets the role.
func (b *MemberBuilder) WithRole(role domain.Role) *MemberBuilder {
	b.member.Role = role
	return b
}

// WithAdminRole sets the role to admin.
func (b *MemberBuilder) WithAdminRole() *MemberBuilder {
	b.member.Role = domain.RoleAdmin
	return b
}

// AsAdmin is an alias for WithAdminRole.
func (b *MemberBuilder) AsAdmin() *MemberBuilder {
	return b.WithAdminRole()
}

// AsUser sets the role to user.
func (b *MemberBuilder) AsUser() *MemberBuilder {
	b.member.Role = domain.RoleUser
	return b
}

// WithPasswordHash sets the password hash.
func (b *MemberBuilder) WithPasswordHash(hash string) *MemberBuilder {
	b.member.PasswordHash = hash
	return b
}

// WithBooks sets the member's borrowed books.
func (b *MemberBuilder) WithBooks(bookIDs ...string) *MemberBuilder {
	b.member.Books = bookIDs
	return b
}

// WithLastLoginAt sets the last login time.
func (b *MemberBuilder) WithLastLoginAt(t time.Time) *MemberBuilder {
	b.member.LastLoginAt = &t
	return b
}

// WithCreatedAt sets the created at time.
func (b *MemberBuilder) WithCreatedAt(t time.Time) *MemberBuilder {
	b.member.CreatedAt = t
	return b
}

// WithUpdatedAt sets the updated at time.
func (b *MemberBuilder) WithUpdatedAt(t time.Time) *MemberBuilder {
	b.member.UpdatedAt = t
	return b
}

// Build returns the constructed Member.
func (b *MemberBuilder) Build() domain.Member {
	return b.member
}

// BuildPtr returns a pointer to the constructed Member.
func (b *MemberBuilder) BuildPtr() *domain.Member {
	m := b.member
	return &m
}
