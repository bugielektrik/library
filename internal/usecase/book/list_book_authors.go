package book

import (
	"context"
	"sync"

	"go.uber.org/zap"

	"library-service/internal/domain/author"
	"library-service/internal/domain/book"
	"library-service/internal/infrastructure/log"
	"library-service/internal/infrastructure/store"
	"library-service/pkg/errors"
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
	logger := log.FromContext(ctx).Named("list_book_authors_usecase").With(zap.String("book_id", req.BookID))

	if req.BookID == "" {
		return ListBookAuthorsResponse{}, errors.ErrInvalidInput.WithDetails("field", "book_id")
	}

	// Get the book
	bookEntity, err := uc.bookRepo.Get(ctx, req.BookID)
	if err != nil {
		if errors.Is(err, store.ErrorNotFound) {
			logger.Warn("book not found")
			return ListBookAuthorsResponse{}, errors.ErrBookNotFound.WithDetails("id", req.BookID)
		}
		logger.Error("failed to get book", zap.Error(err))
		return ListBookAuthorsResponse{}, errors.ErrDatabase.Wrap(err)
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
					log.FromContext(ctx).Warn("failed to cache author", zap.Error(cacheErr))
				}
			}

			mu.Lock()
			authors = append(authors, AuthorResponse{
				ID:        authorEntity.ID,
				FullName:  safeString(authorEntity.FullName),
				Pseudonym: safeString(authorEntity.Pseudonym),
				Specialty: safeString(authorEntity.Specialty),
			})
			mu.Unlock()
		}(authorID)
	}

	wg.Wait()

	if len(errs) > 0 {
		return nil, errors.ErrDatabase.Wrap(errs[0])
	}

	return authors, nil
}
