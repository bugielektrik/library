package member

// Member represents a member in the system.
type Member struct {
	// ID is the unique identifier for the member.
	ID string `db:"id" bson:"_id"`

	// FullName is the full name of the member.
	FullName *string `db:"full_name" bson:"full_name"`

	// Books is a list of book IDs that the member has borrowed.
	Books []string `db:"books" bson:"books"`
}

// New creates a new Member instance.
func New(req Request) Member {
	return Member{
		FullName: &req.FullName,
		Books:    req.Books,
	}
}
