package service

import (
	"context"
	"library-service/internal/pkg/errors"
	"library-service/internal/pkg/logutil"
	"library-service/internal/pkg/strutil"

	"go.uber.org/zap"

	"library-service/internal/books/domain/book"
)

// ListBooksRequest represents the input for listing books
type ListBooksRequest struct {
	// Future: Add pagination, filtering, sorting
	Limit  int
	Offset int
}

// ListBooksResponse represents the output of listing books
type ListBooksResponse struct {
	Books []GetBookResponse
	Total int
}

// ListBooksUseCase handles retrieving all books
type ListBooksUseCase struct {
	bookRepo book.Repository
}

// NewListBooksUseCase creates a new instance of ListBooksUseCase
func NewListBooksUseCase(bookRepo book.Repository) *ListBooksUseCase {
	return &ListBooksUseCase{
		bookRepo: bookRepo,
	}
}

// Execute retrieves all books from the repository
func (uc *ListBooksUseCase) Execute(ctx context.Context, req ListBooksRequest) (ListBooksResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "book", "list")

	// Get all books from repository
	books, err := uc.bookRepo.List(ctx)
	if err != nil {
		logger.Error("failed to list books", zap.Error(err))
		return ListBooksResponse{}, errors.Database("database operation", err)
	}

	// Convert to response
	response := ListBooksResponse{
		Books: make([]GetBookResponse, len(books)),
		Total: len(books),
	}

	for i, b := range books {
		response.Books[i] = GetBookResponse{
			ID:      b.ID,
			Name:    strutil.SafeString(b.Name),
			Genre:   strutil.SafeString(b.Genre),
			ISBN:    strutil.SafeString(b.ISBN),
			Authors: b.Authors,
		}
	}

	logger.Info("books listed successfully", zap.Int("count", len(books)))
	return response, nil
}
