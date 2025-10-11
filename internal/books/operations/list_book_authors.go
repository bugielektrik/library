package operations

import (
	"context"
	"sync"

	"go.uber.org/zap"

	"library-service/internal/books/domain/author"
	"library-service/internal/books/domain/book"
	"library-service/internal/infrastructure/store"
	"library-service/pkg/errors"
	"library-service/pkg/logutil"
	"library-service/pkg/strutil"
)

// ListBookAuthorsRequest represents the input for listing book authors
type ListBookAuthorsRequest struct {
	BookID string
}

// AuthorResponse represents an author in the response
type AuthorResponse struct {
	ID        string
	FullName  string
	Pseudonym string
	Specialty string
}

// ListBookAuthorsResponse represents the output of listing book authors
type ListBookAuthorsResponse struct {
	BookID  string
	Authors []AuthorResponse
}

// ListBookAuthorsUseCase handles retrieving all authors of a book
// This is a complex usecase that orchestrates multiple repository calls
type ListBookAuthorsUseCase struct {
	bookRepo    book.Repository
	authorRepo  author.Repository
	authorCache author.Cache
}

// NewListBookAuthorsUseCase creates a new instance of ListBookAuthorsUseCase
func NewListBookAuthorsUseCase(
	bookRepo book.Repository,
	authorRepo author.Repository,
	authorCache author.Cache,
) *ListBookAuthorsUseCase {
	return &ListBookAuthorsUseCase{
		bookRepo:    bookRepo,
		authorRepo:  authorRepo,
		authorCache: authorCache,
	}
}

// Execute retrieves all authors for a given book
func (uc *ListBookAuthorsUseCase) Execute(ctx context.Context, req ListBookAuthorsRequest) (ListBookAuthorsResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "book", "list_authors")

	if req.BookID == "" {
		return ListBookAuthorsResponse{}, errors.ValidationRequired("book_id")
	}

	// Get the book
	bookEntity, err := uc.bookRepo.Get(ctx, req.BookID)
	if err != nil {
		if errors.Is(err, store.ErrorNotFound) {
			logger.Warn("book not found")
			return ListBookAuthorsResponse{}, errors.NotFoundWithID("book", req.BookID)
		}
		logger.Error("failed to get book", zap.Error(err))
		return ListBookAuthorsResponse{}, errors.Database("database operation", err)
	}

	if len(bookEntity.Authors) == 0 {
		logger.Debug("book has no authors")
		return ListBookAuthorsResponse{
			BookID:  req.BookID,
			Authors: []AuthorResponse{},
		}, nil
	}

	// Fetch authors concurrently for better performance
	authors, err := uc.fetchAuthorsConcurrently(ctx, bookEntity.Authors)
	if err != nil {
		logger.Error("failed to fetch authors", zap.Error(err))
		return ListBookAuthorsResponse{}, err
	}

	logger.Info("book authors retrieved successfully", zap.Int("count", len(authors)))

	return ListBookAuthorsResponse{
		BookID:  req.BookID,
		Authors: authors,
	}, nil
}

// fetchAuthorsConcurrently fetches multiple authors concurrently
func (uc *ListBookAuthorsUseCase) fetchAuthorsConcurrently(ctx context.Context, authorIDs []string) ([]AuthorResponse, error) {
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		authors = make([]AuthorResponse, 0, len(authorIDs))
		errs    = make([]error, 0)
	)

	for _, authorID := range authorIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()

			// Try cache first
			authorEntity, err := uc.authorCache.Get(ctx, id)
			if err != nil {
				// Get from repository
				authorEntity, err = uc.authorRepo.Get(ctx, id)
				if err != nil {
					if !errors.Is(err, store.ErrorNotFound) {
						mu.Lock()
						errs = append(errs, err)
						mu.Unlock()
					}
					// Skip not found authors
					return
				}

				// Update cache
				if cacheErr := uc.authorCache.Set(ctx, id, authorEntity); cacheErr != nil {
					// Log but don't fail
					logutil.UseCaseLogger(ctx, "book", "list_authors").Warn("failed to cache author", zap.Error(cacheErr))
				}
			}

			mu.Lock()
			authors = append(authors, AuthorResponse{
				ID:        authorEntity.ID,
				FullName:  strutil.SafeString(authorEntity.FullName),
				Pseudonym: strutil.SafeString(authorEntity.Pseudonym),
				Specialty: strutil.SafeString(authorEntity.Specialty),
			})
			mu.Unlock()
		}(authorID)
	}

	wg.Wait()

	if len(errs) > 0 {
		return nil, errors.Database("database operation", errs[0])
	}

	return authors, nil
}
