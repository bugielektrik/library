package operations

import (
	"context"

	"go.uber.org/zap"

	"library-service/internal/books/domain/book"
	"library-service/internal/infrastructure/store"
	"library-service/pkg/errors"
	"library-service/pkg/logutil"
)

// DeleteBookRequest represents the input for deleting a book
type DeleteBookRequest struct {
	ID string
}

// DeleteBookResponse represents the output of deleting a book
type DeleteBookResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
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
func (uc *DeleteBookUseCase) Execute(ctx context.Context, req DeleteBookRequest) (DeleteBookResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "book", "delete")

	if req.ID == "" {
		return DeleteBookResponse{}, errors.ValidationRequired("id")
	}

	// Delete from repository
	err := uc.bookRepo.Delete(ctx, req.ID)
	if err != nil {
		if errors.Is(err, store.ErrorNotFound) {
			logger.Warn("book not found")
			return DeleteBookResponse{}, errors.NotFoundWithID("book", req.ID)
		}
		logger.Error("failed to delete book", zap.Error(err))
		return DeleteBookResponse{}, errors.Database("database operation", err)
	}

	// Remove from cache
	if err := uc.bookCache.Set(ctx, req.ID, book.Book{}); err != nil {
		logger.Warn("failed to remove book from cache", zap.Error(err))
		// Non-critical, continue
	}

	logger.Info("book deleted successfully", zap.String("id", req.ID))

	return DeleteBookResponse{
		Success: true,
		Message: "book deleted successfully",
	}, nil
}
