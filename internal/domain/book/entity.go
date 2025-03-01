package book

// Entity represents a book in the system.
type Entity struct {
	// ID is the unique identifier for the book.
	ID string `db:"id" bson:"_id"`

	// Name is the name of the book.
	Name *string `db:"name" bson:"name"`

	// Genre is the genre of the book.
	Genre *string `db:"genre" bson:"genre"`

	// ISBN is the ISBN of the book.
	ISBN *string `db:"isbn" bson:"isbn"`

	// Authors is the list of author IDs associated with the book.
	Authors []string `db:"authors" bson:"authors"`
}
