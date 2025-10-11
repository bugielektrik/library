package book

// Book represents a book entity in the library system.
//
// TYPE HIERARCHY:
//   - This is a DOMAIN ENTITY (pure business object)
//   - Has NO external dependencies (no database, HTTP, frameworks)
//   - Used across all layers: Use Case → Adapter → Infrastructure
//
// FIELD DESIGN DECISIONS:
//
// 1. ID (string, NOT pointer):
//   - Always required, never null
//   - UUID format enforced at creation time
//   - Value type for consistency and safety
//
// 2. Name, Genre, ISBN (*string, pointers):
//   - Nullable in database (optional fields)
//   - Pointer allows distinguishing between empty string "" and null
//   - Required for UPDATE operations (nil = don't update, value = update)
//
// 3. Authors ([]string, NOT pointer to slice):
//   - Empty slice [] represents "no authors" (valid state)
//   - nil slice and empty slice treated identically by database
//   - Simpler to work with (no nil checks needed)
//
// DATABASE MAPPING:
//   - `db:"field"` tags for PostgreSQL (sqlx library)
//   - `bson:"field"` tags for MongoDB (future compatibility)
//   - MongoDB uses "_id" for primary key, PostgreSQL uses "id"
//
// RELATIONSHIPS:
//   - Authors: Many-to-Many with Author entity
//   - Stored as array of author IDs (denormalized for read performance)
//   - Author details fetched separately via AuthorRepository.GetByIDs()
//
// VALIDATION:
//   - ISBN validation in book.Service.ValidateISBN()
//   - Name/Genre length limits enforced at HTTP layer (validator)
//   - Business rules kept separate from entity definition
type Book struct {
	// ID is the unique identifier for the book (UUID v4 format).
	// Required, immutable after creation.
	ID string `db:"id" bson:"_id"`

	// Name is the title of the book (e.g., "The Great Gatsby").
	// Nullable: nil during creation (auto-generated), *string afterward.
	Name *string `db:"name" bson:"name"`

	// Genre is the literary category (e.g., "fiction", "non-fiction", "science").
	// Nullable: allows books without assigned genre.
	Genre *string `db:"genre" bson:"genre"`

	// ISBN is the International Standard Book Number (13-digit format).
	// Nullable: allows books without ISBN (e.g., unpublished manuscripts).
	// Validated by book.Service.ValidateISBN() for checksum correctness.
	ISBN *string `db:"isbn" bson:"isbn"`

	// Authors is the list of author UUIDs associated with this book.
	// Empty slice indicates book has no authors assigned yet.
	// Many-to-Many relationship: one book can have multiple authors.
	Authors []string `db:"authors" bson:"authors"`
}

// New creates a new Book instance.
func New(req Request) Book {
	return Book{
		Name:    &req.Name,
		Genre:   &req.Genre,
		ISBN:    &req.ISBN,
		Authors: req.Authors,
	}
}
