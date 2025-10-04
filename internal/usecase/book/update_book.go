package book

import (
	"context"

	"go.uber.org/zap"

	"library-service/internal/domain/book"
	"library-service/internal/infrastructure/log"
	"library-service/internal/infrastructure/store"
	"library-service/pkg/errors"
)

// UpdateBookRequest represents the input for updating a book
type UpdateBookRequest struct {
	ID      string
	Name    *string
	Genre   *string
	ISBN    *string
	Authors []string
}

// UpdateBookUseCase handles updating an existing book
type UpdateBookUseCase struct {
	bookRepo  book.Repository
	bookCache book.Cache
}

// NewUpdateBookUseCase creates a new instance of UpdateBookUseCase
func NewUpdateBookUseCase(bookRepo book.Repository, bookCache book.Cache) *UpdateBookUseCase {
	return &UpdateBookUseCase{
		bookRepo:  bookRepo,
		bookCache: bookCache,
	}
}

// Execute updates an existing book
func (uc *UpdateBookUseCase) Execute(ctx context.Context, req UpdateBookRequest) error {
	logger := log.FromContext(ctx).Named("update_book_usecase").With(zap.String("id", req.ID))

	if req.ID == "" {
		return errors.ErrInvalidInput.WithDetails("field", "id")
	}

	// Build the update request
	updateReq := book.Request{}
	if req.Name != nil {
		updateReq.Name = *req.Name
	}
	if req.Genre != nil {
		updateReq.Genre = *req.Genre
	}
	if req.ISBN != nil {
		updateReq.ISBN = *req.ISBN
	}
	if req.Authors != nil {
		updateReq.Authors = req.Authors
	}

	// Create updated book entity
	updatedBook := book.New(updateReq)

	// Update in repository
	err := uc.bookRepo.Update(ctx, req.ID, updatedBook)
	if err != nil {
		if errors.Is(err, store.ErrorNotFound) {
			logger.Warn("book not found")
			return errors.ErrBookNotFound.WithDetails("id", req.ID)
		}
		logger.Error("failed to update book", zap.Error(err))
		return errors.ErrDatabase.Wrap(err)
	}

	// Update cache
	updatedBook.ID = req.ID
	if err := uc.bookCache.Set(ctx, req.ID, updatedBook); err != nil {
		logger.Warn("failed to update cache", zap.Error(err))
		// Non-critical, continue
	}

	logger.Info("book updated successfully")
	return nil
}
