package fixtures

import (
	"library-service/internal/books/domain/author"
	"library-service/internal/pkg/strutil"
)

// ValidAuthor returns a valid author entity for testing
func ValidAuthor() author.Author {
	return author.Author{
		ID:        "550e8400-e29b-41d4-a716-446655440001",
		FullName:  strutil.SafeStringPtr("Robert C. Martin"),
		Pseudonym: strutil.SafeStringPtr("Uncle Bob"),
		Specialty: strutil.SafeStringPtr("Software Engineering"),
	}
}

// AuthorWithoutPseudonym returns an author without a pseudonym
func AuthorWithoutPseudonym() author.Author {
	return author.Author{
		ID:        "550e8400-e29b-41d4-a716-446655440002",
		FullName:  strutil.SafeStringPtr("Martin Fowler"),
		Pseudonym: nil,
		Specialty: strutil.SafeStringPtr("Software Engineering"),
	}
}

// MinimalAuthor returns an author with only required fields
func MinimalAuthor() author.Author {
	return author.Author{
		ID:        "550e8400-e29b-41d4-a716-446655440003",
		FullName:  strutil.SafeStringPtr("Test Author"),
		Pseudonym: nil,
		Specialty: nil,
	}
}

// AuthorResponse returns a valid author response
func AuthorResponse() author.Response {
	return author.Response{
		ID:        "550e8400-e29b-41d4-a716-446655440001",
		FullName:  "Robert C. Martin",
		Pseudonym: "Uncle Bob",
		Specialty: "Software Engineering",
	}
}

// AuthorResponses returns a slice of author responses for testing list operations
func AuthorResponses() []author.Response {
	return []author.Response{
		{
			ID:        "550e8400-e29b-41d4-a716-446655440001",
			FullName:  "Robert C. Martin",
			Pseudonym: "Uncle Bob",
			Specialty: "Software Engineering",
		},
		{
			ID:        "550e8400-e29b-41d4-a716-446655440002",
			FullName:  "Martin Fowler",
			Pseudonym: "",
			Specialty: "Software Engineering",
		},
		{
			ID:        "550e8400-e29b-41d4-a716-446655440003",
			FullName:  "Kent Beck",
			Pseudonym: "",
			Specialty: "Extreme Programming",
		},
	}
}

// AuthorForCreate returns an author entity suitable for repository creation (no ID)
func AuthorForCreate() author.Author {
	return author.Author{
		FullName:  strutil.SafeStringPtr("New Test Author"),
		Pseudonym: strutil.SafeStringPtr("Testy"),
		Specialty: strutil.SafeStringPtr("Testing"),
	}
}

// AuthorUpdate returns partial author data for update operations
func AuthorUpdate() author.Author {
	return author.Author{
		FullName:  strutil.SafeStringPtr("Updated Author Name"),
		Specialty: strutil.SafeStringPtr("Updated Specialty"),
	}
}

// Authors returns a collection of sample authors for batch testing
func Authors() []author.Author {
	return []author.Author{
		{FullName: strutil.SafeStringPtr("Author One"), Specialty: strutil.SafeStringPtr("Fiction")},
		{FullName: strutil.SafeStringPtr("Author Two"), Specialty: strutil.SafeStringPtr("Non-Fiction")},
		{FullName: strutil.SafeStringPtr("Author Three"), Specialty: strutil.SafeStringPtr("Science")},
		{FullName: strutil.SafeStringPtr("Author Four"), Specialty: strutil.SafeStringPtr("History")},
		{FullName: strutil.SafeStringPtr("Author Five"), Specialty: strutil.SafeStringPtr("Philosophy")},
	}
}

// Author returns a single sample author
func Author() author.Author {
	return author.Author{
		FullName:  strutil.SafeStringPtr("Sample Author"),
		Pseudonym: strutil.SafeStringPtr("SA"),
		Specialty: strutil.SafeStringPtr("General"),
	}
}
