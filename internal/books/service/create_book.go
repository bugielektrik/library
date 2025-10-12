package service

import (
	"context"
	errors2 "library-service/internal/pkg/errors"
	"library-service/internal/pkg/logutil"
	"library-service/internal/pkg/strutil"

	"go.uber.org/zap"

	"library-service/internal/books/domain/book"
)

// Validator defines the validation interface
type Validator interface {
	Validate(i interface{}) error
}

// CreateBookRequest represents the input for creating a book
type CreateBookRequest struct {
	Name    string   `validate:"required,min=1,max=200"`
	Genre   string   `validate:"required,min=1,max=100"`
	ISBN    string   `validate:"required,isbn10|isbn13"`
	Authors []string `validate:"required,min=1,dive,required"`
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
//   - HTTP handler: internal/books/handler/crud.go
//   - Repository: internal/books/repository/book.go
//   - ADR: .claude/adr/002-clean-architecture-boundaries.md (layer rules)
//   - Test: internal/books/service/create_book_test.go
type CreateBookUseCase struct {
	bookRepo    book.Repository
	bookCache   book.Cache
	bookService *book.Service
	validator   Validator
}

// NewCreateBookUseCase creates a new instance of CreateBookUseCase
func NewCreateBookUseCase(bookRepo book.Repository, bookCache book.Cache, bookService *book.Service, validator Validator) *CreateBookUseCase {
	return &CreateBookUseCase{
		bookRepo:    bookRepo,
		bookCache:   bookCache,
		bookService: bookService,
		validator:   validator,
	}
}

// Execute creates a new book in the system
func (uc *CreateBookUseCase) Execute(ctx context.Context, req CreateBookRequest) (CreateBookResponse, error) {
	// Validate request using go-playground/validator
	if err := uc.validator.Validate(req); err != nil {
		return CreateBookResponse{}, errors2.ErrValidation.Wrap(err)
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
		return CreateBookResponse{}, errors2.AlreadyExists("book", "ISBN", req.ISBN)
	}

	// Save to repository
	id, err := uc.bookRepo.Add(ctx, bookEntity)
	if err != nil {
		logger.Error("failed to add book to repository", zap.Error(err))
		return CreateBookResponse{}, errors2.Database("create book", err)
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
