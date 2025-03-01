package member

// Entity represents a member in the system.
type Entity struct {
	// ID is the unique identifier for the member.
	ID string `db:"id" bson:"_id"`

	// FullName is the full name of the member.
	FullName *string `db:"full_name" bson:"full_name"`

	// Books is a list of book IDs that the member has borrowed.
	Books []string `db:"books" bson:"books"`
}

// New creates a new Member instance.
func New(req Request) Entity {
	return Entity{
		FullName: &req.FullName,
		Books:    req.Books,
	}
}
