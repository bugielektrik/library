package author

// Entity represents an author in the system.
type Entity struct {
	// ID is the unique identifier for the author.
	ID string `db:"id" bson:"_id"`

	// FullName is the full name of the author.
	FullName *string `db:"full_name" bson:"full_name"`

	// Pseudonym is the pseudonym of the author, if any.
	Pseudonym *string `db:"pseudonym" bson:"pseudonym"`

	// Specialty is the specialty of the author.
	Specialty *string `db:"specialty" bson:"specialty"`
}
