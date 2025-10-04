package member

import (
	"time"
)

// Role represents the role of a member in the system.
type Role string

const (
	// RoleUser is a regular member who can borrow books
	RoleUser Role = "user"
	// RoleAdmin has full access to manage the library
	RoleAdmin Role = "admin"
)

// Member represents a member in the system.
type Member struct {
	// ID is the unique identifier for the member.
	ID string `db:"id" bson:"_id"`

	// Email is the unique email address used for authentication.
	Email string `db:"email" bson:"email"`

	// PasswordHash is the bcrypt hash of the member's password.
	PasswordHash string `db:"password_hash" bson:"password_hash"`

	// FullName is the full name of the member.
	FullName *string `db:"full_name" bson:"full_name"`

	// Role defines the member's access level (user or admin).
	Role Role `db:"role" bson:"role"`

	// Books is a list of book IDs that the member has borrowed.
	Books []string `db:"books" bson:"books"`

	// CreatedAt is the timestamp when the member was created.
	CreatedAt time.Time `db:"created_at" bson:"created_at"`

	// UpdatedAt is the timestamp when the member was last updated.
	UpdatedAt time.Time `db:"updated_at" bson:"updated_at"`

	// LastLoginAt is the timestamp of the member's last login.
	LastLoginAt *time.Time `db:"last_login_at" bson:"last_login_at"`
}

// New creates a new Member instance.
func New(req Request) Member {
	now := time.Now()
	return Member{
		Email:     req.Email,
		FullName:  &req.FullName,
		Role:      RoleUser, // Default role
		Books:     req.Books,
		CreatedAt: now,
		UpdatedAt: now,
	}
}
