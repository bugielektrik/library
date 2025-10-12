# Use Case Pattern

## Overview

Use cases orchestrate domain services and repositories to implement application business logic.

## Standard Structure

```go
package operations

import (
    "context"
    "library-service/internal/books/domain/book"
    "library-service/internal/infrastructure/pkg/logutil"
)

// CreateBookUseCase handles book creation with business rules
type CreateBookUseCase struct {
    bookRepo    book.Repository
    bookCache   book.Cache
    bookService *book.Service
}

func NewCreateBookUseCase(
    bookRepo book.Repository,
    bookCache book.Cache,
    bookService *book.Service,
) *CreateBookUseCase {
    return &CreateBookUseCase{
        bookRepo:    bookRepo,
        bookCache:   bookCache,
        bookService: bookService,
    }
}
```

## Execute Method Pattern

**All use cases follow the Execute(ctx, Request) (Response, error) pattern:**

```go
// Request struct
type CreateBookRequest struct {
    Name    *string
    Genre   *string
    ISBN    *string
    Authors []string
}

// Response struct
type CreateBookResponse struct {
    Book book.Book
}

// Execute implements the use case logic
func (uc *CreateBookUseCase) Execute(
    ctx context.Context,
    req CreateBookRequest,
) (CreateBookResponse, error) {
    // 1. Create logger with use case context
    logger := logutil.UseCaseLogger(ctx, "book", "create")

    // 2. Validate using domain service
    bookEntity := book.Book{
        Name:    req.Name,
        Genre:   req.Genre,
        ISBN:    req.ISBN,
        Authors: req.Authors,
    }

    if err := uc.bookService.Validate(bookEntity); err != nil {
        return CreateBookResponse{}, err
    }

    // 3. Execute business logic (repository operations)
    createdBook, err := uc.bookRepo.Create(ctx, bookEntity)
    if err != nil {
        return CreateBookResponse{}, fmt.Errorf("creating book: %w", err)
    }

    // 4. Update cache (if applicable)
    if err := uc.bookCache.Set(ctx, createdBook.ID, createdBook); err != nil {
        logger.Warn("failed to cache book", zap.Error(err))
        // Don't fail - cache is not critical
    }

    // 5. Return domain entity (NOT DTO)
    return CreateBookResponse{Book: createdBook}, nil
}
```

## Key Principles

### 1. Request/Response Pattern
Every use case has Request and Response structs:

```go
// Request contains input parameters
type GetBookRequest struct {
    ID string
}

// Response contains domain entities
type GetBookResponse struct {
    Book book.Book
}

// Execute signature is consistent
func (uc *GetBookUseCase) Execute(
    ctx context.Context,
    req GetBookRequest,
) (GetBookResponse, error)
```

### 2. Domain Service Usage
Use domain services for business rules:

```go
// ✅ CORRECT - Use domain service
if err := uc.bookService.ValidateISBN(req.ISBN); err != nil {
    return CreateBookResponse{}, err
}

// ❌ WRONG - Business logic in use case
if len(req.ISBN) != 13 {
    return CreateBookResponse{}, errors.New("invalid ISBN")
}
```

### 3. Error Wrapping
Wrap errors with context:

```go
// ✅ CORRECT - Context preserved
if err := uc.bookRepo.Create(ctx, book); err != nil {
    return CreateBookResponse{}, fmt.Errorf("creating book: %w", err)
}

// ❌ WRONG - Context lost
if err := uc.bookRepo.Create(ctx, book); err != nil {
    return CreateBookResponse{}, err
}
```

### 4. Return Domain Entities
Use cases return domain entities, NOT DTOs:

```go
// ✅ CORRECT - Domain entity
func (uc *GetBookUseCase) Execute(ctx context.Context, req GetBookRequest) (GetBookResponse, error) {
    book, err := uc.bookRepo.GetByID(ctx, req.ID)
    return GetBookResponse{Book: book}, err  // Domain entity
}

// ❌ WRONG - DTO in use case
func (uc *GetBookUseCase) Execute(ctx context.Context, req GetBookRequest) (dto.BookResponse, error) {
    // DTOs belong in HTTP layer only!
}
```

## Dependency Injection

Use cases receive dependencies via constructor:

```go
// Domain service created in container.go
bookService := book.NewService()

// Use case receives all dependencies
createBookUC := operations.NewCreateBookUseCase(
    repos.Book,       // Repository interface
    caches.Book,      // Cache interface
    bookService,      // Domain service
)
```

## Logging Pattern

Use structured logging with context:

```go
logger := logutil.UseCaseLogger(ctx, "domain", "operation")

// Log at appropriate levels
logger.Debug("starting operation", zap.String("id", id))
logger.Info("operation completed", zap.Duration("elapsed", elapsed))
logger.Warn("cache miss", zap.String("key", key))
logger.Error("operation failed", zap.Error(err))
```

## Transaction Handling

For operations requiring transactions:

```go
func (uc *CreateBookWithAuthorsUseCase) Execute(
    ctx context.Context,
    req CreateBookWithAuthorsRequest,
) (CreateBookWithAuthorsResponse, error) {
    // Begin transaction
    tx, err := uc.db.BeginTx(ctx)
    if err != nil {
        return CreateBookWithAuthorsResponse{}, err
    }
    defer tx.Rollback() // Safe to call even after commit

    // Execute operations in transaction
    book, err := uc.bookRepo.CreateTx(ctx, tx, bookEntity)
    if err != nil {
        return CreateBookWithAuthorsResponse{}, err
    }

    for _, author := range req.Authors {
        if err := uc.authorRepo.CreateTx(ctx, tx, author); err != nil {
            return CreateBookWithAuthorsResponse{}, err
        }
    }

    // Commit transaction
    if err := tx.Commit(); err != nil {
        return CreateBookWithAuthorsResponse{}, err
    }

    return CreateBookWithAuthorsResponse{Book: book}, nil
}
```

## File Organization

Use cases organized by domain in bounded contexts:

```
internal/books/operations/
├── create_book.go           # CreateBookUseCase
├── get_book.go              # GetBookUseCase
├── update_book.go           # UpdateBookUseCase
├── delete_book.go           # DeleteBookUseCase
├── list_books.go            # ListBooksUseCase
├── author/                  # Author subdomain
│   └── list_authors.go
└── doc.go
```

## Testing Pattern

Use cases are tested with mocked repositories:

```go
func TestCreateBookUseCase_Execute(t *testing.T) {
    // Setup mocks
    mockRepo := new(MockBookRepository)
    mockCache := new(MockBookCache)
    bookService := book.NewService()

    // Create use case
    uc := operations.NewCreateBookUseCase(mockRepo, mockCache, bookService)

    // Setup expectations
    mockRepo.On("Create", mock.Anything, mock.Anything).
        Return(book.Book{ID: "123"}, nil)
    mockCache.On("Set", mock.Anything, mock.Anything, mock.Anything).
        Return(nil)

    // Execute
    result, err := uc.Execute(context.Background(), operations.CreateBookRequest{
        Name: strPtr("Test Book"),
    })

    // Assert
    require.NoError(t, err)
    assert.Equal(t, "123", result.Book.ID)
    mockRepo.AssertExpectations(t)
}
```

## Complete Examples

See actual use cases in:
- `internal/books/operations/create_book.go` - Book creation
- `internal/members/operations/auth/register.go` - Member registration
- `internal/payments/operations/payment/initiate_payment.go` - Payment initiation
- `internal/reservations/operations/create_reservation.go` - Reservation creation
