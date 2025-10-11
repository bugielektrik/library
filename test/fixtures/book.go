package fixtures

import (
	"library-service/internal/domain/book"
	"library-service/internal/usecase/bookops"
	"library-service/pkg/strutil"
)

// ValidBook returns a valid book entity for testing
func ValidBook() book.Book {
	return book.Book{
		ID:      "550e8400-e29b-41d4-a716-446655440000",
		Name:    strutil.SafeStringPtr("Clean Code: A Handbook of Agile Software Craftsmanship"),
		Genre:   strutil.SafeStringPtr("Software Engineering"),
		ISBN:    strutil.SafeStringPtr("978-0132350884"),
		Authors: []string{"550e8400-e29b-41d4-a716-446655440001"},
	}
}

// ValidBookWithMultipleAuthors returns a book with multiple authors
func ValidBookWithMultipleAuthors() book.Book {
	return book.Book{
		ID:    "550e8400-e29b-41d4-a716-446655440002",
		Name:  strutil.SafeStringPtr("Design Patterns: Elements of Reusable Object-Oriented Software"),
		Genre: strutil.SafeStringPtr("Software Engineering"),
		ISBN:  strutil.SafeStringPtr("978-0201633610"),
		Authors: []string{
			"550e8400-e29b-41d4-a716-446655440001",
			"550e8400-e29b-41d4-a716-446655440002",
			"550e8400-e29b-41d4-a716-446655440003",
		},
	}
}

// MinimalBook returns a book with only required fields
func MinimalBook() book.Book {
	return book.Book{
		ID:      "550e8400-e29b-41d4-a716-446655440003",
		Name:    strutil.SafeStringPtr("Test Book"),
		Genre:   nil,
		ISBN:    strutil.SafeStringPtr("978-0000000000"),
		Authors: []string{},
	}
}

// BookWithInvalidISBN returns a book with invalid ISBN for testing validation
func BookWithInvalidISBN() book.Book {
	return book.Book{
		ID:      "550e8400-e29b-41d4-a716-446655440004",
		Name:    strutil.SafeStringPtr("Invalid ISBN Book"),
		Genre:   strutil.SafeStringPtr("Test"),
		ISBN:    strutil.SafeStringPtr("invalid-isbn"),
		Authors: []string{},
	}
}

// CreateBookRequest returns a valid create book request
func CreateBookRequest() bookops.CreateBookRequest {
	return bookops.CreateBookRequest{
		Name:    "The Pragmatic Programmer",
		Genre:   "Software Engineering",
		ISBN:    "978-0135957059",
		Authors: []string{"550e8400-e29b-41d4-a716-446655440001"},
	}
}

// UpdateBookRequest returns a valid update book request
func UpdateBookRequest() bookops.UpdateBookRequest {
	return bookops.UpdateBookRequest{
		ID:      "550e8400-e29b-41d4-a716-446655440000",
		Name:    strutil.SafeStringPtr("Clean Code: Updated Edition"),
		Genre:   strutil.SafeStringPtr("Software Engineering"),
		ISBN:    strutil.SafeStringPtr("978-0132350884"),
		Authors: []string{"550e8400-e29b-41d4-a716-446655440001"},
	}
}

// BookResponse returns a valid book response
func BookResponse() book.Response {
	return book.Response{
		ID:      "550e8400-e29b-41d4-a716-446655440000",
		Name:    "Clean Code: A Handbook of Agile Software Craftsmanship",
		Genre:   "Software Engineering",
		ISBN:    "978-0132350884",
		Authors: []string{"550e8400-e29b-41d4-a716-446655440001"},
	}
}

// BookResponses returns a slice of book responses for testing list operations
func BookResponses() []book.Response {
	return []book.Response{
		{
			ID:      "550e8400-e29b-41d4-a716-446655440000",
			Name:    "Clean Code",
			Genre:   "Software Engineering",
			ISBN:    "978-0132350884",
			Authors: []string{"550e8400-e29b-41d4-a716-446655440001"},
		},
		{
			ID:      "550e8400-e29b-41d4-a716-446655440002",
			Name:    "Design Patterns",
			Genre:   "Software Engineering",
			ISBN:    "978-0201633610",
			Authors: []string{"550e8400-e29b-41d4-a716-446655440001"},
		},
		{
			ID:      "550e8400-e29b-41d4-a716-446655440005",
			Name:    "Refactoring",
			Genre:   "Software Engineering",
			ISBN:    "978-0134757599",
			Authors: []string{"550e8400-e29b-41d4-a716-446655440002"},
		},
	}
}

// BookForCreate returns a book entity suitable for repository creation (no ID)
func BookForCreate() book.Book {
	return book.Book{
		Name:    strutil.SafeStringPtr("New Test Book"),
		Genre:   strutil.SafeStringPtr("Fiction"),
		ISBN:    strutil.SafeStringPtr("978-1234567890"),
		Authors: []string{},
	}
}

// BookUpdate returns partial book data for update operations
func BookUpdate() book.Book {
	return book.Book{
		Name:  strutil.SafeStringPtr("Updated Book Title"),
		Genre: strutil.SafeStringPtr("Technology"),
	}
}
