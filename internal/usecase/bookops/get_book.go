package bookops

import (
	"context"

	"go.uber.org/zap"

	"library-service/internal/domain/book"
	"library-service/internal/infrastructure/log"
	"library-service/internal/infrastructure/store"
	"library-service/pkg/errors"
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
	logger := log.FromContext(ctx).Named("get_book_usecase").With(zap.String("id", req.ID))

	if req.ID == "" {
		return GetBookResponse{}, errors.ErrInvalidInput.WithDetails("field", "id")
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
		if errors.Is(err, store.ErrorNotFound) {
			logger.Warn("book not found")
			return GetBookResponse{}, errors.ErrBookNotFound.WithDetails("id", req.ID)
		}
		logger.Error("failed to get book from repository", zap.Error(err))
		return GetBookResponse{}, errors.ErrDatabase.Wrap(err)
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
		Name:    safeString(entity.Name),
		Genre:   safeString(entity.Genre),
		ISBN:    safeString(entity.ISBN),
		Authors: entity.Authors,
	}
}
