package author

// Author represents an author entity in the library system.
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
//   - UUID format for consistency with other entities
//   - Value type prevents nil pointer issues
//
// 2. FullName, Pseudonym, Specialty (*string, pointers):
//   - Nullable fields in database (optional information)
//   - Pointer distinguishes between "" (explicitly empty) and null (not provided)
//   - Critical for UPDATE operations (nil = don't change, value = update)
//
// DATABASE MAPPING:
//   - Uses snake_case for PostgreSQL convention (full_name, not fullName)
//   - `db:` tag for sqlx library
//   - `bson:` tag for future MongoDB support
//
// RELATIONSHIPS:
//   - Many-to-Many with Book entity
//   - Author IDs stored in Book.Authors field (denormalized)
//   - Full author objects fetched via AuthorRepository when needed
//
// CACHING STRATEGY:
//   - Authors are cached (read frequently, change infrequently)
//   - Cache key: author:{id}
//   - Cache invalidated on UPDATE/DELETE operations
type Author struct {
	// ID is the unique identifier for the author (UUID v4 format).
	// Required, immutable after creation.
	ID string `db:"id" bson:"_id"`

	// FullName is the complete legal name of the author (e.g., "J.K. Rowling").
	// Nullable: allows partial information during creation.
	FullName *string `db:"full_name" bson:"full_name"`

	// Pseudonym is the pen name used by the author (e.g., "George Orwell" for Eric Blair).
	// Nullable: most authors don't use pseudonyms.
	// Searchable field for book lookups.
	Pseudonym *string `db:"pseudonym" bson:"pseudonym"`

	// Specialty is the author's primary domain or genre (e.g., "Science Fiction", "Poetry").
	// Nullable: allows authors without defined specialty.
	// Used for filtering and recommendation systems.
	Specialty *string `db:"specialty" bson:"specialty"`
}

// New creates a new Author instance.
func New(req Request) Author {
	return Author{
		FullName:  &req.FullName,
		Pseudonym: &req.Pseudonym,
		Specialty: &req.Specialty,
	}
}
