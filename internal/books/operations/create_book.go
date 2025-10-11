package operations

import (
	"context"

	"go.uber.org/zap"

	"library-service/internal/books/domain/book"
	"library-service/pkg/errors"
	"library-service/pkg/logutil"
	"library-service/pkg/strutil"
	"library-service/pkg/validation"
)

// CreateBookRequest represents the input for creating a book
type CreateBookRequest struct {
	Name    string
	Genre   string
	ISBN    string
	Authors []string
}

// Validate validates the CreateBookRequest
func (r CreateBookRequest) Validate() error {
	// Validate required fields
	if err := validation.RequiredString(r.Name, "Name"); err != nil {
		return err
	}

	if err := validation.RequiredString(r.Genre, "Genre"); err != nil {
		return err
	}

	if err := validation.RequiredString(r.ISBN, "ISBN"); err != nil {
		return err
	}

	// Basic ISBN length check (without strutil dependency)
	// Remove dashes and spaces for length check
	cleanISBN := ""
	for _, ch := range r.ISBN {
		if (ch >= '0' && ch <= '9') || (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') {
			cleanISBN += string(ch)
		}
	}

	if len(cleanISBN) != 10 && len(cleanISBN) != 13 {
		return errors.ErrValidation.
			WithDetails("field", "ISBN").
			WithDetails("reason", "invalid format").
			WithDetails("expected", "10 or 13 characters").
			WithDetails("actual", len(cleanISBN))
	}

	// Validate authors list
	if err := validation.RequiredSlice(r.Authors, "Authors"); err != nil {
		return err
	}

	// Validate each author name is not empty
	for i, author := range r.Authors {
		if author == "" {
			return errors.ErrValidation.
				WithDetails("field", "Authors").
				WithDetails("reason", "empty author name").
				WithDetails("index", i)
		}
	}

	return nil
}

// CreateBookResponse represents the output of creating a book
type CreateBookResponse struct {
	ID      string
	Name    string
	Genre   string
	ISBN    string
	Authors []string
}

// CreateBookUseCase handles the creation of a new book.
//
// Architecture Pattern: Standard CRUD use case with repository, cache, and domain service.
//
// See Also:
//   - Similar pattern: internal/usecase/authops/register.go (registration flow)
//   - Domain layer: internal/books/domain/book/service.go (ISBN validation)
//   - HTTP handler: internal/books/http/crud.go
//   - Repository: internal/books/repository/book.go
//   - ADR: .claude/adr/002-clean-architecture-boundaries.md (layer rules)
//   - Test: internal/books/operations/create_book_test.go
type CreateBookUseCase struct {
	bookRepo    book.Repository
	bookCache   book.Cache
	bookService *book.Service
}

// NewCreateBookUseCase creates a new instance of CreateBookUseCase
func NewCreateBookUseCase(bookRepo book.Repository, bookCache book.Cache, bookService *book.Service) *CreateBookUseCase {
	return &CreateBookUseCase{
		bookRepo:    bookRepo,
		bookCache:   bookCache,
		bookService: bookService,
	}
}

// Execute creates a new book in the system
func (uc *CreateBookUseCase) Execute(ctx context.Context, req CreateBookRequest) (CreateBookResponse, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return CreateBookResponse{}, err
	}

	// Create logger with use case context
	logger := logutil.UseCaseLogger(ctx, "book", "create")
	logger.Debug("creating book",
		zap.String("isbn", req.ISBN),
		zap.String("name", req.Name),
		zap.Int("authors_count", len(req.Authors)),
	)

	// Create book entity from request
	bookEntity := book.New(book.Request{
		Name:    req.Name,
		Genre:   req.Genre,
		ISBN:    req.ISBN,
		Authors: req.Authors,
	})

	// Validate book using domain service
	if err := uc.bookService.Validate(bookEntity); err != nil {
		logger.Warn("validation failed", zap.Error(err))
		return CreateBookResponse{}, err
	}

	// Check if book with ISBN already exists
	exists, err := uc.bookRepo.Get(ctx, req.ISBN)
	if err == nil && exists.ID != "" {
		logger.Warn("book with ISBN already exists", zap.String("existing_id", exists.ID))
		return CreateBookResponse{}, errors.AlreadyExists("book", "ISBN", req.ISBN)
	}

	// Save to repository
	id, err := uc.bookRepo.Add(ctx, bookEntity)
	if err != nil {
		logger.Error("failed to add book to repository", zap.Error(err))
		return CreateBookResponse{}, errors.Database("create book", err)
	}
	bookEntity.ID = id

	// Cache the new book
	if err := uc.bookCache.Set(ctx, id, bookEntity); err != nil {
		logger.Warn("failed to cache book", zap.Error(err))
		// Non-critical error, continue
	}

	logger.Info("book created successfully", zap.String("id", id))

	return uc.toResponse(bookEntity), nil
}

// toResponse converts book to response
func (uc *CreateBookUseCase) toResponse(b book.Book) CreateBookResponse {
	return CreateBookResponse{
		ID:      b.ID,
		Name:    strutil.SafeString(b.Name),
		Genre:   strutil.SafeString(b.Genre),
		ISBN:    strutil.SafeString(b.ISBN),
		Authors: b.Authors,
	}
}
