package service

import (
	"context"
	errors2 "library-service/internal/pkg/errors"
	"library-service/internal/pkg/logutil"

	"go.uber.org/zap"

	"library-service/internal/books/domain/book"
	"library-service/internal/infrastructure/store"
)

// UpdateBookRequest represents the input for updating a book
type UpdateBookRequest struct {
	ID      string
	Name    *string
	Genre   *string
	ISBN    *string
	Authors []string
}

// UpdateBookResponse represents the output of updating a book
type UpdateBookResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
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
func (uc *UpdateBookUseCase) Execute(ctx context.Context, req UpdateBookRequest) (UpdateBookResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "book", "update")

	if req.ID == "" {
		return UpdateBookResponse{}, errors2.ValidationRequired("id")
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
		if errors2.Is(err, store.ErrorNotFound) {
			logger.Warn("book not found")
			return UpdateBookResponse{}, errors2.NotFoundWithID("book", req.ID)
		}
		logger.Error("failed to update book", zap.Error(err))
		return UpdateBookResponse{}, errors2.Database("database operation", err)
	}

	// Update cache
	updatedBook.ID = req.ID
	if err := uc.bookCache.Set(ctx, req.ID, updatedBook); err != nil {
		logger.Warn("failed to update cache", zap.Error(err))
		// Non-critical, continue
	}

	logger.Info("book updated successfully", zap.String("id", req.ID))

	return UpdateBookResponse{
		Success: true,
		Message: "book updated successfully",
	}, nil
}
