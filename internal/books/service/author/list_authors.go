package author

import (
	"context"
	"library-service/internal/pkg/logutil"

	"go.uber.org/zap"

	authordomain "library-service/internal/books/domain/author"
)

// ListAuthorsRequest represents the input for listing authors.
type ListAuthorsRequest struct {
	// Future: Add pagination, filters, sorting
}

// ListAuthorsResponse represents the output of listing authors.
type ListAuthorsResponse struct {
	Authors []authordomain.Author
	Total   int
}

// ListAuthorsUseCase handles listing all authors.
type ListAuthorsUseCase struct {
	authorRepo authordomain.Repository
}

// NewListAuthorsUseCase creates a new instance of ListAuthorsUseCase.
func NewListAuthorsUseCase(authorRepo authordomain.Repository) *ListAuthorsUseCase {
	return &ListAuthorsUseCase{
		authorRepo: authorRepo,
	}
}

// Execute lists all authors.
func (uc *ListAuthorsUseCase) Execute(ctx context.Context, req ListAuthorsRequest) (ListAuthorsResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "author", "list")

	authors, err := uc.authorRepo.List(ctx)
	if err != nil {
		logger.Error("failed to list authors", zap.Error(err))
		return ListAuthorsResponse{}, err
	}

	logger.Info("authors listed successfully", zap.Int("count", len(authors)))

	return ListAuthorsResponse{
		Authors: authors,
		Total:   len(authors),
	}, nil
}
