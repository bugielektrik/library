package service

import (
	"context"
	errors2 "library-service/internal/pkg/errors"
	"library-service/internal/pkg/logutil"
	"library-service/internal/pkg/strutil"

	"go.uber.org/zap"

	"library-service/internal/books/domain/book"
	"library-service/internal/infrastructure/store"
)

// GetBookRequest represents the input for getting a book
type GetBookRequest struct {
	ID string
}

// GetBookResponse represents the output of getting a book
type GetBookResponse struct {
	ID      string
	Name    string
	Genre   string
	ISBN    string
	Authors []string
}

// GetBookUseCase handles retrieving a book by ID
type GetBookUseCase struct {
	bookRepo  book.Repository
	bookCache book.Cache
}

// NewGetBookUseCase creates a new instance of GetBookUseCase
func NewGetBookUseCase(bookRepo book.Repository, bookCache book.Cache) *GetBookUseCase {
	return &GetBookUseCase{
		bookRepo:  bookRepo,
		bookCache: bookCache,
	}
}

// Execute retrieves a book from cache or repository
func (uc *GetBookUseCase) Execute(ctx context.Context, req GetBookRequest) (GetBookResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "book", "get")

	if req.ID == "" {
		return GetBookResponse{}, errors2.ValidationRequired("id")
	}

	// Try cache first
	bookEntity, err := uc.bookCache.Get(ctx, req.ID)
	if err == nil && bookEntity.ID != "" {
		logger.Debug("book found in cache")
		return uc.toResponse(bookEntity), nil
	}

	// Get from repository
	bookEntity, err = uc.bookRepo.Get(ctx, req.ID)
	if err != nil {
		if errors2.Is(err, store.ErrorNotFound) {
			logger.Warn("book not found")
			return GetBookResponse{}, errors2.NotFoundWithID("book", req.ID)
		}
		logger.Error("failed to get book from repository", zap.Error(err))
		return GetBookResponse{}, errors2.Database("database operation", err)
	}

	// Update cache
	if err := uc.bookCache.Set(ctx, req.ID, bookEntity); err != nil {
		logger.Warn("failed to cache book", zap.Error(err))
		// Non-critical, continue
	}

	logger.Debug("book retrieved successfully")
	return uc.toResponse(bookEntity), nil
}

// toResponse converts book entity to response
func (uc *GetBookUseCase) toResponse(entity book.Book) GetBookResponse {
	return GetBookResponse{
		ID:      entity.ID,
		Name:    strutil.SafeString(entity.Name),
		Genre:   strutil.SafeString(entity.Genre),
		ISBN:    strutil.SafeString(entity.ISBN),
		Authors: entity.Authors,
	}
}
