# Repository Pattern

## Overview

Repositories provide data access abstraction. Interfaces are defined in the domain layer, implementations in adapters.

## Interface Definition (Domain Layer)

```go
package book

import "context"

// Repository interface defined in domain layer
type Repository interface {
    Create(ctx context.Context, book Book) (Book, error)
    GetByID(ctx context.Context, id string) (Book, error)
    Update(ctx context.Context, book Book) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context) ([]Book, error)
}
```

## PostgreSQL Implementation

### Using BaseRepository Pattern

```go
package repository

import (
    "github.com/jmoiron/sqlx"
    "library-service/internal/infrastructure/pkg/repository/postgres"
    bookdomain "library-service/internal/books/domain/book"
)

// BookRepository implements book.Repository for PostgreSQL
type BookRepository struct {
    postgres.BaseRepository[bookdomain.Book]
}

func NewBookRepository(db *sqlx.DB) *BookRepository {
    return &BookRepository{
        BaseRepository: postgres.NewBaseRepository[bookdomain.Book](db, "books"),
    }
}
```

The `BaseRepository` provides:
- `Get(ctx, id) (T, error)` - Get by ID
- `Delete(ctx, id) error` - Delete by ID
- `GetDB() *sqlx.DB` - Access to database

### Custom Methods

Add domain-specific methods:

```go
// GetByISBN retrieves a book by ISBN
func (r *BookRepository) GetByISBN(ctx context.Context, isbn string) (bookdomain.Book, error) {
    query := `
        SELECT id, name, genre, isbn, authors, created_at, updated_at
        FROM books
        WHERE isbn = $1
    `

    var book bookdomain.Book
    err := r.GetDB().GetContext(ctx, &book, query, isbn)
    if err != nil {
        return bookdomain.Book{}, postgres.HandleSQLError(err)
    }

    return book, nil
}
```

### Create Operation

```go
func (r *BookRepository) Create(ctx context.Context, book bookdomain.Book) (bookdomain.Book, error) {
    query := `
        INSERT INTO books (id, name, genre, isbn, authors, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, created_at, updated_at
    `

    var result struct {
        ID        string    `db:"id"`
        CreatedAt time.Time `db:"created_at"`
        UpdatedAt time.Time `db:"updated_at"`
    }

    err := r.GetDB().QueryRowContext(
        ctx,
        query,
        book.ID,
        book.Name,
        book.Genre,
        book.ISBN,
        pq.Array(book.Authors),
        time.Now(),
        time.Now(),
    ).Scan(&result.ID, &result.CreatedAt, &result.UpdatedAt)

    if err != nil {
        return bookdomain.Book{}, postgres.HandleSQLError(err)
    }

    book.ID = result.ID
    book.CreatedAt = result.CreatedAt
    book.UpdatedAt = result.UpdatedAt

    return book, nil
}
```

### Update Operation

```go
func (r *BookRepository) Update(ctx context.Context, book bookdomain.Book) error {
    query := `
        UPDATE books
        SET name = $2, genre = $3, isbn = $4, authors = $5, updated_at = $6
        WHERE id = $1
        RETURNING updated_at
    `

    var updatedAt time.Time
    err := r.GetDB().QueryRowContext(
        ctx,
        query,
        book.ID,
        book.Name,
        book.Genre,
        book.ISBN,
        pq.Array(book.Authors),
        time.Now(),
    ).Scan(&updatedAt)

    if err != nil {
        return postgres.HandleSQLError(err)
    }

    return nil
}
```

### List with Filtering

```go
func (r *BookRepository) List(ctx context.Context) ([]bookdomain.Book, error) {
    query := `
        SELECT id, name, genre, isbn, authors, created_at, updated_at
        FROM books
        ORDER BY created_at DESC
    `

    var books []bookdomain.Book
    err := r.GetDB().SelectContext(ctx, &books, query)
    if err != nil {
        return nil, fmt.Errorf("listing books: %w", err)
    }

    return books, nil
}
```

## Error Handling

Use `postgres.HandleSQLError` for consistent error mapping:

```go
err := r.GetDB().GetContext(ctx, &book, query, id)
if err != nil {
    return bookdomain.Book{}, postgres.HandleSQLError(err)
}

// Maps:
// sql.ErrNoRows         → store.ErrorNotFound (404)
// pq.UniqueViolation    → store.ErrorAlreadyExists (409)
// Other errors          → wrapped error
```

## File Organization

Repositories are in bounded context repository layer:

```
internal/books/repository/
├── book.go           # BookRepository
├── author.go         # AuthorRepository
└── doc.go

internal/members/repository/
├── member.go         # MemberRepository
└── doc.go

internal/payments/repository/
├── payment.go        # PaymentRepository
├── saved_card.go     # SavedCardRepository
├── receipt.go        # ReceiptRepository
└── callback_retry.go # CallbackRetryRepository
```

## Generic Helpers

Use generic helpers from `internal/infrastructure/pkg/repository/postgres/`:

```go
// Generic get with custom query
book, err := postgres.GetOne[bookdomain.Book](ctx, r.GetDB(), query, args...)

// Generic list with custom query
books, err := postgres.GetMany[bookdomain.Book](ctx, r.GetDB(), query, args...)

// Existence check
exists, err := postgres.Exists(ctx, r.GetDB(), "books", "isbn", isbn)

// Count
count, err := postgres.Count(ctx, r.GetDB(), "books", "WHERE genre = ?", genre)
```

## Testing

Test repositories with testcontainers or real database:

```go
//go:build integration

func TestBookRepository_Create(t *testing.T) {
    // Setup test database
    db := test.SetupTestDB(t)
    defer db.Close()

    repo := repository.NewBookRepository(db)

    // Test
    book := bookdomain.Book{
        ID:      "123",
        Name:    strPtr("Test Book"),
        ISBN:    strPtr("978-0-306-40615-7"),
        Authors: []string{"Author 1"},
    }

    created, err := repo.Create(context.Background(), book)

    // Assert
    require.NoError(t, err)
    assert.Equal(t, "123", created.ID)
    assert.NotZero(t, created.CreatedAt)
}
```

## Multiple Storage Implementations

You can have multiple implementations of the same interface:

```
internal/infrastructure/pkg/repository/
├── postgres/      # PostgreSQL implementations
├── mongo/         # MongoDB implementations
├── memory/        # In-memory implementations (testing)
└── mocks/         # Auto-generated mocks (testify/mock)
```

## Cache Interface

Similar pattern for caches:

```go
// Interface in domain
type Cache interface {
    Get(ctx context.Context, key string) (Book, error)
    Set(ctx context.Context, key string, book Book) error
    Delete(ctx context.Context, key string) error
}

// Implementations in adapters
internal/infrastructure/pkg/cache/
├── redis/book.go    # Redis implementation
└── memory/book.go   # In-memory implementation
```

## Complete Examples

See actual repositories in:
- `internal/books/repository/book.go` - Book repository with BaseRepository
- `internal/members/repository/member.go` - Member repository
- `internal/payments/repository/payment.go` - Payment repository
- `internal/reservations/repository/reservation.go` - Reservation repository
