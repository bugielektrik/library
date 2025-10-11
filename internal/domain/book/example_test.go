package book_test

import (
	"fmt"

	"library-service/internal/domain/book"
)

// Example shows basic usage of the book domain service
func Example() {
	svc := book.NewService()

	// Validate an ISBN-13
	err := svc.ValidateISBN("978-0-306-40615-7")
	fmt.Println(err == nil)
	// Output: true
}

// ExampleService_ValidateISBN demonstrates ISBN-13 validation
func ExampleService_ValidateISBN() {
	svc := book.NewService()

	// Valid ISBN-13 with hyphens
	err := svc.ValidateISBN("978-0-132-35088-4")
	fmt.Println("Valid ISBN-13:", err == nil)

	// Valid ISBN-13 without hyphens
	err = svc.ValidateISBN("9780132350884")
	fmt.Println("Without hyphens:", err == nil)

	// Invalid checksum
	err = svc.ValidateISBN("978-0-132-35088-5")
	fmt.Println("Invalid checksum:", err != nil)

	// Output:
	// Valid ISBN-13: true
	// Without hyphens: true
	// Invalid checksum: true
}

// ExampleService_ValidateISBN_isbn10 demonstrates ISBN-10 validation
func ExampleService_ValidateISBN_isbn10() {
	svc := book.NewService()

	// Valid ISBN-10
	err := svc.ValidateISBN("0-306-40615-2")
	fmt.Println("Valid ISBN-10:", err == nil)

	// Valid ISBN-10 with X checksum
	err = svc.ValidateISBN("043942089X")
	fmt.Println("With X checksum:", err == nil)

	// Output:
	// Valid ISBN-10: true
	// With X checksum: true
}

// ExampleService_NormalizeISBN demonstrates ISBN normalization
func ExampleService_NormalizeISBN() {
	svc := book.NewService()

	// Normalize ISBN-13 (removes hyphens)
	normalized, _ := svc.NormalizeISBN("978-0-13-235088-4")
	fmt.Println("ISBN-13:", normalized)

	// Normalize ISBN-10 (converts to ISBN-13)
	normalized, _ = svc.NormalizeISBN("0-306-40615-2")
	fmt.Println("ISBN-10 to 13:", normalized)

	// Output:
	// ISBN-13: 9780132350884
	// ISBN-10 to 13: 9780306406157
}

// ExampleService_Validate demonstrates complete book validation
func ExampleService_Validate() {
	svc := book.NewService()

	name := "Clean Code: A Handbook of Agile Software Craftsmanship"
	genre := "Technology"
	isbn := "978-0132350884"

	b := book.Book{
		Name:    &name,
		Genre:   &genre,
		ISBN:    &isbn,
		Authors: []string{"Robert C. Martin"},
	}

	err := svc.Validate(b)
	fmt.Println("Valid book:", err == nil)

	// Output:
	// Valid book: true
}

// ExampleService_Validate_missingFields demonstrates validation errors
func ExampleService_Validate_missingFields() {
	svc := book.NewService()

	// Book without authors
	name := "Some Book"
	genre := "Fiction"
	isbn := "978-0132350884"

	b := book.Book{
		Name:    &name,
		Genre:   &genre,
		ISBN:    &isbn,
		Authors: []string{}, // Empty authors
	}

	err := svc.Validate(b)
	fmt.Println("Missing authors error:", err != nil)

	// Output:
	// Missing authors error: true
}

// ExampleService_CanBookBeDeleted demonstrates deletion validation
func ExampleService_CanBookBeDeleted() {
	svc := book.NewService()

	b := book.Book{
		ID: "book-123",
	}

	err := svc.CanBookBeDeleted(b)
	fmt.Println("Can delete:", err == nil)

	// Output:
	// Can delete: true
}

// Example_isbnConversion demonstrates ISBN-10 to ISBN-13 conversion
func Example_isbnConversion() {
	svc := book.NewService()

	// ISBN-10 for "The Pragmatic Programmer"
	isbn10 := "020161622X"

	// Normalize converts to ISBN-13
	isbn13, err := svc.NormalizeISBN(isbn10)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println("Original ISBN-10:", isbn10)
	fmt.Println("Converted ISBN-13:", isbn13)
	fmt.Println("Starts with 978:", isbn13[:3] == "978")

	// Output:
	// Original ISBN-10: 020161622X
	// Converted ISBN-13: 9780201616224
	// Starts with 978: true
}

// Example_multipleFormats demonstrates handling different ISBN formats
func Example_multipleFormats() {
	svc := book.NewService()

	isbns := []string{
		"978-0-306-40615-7", // ISBN-13 with hyphens
		"9780306406157",     // ISBN-13 without hyphens
		"0-306-40615-2",     // ISBN-10 with hyphens
		"0306406152",        // ISBN-10 without hyphens
		"978 0 306 40615 7", // ISBN-13 with spaces
	}

	for _, isbn := range isbns {
		normalized, err := svc.NormalizeISBN(isbn)
		if err != nil {
			fmt.Printf("Error for %s: %v\n", isbn, err)
			continue
		}
		fmt.Println(normalized)
	}

	// Output:
	// 9780306406157
	// 9780306406157
	// 9780306406157
	// 9780306406157
	// 9780306406157
}
