# Testing Pattern

## Overview

Tests in this codebase follow Go best practices with table-driven tests, mocking, and clear assertions.

## Table-Driven Test Pattern

**Standard pattern used throughout the codebase:**

```go
func TestBookService_ValidateISBN(t *testing.T) {
    tests := []struct {
        name    string
        isbn    string
        wantErr bool
    }{
        {
            name:    "valid ISBN-13",
            isbn:    "978-0-306-40615-7",
            wantErr: false,
        },
        {
            name:    "valid ISBN-13 without hyphens",
            isbn:    "9780306406157",
            wantErr: false,
        },
        {
            name:    "valid ISBN-10",
            isbn:    "0-306-40615-2",
            wantErr: false,
        },
        {
            name:    "invalid checksum",
            isbn:    "978-0-306-40615-8",
            wantErr: true,
        },
        {
            name:    "too short",
            isbn:    "12345",
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            service := book.NewService()
            err := service.ValidateISBN(tt.isbn)

            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateISBN() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## Mocking Pattern

### Using testify/mock

```go
package operations_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"

    "library-service/internal/books/domain/book"
)

// MockBookRepository implements book.Repository for testing
type MockBookRepository struct {
    mock.Mock
}

func (m *MockBookRepository) Create(ctx context.Context, b book.Book) (book.Book, error) {
    args := m.Called(ctx, b)
    return args.Get(0).(book.Book), args.Error(1)
}

func (m *MockBookRepository) GetByID(ctx context.Context, id string) (book.Book, error) {
    args := m.Called(ctx, id)
    if args.Get(0) == nil {
        return book.Book{}, args.Error(1)
    }
    return args.Get(0).(book.Book), args.Error(1)
}

// Test using mock
func TestCreateBookUseCase_Execute(t *testing.T) {
    tests := []struct {
        name        string
        request     CreateBookRequest
        setupMocks  func(*MockBookRepository, *MockBookCache)
        expectError bool
        validate    func(*testing.T, CreateBookResponse)
    }{
        {
            name: "successful creation",
            request: CreateBookRequest{
                Name:    strPtr("Test Book"),
                ISBN:    strPtr("978-0-306-40615-7"),
                Authors: []string{"Author 1"},
            },
            setupMocks: func(repo *MockBookRepository, cache *MockBookCache) {
                // Setup expectations
                repo.On("Create", mock.Anything, mock.MatchedBy(func(b book.Book) bool {
                    return *b.Name == "Test Book"
                })).Return(book.Book{
                    ID:   "book-123",
                    Name: strPtr("Test Book"),
                }, nil).Once()

                cache.On("Set", mock.Anything, "book-123", mock.Anything).
                    Return(nil).Once()
            },
            expectError: false,
            validate: func(t *testing.T, resp CreateBookResponse) {
                assert.Equal(t, "book-123", resp.Book.ID)
                assert.Equal(t, "Test Book", *resp.Book.Name)
            },
        },
        {
            name: "validation error",
            request: CreateBookRequest{
                Name: strPtr("Test Book"),
                ISBN: strPtr("invalid-isbn"),
            },
            setupMocks: func(repo *MockBookRepository, cache *MockBookCache) {
                // No expectations - should fail before repository call
            },
            expectError: true,
            validate:    nil,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup mocks
            mockRepo := new(MockBookRepository)
            mockCache := new(MockBookCache)
            bookService := book.NewService()

            if tt.setupMocks != nil {
                tt.setupMocks(mockRepo, mockCache)
            }

            // Create use case
            uc := NewCreateBookUseCase(mockRepo, mockCache, bookService)

            // Execute
            result, err := uc.Execute(context.Background(), tt.request)

            // Assert
            if tt.expectError {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
                if tt.validate != nil {
                    tt.validate(t, result)
                }
            }

            // Verify mock expectations
            mockRepo.AssertExpectations(t)
            mockCache.AssertExpectations(t)
        })
    }
}
```

### Auto-Generated Mocks

Use mockery for auto-generated mocks:

```bash
# Generate mocks
mockery --name Repository --dir internal/books/domain/book \
    --output internal/adapters/repository/mocks \
    --outpkg mocks \
    --filename book_repository_mock.go \
    --structname MockBookRepository
```

## Integration Tests

Use build tags for integration tests:

```go
//go:build integration

package integration

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"

    "library-service/internal/books/repository"
    "library-service/test/testdb"
)

func TestBookRepository_Create(t *testing.T) {
    // Setup test database
    db := testdb.SetupTestDB(t)
    defer db.Close()

    repo := repository.NewBookRepository(db)

    tests := []struct {
        name    string
        book    book.Book
        wantErr bool
    }{
        {
            name: "valid book",
            book: book.Book{
                ID:      uuid.New().String(),
                Name:    strPtr("Test Book"),
                ISBN:    strPtr("978-0-306-40615-7"),
                Authors: []string{"Author 1"},
            },
            wantErr: false,
        },
        {
            name: "duplicate ISBN",
            book: book.Book{
                ID:      uuid.New().String(),
                Name:    strPtr("Another Book"),
                ISBN:    strPtr("978-0-306-40615-7"), // Same as above
                Authors: []string{"Author 2"},
            },
            wantErr: true, // Should fail on unique constraint
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            created, err := repo.Create(context.Background(), tt.book)

            if tt.wantErr {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
                assert.Equal(t, tt.book.ID, created.ID)
                assert.NotZero(t, created.CreatedAt)
            }
        })
    }
}

// Run integration tests
// go test -tags=integration ./test/integration/...
```

## Test Helpers

Use helpers for common operations:

```go
// test/helpers/fixtures.go
package helpers

func strPtr(s string) *string {
    return &s
}

func intPtr(i int) *int {
    return &i
}

func MustParseTime(s string) time.Time {
    t, _ := time.Parse(time.RFC3339, s)
    return t
}
```

## Test Fixtures

Use builders for complex test data:

```go
// test/builders/book.go
package builders

import "library-service/internal/books/domain/book"

type BookBuilder struct {
    book book.Book
}

func NewBook() *BookBuilder {
    return &BookBuilder{
        book: book.Book{
            ID:      uuid.New().String(),
            Name:    strPtr("Default Book"),
            Genre:   strPtr("Fiction"),
            ISBN:    strPtr("978-0-306-40615-7"),
            Authors: []string{"Default Author"},
        },
    }
}

func (b *BookBuilder) WithID(id string) *BookBuilder {
    b.book.ID = id
    return b
}

func (b *BookBuilder) WithName(name string) *BookBuilder {
    b.book.Name = &name
    return b
}

func (b *BookBuilder) Build() book.Book {
    return b.book
}

// Usage in tests
book := builders.NewBook().
    WithID("custom-id").
    WithName("Custom Book").
    Build()
```

## Benchmarking

Add benchmarks for performance-critical code:

```go
func BenchmarkBookService_ValidateISBN(b *testing.B) {
    service := book.NewService()
    isbn := "978-0-306-40615-7"

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = service.ValidateISBN(isbn)
    }
}

// Run benchmarks
// go test -bench=. ./internal/books/domain/book/
```

## Test Organization

```
test/
├── builders/          # Test data builders
│   ├── book.go
│   ├── member.go
│   └── payment.go
├── fixtures/          # Static test data
│   ├── books.go
│   ├── members.go
│   └── payments.go
├── helpers/           # Test helpers
│   └── fixtures.go
├── integration/       # Integration tests
│   ├── book_repository_test.go
│   └── payment_repository_test.go
└── testdb/           # Test database setup
    └── setup.go
```

## Coverage Goals

- **Domain layer**: 100% coverage (critical business logic)
- **Use cases**: 80%+ coverage
- **Handlers**: 70%+ coverage
- **Overall**: 60%+ coverage

```bash
# Check coverage
make test-coverage

# Or manually
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

## Testing Commands

```bash
# Unit tests (fast)
make test-unit
go test ./internal/... -short

# Integration tests (requires database)
make test-integration
go test -tags=integration ./test/integration/...

# All tests with coverage
make test
go test ./... -race -coverprofile=coverage.out

# Specific package
go test -v ./internal/books/domain/book/
go test -v -run TestBookService_ValidateISBN ./internal/books/domain/book/

# With race detection
go test -race ./...

# Benchmarks
go test -bench=. -benchmem ./internal/books/domain/book/
```

## Complete Examples

See actual tests in:
- `internal/books/domain/book/service_test.go` - Domain service tests
- `internal/members/operations/auth/register_test.go` - Use case tests with mocks
- `test/integration/book_repository_test.go` - Integration tests
- `internal/infrastructure/auth/jwt_test.go` - Infrastructure tests
