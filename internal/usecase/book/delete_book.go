package book

import (
	"context"

	"go.uber.org/zap"

	"library-service/internal/domain/book"
	store "library-service/internal/infrastructure/database"
	log "library-service/internal/infrastructure/logger"
	"library-service/pkg/errors"
)

// DeleteBookRequest represents the input for deleting a book
type DeleteBookRequest struct {
	ID string
}

// DeleteBookUseCase handles deleting a book
type DeleteBookUseCase struct {
	bookRepo  book.Repository
	bookCache book.Cache
}

// NewDeleteBookUseCase creates a new instance of DeleteBookUseCase
func NewDeleteBookUseCase(bookRepo book.Repository, bookCache book.Cache) *DeleteBookUseCase {
	return &DeleteBookUseCase{
		bookRepo:  bookRepo,
		bookCache: bookCache,
	}
}

// Execute deletes a book from the system
func (uc *DeleteBookUseCase) Execute(ctx context.Context, req DeleteBookRequest) error {
	logger := log.FromContext(ctx).Named("delete_book_usecase").With(zap.String("id", req.ID))

	if req.ID == "" {
		return errors.ErrInvalidInput.WithDetails("field", "id")
	}

	// Delete from repository
	err := uc.bookRepo.Delete(ctx, req.ID)
	if err != nil {
		if errors.Is(err, store.ErrorNotFound) {
			logger.Warn("book not found")
			return errors.ErrBookNotFound.WithDetails("id", req.ID)
		}
		logger.Error("failed to delete book", zap.Error(err))
		return errors.ErrDatabase.Wrap(err)
	}

	// Remove from cache
	if err := uc.bookCache.Set(ctx, req.ID, book.Entity{}); err != nil {
		logger.Warn("failed to remove book from cache", zap.Error(err))
		// Non-critical, continue
	}

	logger.Info("book deleted successfully")
	return nil
}
