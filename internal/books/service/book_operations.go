package service

import (
	"context"
	"sync"

	"go.uber.org/zap"

	"library-service/internal/books/domain/author"
	"library-service/internal/books/domain/book"
	"library-service/internal/infrastructure/store"
	errors2 "library-service/internal/pkg/errors"
	"library-service/internal/pkg/logutil"
	"library-service/internal/pkg/strutil"
)

// ================================================================================
// Common Types & Interfaces
// ================================================================================

// Validator defines the validation interface
type Validator interface {
	Validate(i interface{}) error
}

// ================================================================================
// Create Book Use Case
// ================================================================================

// CreateBookRequest represents the input for creating a book
type CreateBookRequest struct {
	Name    string   `validate:"required,min=1,max=200"`
	Genre   string   `validate:"required,min=1,max=100"`
	ISBN    string   `validate:"required,isbn10|isbn13"`
	Authors []string `validate:"required,min=1,dive,required"`
}

// CreateBookResponse represents the output of creating a book
type CreateBookResponse struct {
	ID      string
	Name    string
	Genre   string
	ISBN    string
	Authors []string
}

// CreateBookUseCase handles the creation of a new book.
//
// Architecture Pattern: Standard CRUD use case with repository, cache, and domain service.
//
// See Also:
//   - Similar pattern: internal/usecase/authops/register.go (registration flow)
//   - Domain layer: internal/books/domain/book/service.go (ISBN validation)
//   - HTTP handler: internal/books/handler/crud.go
//   - Repository: internal/books/repository/book.go
//   - ADR: .claude/adr/002-clean-architecture-boundaries.md (layer rules)
//   - Test: internal/books/service/create_book_test.go
type CreateBookUseCase struct {
	bookRepo    book.Repository
	bookCache   book.Cache
	bookService *book.Service
	validator   Validator
}

// NewCreateBookUseCase creates a new instance of CreateBookUseCase
func NewCreateBookUseCase(bookRepo book.Repository, bookCache book.Cache, bookService *book.Service, validator Validator) *CreateBookUseCase {
	return &CreateBookUseCase{
		bookRepo:    bookRepo,
		bookCache:   bookCache,
		bookService: bookService,
		validator:   validator,
	}
}

// Execute creates a new book in the system
func (uc *CreateBookUseCase) Execute(ctx context.Context, req CreateBookRequest) (CreateBookResponse, error) {
	// Validate request using go-playground/validator
	if err := uc.validator.Validate(req); err != nil {
		return CreateBookResponse{}, errors2.ErrValidation.Wrap(err)
	}

	// Create logger with use case context
	logger := logutil.UseCaseLogger(ctx, "book", "create")
	logger.Debug("creating book",
		zap.String("isbn", req.ISBN),
		zap.String("name", req.Name),
		zap.Int("authors_count", len(req.Authors)),
	)

	// Create book entity from request
	bookEntity := book.New(book.Request{
		Name:    req.Name,
		Genre:   req.Genre,
		ISBN:    req.ISBN,
		Authors: req.Authors,
	})

	// Validate book using domain service
	if err := uc.bookService.Validate(bookEntity); err != nil {
		logger.Warn("validation failed", zap.Error(err))
		return CreateBookResponse{}, err
	}

	// Check if book with ISBN already exists
	exists, err := uc.bookRepo.Get(ctx, req.ISBN)
	if err == nil && exists.ID != "" {
		logger.Warn("book with ISBN already exists", zap.String("existing_id", exists.ID))
		return CreateBookResponse{}, errors2.AlreadyExists("book", "ISBN", req.ISBN)
	}

	// Save to repository
	id, err := uc.bookRepo.Add(ctx, bookEntity)
	if err != nil {
		logger.Error("failed to add book to repository", zap.Error(err))
		return CreateBookResponse{}, errors2.Database("create book", err)
	}
	bookEntity.ID = id

	// Cache the new book
	if err := uc.bookCache.Set(ctx, id, bookEntity); err != nil {
		logger.Warn("failed to cache book", zap.Error(err))
		// Non-critical error, continue
	}

	logger.Info("book created successfully", zap.String("id", id))

	return CreateBookResponse{
		ID:      bookEntity.ID,
		Name:    strutil.SafeString(bookEntity.Name),
		Genre:   strutil.SafeString(bookEntity.Genre),
		ISBN:    strutil.SafeString(bookEntity.ISBN),
		Authors: bookEntity.Authors,
	}, nil
}

// ================================================================================
// Update Book Use Case
// ================================================================================

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

// ================================================================================
// Delete Book Use Case
// ================================================================================

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
		return DeleteBookResponse{}, errors2.ValidationRequired("id")
	}

	// Delete from repository
	err := uc.bookRepo.Delete(ctx, req.ID)
	if err != nil {
		if errors2.Is(err, store.ErrorNotFound) {
			logger.Warn("book not found")
			return DeleteBookResponse{}, errors2.NotFoundWithID("book", req.ID)
		}
		logger.Error("failed to delete book", zap.Error(err))
		return DeleteBookResponse{}, errors2.Database("database operation", err)
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

// ================================================================================
// Get Book Use Case
// ================================================================================

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
		return GetBookResponse{
			ID:      bookEntity.ID,
			Name:    strutil.SafeString(bookEntity.Name),
			Genre:   strutil.SafeString(bookEntity.Genre),
			ISBN:    strutil.SafeString(bookEntity.ISBN),
			Authors: bookEntity.Authors,
		}, nil
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
	return GetBookResponse{
		ID:      bookEntity.ID,
		Name:    strutil.SafeString(bookEntity.Name),
		Genre:   strutil.SafeString(bookEntity.Genre),
		ISBN:    strutil.SafeString(bookEntity.ISBN),
		Authors: bookEntity.Authors,
	}, nil
}

// ================================================================================
// List Books Use Case
// ================================================================================

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
		return ListBooksResponse{}, errors2.Database("database operation", err)
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

// ================================================================================
// List Book Authors Use Case
// ================================================================================

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
		return ListBookAuthorsResponse{}, errors2.ValidationRequired("book_id")
	}

	// Get the book
	bookEntity, err := uc.bookRepo.Get(ctx, req.BookID)
	if err != nil {
		if errors2.Is(err, store.ErrorNotFound) {
			logger.Warn("book not found")
			return ListBookAuthorsResponse{}, errors2.NotFoundWithID("book", req.BookID)
		}
		logger.Error("failed to get book", zap.Error(err))
		return ListBookAuthorsResponse{}, errors2.Database("database operation", err)
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
					if !errors2.Is(err, store.ErrorNotFound) {
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
		return nil, errors2.Database("database operation", errs[0])
	}

	return authors, nil
}
