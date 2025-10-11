package builders

import (
	"library-service/internal/domain/book"
)

// BookBuilder provides a fluent interface for building Book test fixtures.
type BookBuilder struct {
	book book.Book
}

// NewBook creates a BookBuilder with sensible defaults.
func NewBook() *BookBuilder {
	name := "The Go Programming Language"
	genre := "Technology"
	isbn := "978-0134190440"
	return &BookBuilder{
		book: book.Book{
			ID:      "test-book-id",
			Name:    &name,
			Genre:   &genre,
			ISBN:    &isbn,
			Authors: []string{"author-1", "author-2"},
		},
	}
}

// WithID sets the book ID.
func (b *BookBuilder) WithID(id string) *BookBuilder {
	b.book.ID = id
	return b
}

// WithName sets the book name.
func (b *BookBuilder) WithName(name string) *BookBuilder {
	b.book.Name = &name
	return b
}

// WithGenre sets the genre.
func (b *BookBuilder) WithGenre(genre string) *BookBuilder {
	b.book.Genre = &genre
	return b
}

// WithISBN sets the ISBN.
func (b *BookBuilder) WithISBN(isbn string) *BookBuilder {
	b.book.ISBN = &isbn
	return b
}

// WithAuthors sets the authors.
func (b *BookBuilder) WithAuthors(authors ...string) *BookBuilder {
	b.book.Authors = authors
	return b
}

// WithSingleAuthor sets a single author.
func (b *BookBuilder) WithSingleAuthor(author string) *BookBuilder {
	b.book.Authors = []string{author}
	return b
}

// Build returns the constructed Book.
func (b *BookBuilder) Build() book.Book {
	return b.book
}
