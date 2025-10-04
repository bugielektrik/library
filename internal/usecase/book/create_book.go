package book

import (
	"context"

	"go.uber.org/zap"

	"library-service/internal/domain/book"
	"library-service/pkg/errors"

	log "library-service/internal/infrastructure/logger"
)

// CreateBookRequest represents the input for creating a book
type CreateBookRequest struct {
	Name    string
	Genre   string
	ISBN    string
	Authors []string
}

// CreateBookResponse represents the output of creating a book
type CreateBookResponse struct {
	ID      string
	Name    string
	Genre   string
	ISBN    string
	Authors []string
}

// CreateBookUseCase handles the creation of a new book
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
	logger := log.FromContext(ctx).Named("create_book_usecase").With(
		zap.String("isbn", req.ISBN),
		zap.String("name", req.Name),
	)

	// Create book entity from request
	bookEntity := book.New(book.Request{
		Name:    req.Name,
		Genre:   req.Genre,
		ISBN:    req.ISBN,
		Authors: req.Authors,
	})

	// Validate book using domain service
	if err := uc.bookService.ValidateBook(bookEntity); err != nil {
		logger.Warn("validation failed", zap.Error(err))
		return CreateBookResponse{}, err
	}

	// Check if book with ISBN already exists
	exists, err := uc.bookRepo.Get(ctx, req.ISBN)
	if err == nil && exists.ID != "" {
		logger.Warn("book with ISBN already exists", zap.String("existing_id", exists.ID))
		return CreateBookResponse{}, errors.ErrBookAlreadyExists.WithDetails("isbn", req.ISBN)
	}

	// Save to repository
	id, err := uc.bookRepo.Add(ctx, bookEntity)
	if err != nil {
		logger.Error("failed to add book to repository", zap.Error(err))
		return CreateBookResponse{}, errors.ErrDatabase.Wrap(err)
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

// toResponse converts book entity to response
func (uc *CreateBookUseCase) toResponse(entity book.Entity) CreateBookResponse {
	return CreateBookResponse{
		ID:      entity.ID,
		Name:    safeString(entity.Name),
		Genre:   safeString(entity.Genre),
		ISBN:    safeString(entity.ISBN),
		Authors: entity.Authors,
	}
}

func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
