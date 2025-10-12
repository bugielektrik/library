# Testing Guide

> **Testing patterns, strategies, and best practices**

## Testing Philosophy

- **Domain Layer:** 100% coverage (pure logic, easy to test)
- **Use Case Layer:** 80%+ coverage (with mocked dependencies)
- **Adapter Layer:** Integration tests for repositories, unit tests for handlers
- **Overall:** 60%+ coverage minimum

## Test Organization

```
internal/
├── domain/
│   ├── book/
│   │   ├── service.go
│   │   └── service_test.go          # Unit tests (100% coverage)
│   │   └── service_benchmark_test.go # Benchmarks
│
├── usecase/
│   ├── book/
│   │   ├── create_book.go
│   │   └── create_book_test.go      # Unit tests with mocks
│
├── adapters/
│   ├── http/
│   │   └── handlers/
│   │       ├── book.go
│   │       └── book_test.go         # Handler tests
│   └── repository/
│       └── postgres/
│           ├── book.go
│           └── book_test.go         # Integration tests

test/
├── integration/                      # Full integration tests
│   └── api_test.go
└── fixtures/                         # Shared test data
    └── books.go
```

## Running Tests

```bash
# All tests
make test                # Fast (~2 seconds)

# Specific types
make test-unit           # Unit tests only (with -short flag)
make test-integration    # Integration tests (requires database)
make test-coverage       # Generate HTML coverage report

# Specific package
go test ./internal/domain/book/
go test -v ./internal/usecase/bookops/...

# Single test
go test -v -run TestCreateBook ./internal/usecase/bookops/

# Watch mode
reflex -r '\.go$' -s -- go test ./...
```

## Unit Tests (Domain Layer)

### Table-Driven Tests

```go
// internal/domain/book/service_test.go
package book_test

import (
    "testing"
    "library-service/internal/domain/book"
)

func TestService_ValidateISBN(t *testing.T) {
    svc := book.NewService()
    
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
            name:    "empty ISBN",
            isbn:    "",
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := svc.ValidateISBN(tt.isbn)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateISBN() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### Testing Entity Creation

```go
func TestNewEntity(t *testing.T) {
    book := book.NewEntity("Clean Code", "9780132350884", "Programming")
    
    if book.ID == "" {
        t.Error("Expected ID to be generated")
    }
    
    if book.Name != "Clean Code" {
        t.Errorf("Expected name 'Clean Code', got %s", book.Name)
    }
    
    if book.CreatedAt.IsZero() {
        t.Error("Expected CreatedAt to be set")
    }
}
```

## Use Case Tests (With Mocks)

### Using Testify Mocks

```go
// internal/usecase/bookops/create_book_test.go
package book_test

import (
    "context"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    
    "library-service/internal/domain/book"
    bookmocks "library-service/internal/infrastructure/pkg/repository/mocks"
    bookuc "library-service/internal/usecase/bookops"
)

func TestCreateBookUseCase_Execute(t *testing.T) {
    // Setup
    mockRepo := new(bookmocks.MockRepository)
    mockCache := new(bookmocks.MockCache)
    svc := book.NewService()
    
    uc := bookuc.NewCreateBookUseCase(mockRepo, mockCache, svc)
    
    ctx := context.Background()
    req := bookuc.CreateBookRequest{
        Name:  "Clean Code",
        ISBN:  "9780132350884",
        Genre: "Programming",
    }
    
    // Mock expectations
    mockRepo.On("Create", ctx, mock.AnythingOfType("book.Entity")).
        Return(nil)
    mockCache.On("Set", ctx, mock.Anything, mock.Anything).
        Return(nil)
    
    // Execute
    result, err := uc.Execute(ctx, req)
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, "Clean Code", result.Name)
    
    mockRepo.AssertExpectations(t)
    mockCache.AssertExpectations(t)
}

func TestCreateBookUseCase_InvalidISBN(t *testing.T) {
    mockRepo := new(bookmocks.MockRepository)
    mockCache := new(bookmocks.MockCache)
    svc := book.NewService()
    
    uc := bookuc.NewCreateBookUseCase(mockRepo, mockCache, svc)
    
    req := bookuc.CreateBookRequest{
        Name:  "Test Book",
        ISBN:  "invalid-isbn",
        Genre: "Test",
    }
    
    result, err := uc.Execute(context.Background(), req)
    
    assert.Error(t, err)
    assert.Nil(t, result)
    
    // Repository should not be called for invalid input
    mockRepo.AssertNotCalled(t, "Create")
}
```

### Creating Mocks

**Option 1: Manual Mocks**
```go
// internal/infrastructure/pkg/repository/mocks/repository.go
package mocks

import (
    "context"
    "github.com/stretchr/testify/mock"
    "library-service/internal/domain/book"
)

type MockRepository struct {
    mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, b book.Entity) error {
    args := m.Called(ctx, b)
    return args.Error(0)
}

func (m *MockRepository) GetByID(ctx context.Context, id string) (book.Entity, error) {
    args := m.Called(ctx, id)
    return args.Get(0).(book.Entity), args.Error(1)
}
// ... other methods
```

**Option 2: Using mockgen**
```bash
# Generate mocks
go generate ./...

# Or manually
mockgen -source=internal/domain/book/repository.go \
    -destination=internal/infrastructure/pkg/repository/mocks/repository.go \
    -package=mocks
```

## Integration Tests

### Repository Tests

```go
// internal/infrastructure/pkg/repository/postgres/book_test.go
//go:build integration

package postgres_test

import (
    "context"
    "testing"
    
    "library-service/internal/domain/book"
    "library-service/internal/infrastructure/pkg/repository/postgres"
    "library-service/test/testdb"
)

func TestBookRepository_Create(t *testing.T) {
    // Setup test database
    db := testdb.Setup(t)
    defer testdb.Teardown(t, db)
    
    repo := postgres.NewBookRepository(db)
    
    // Create book
    newBook := book.NewEntity("Integration Test Book", "9780132350884", "Test")
    err := repo.Create(context.Background(), newBook)
    
    if err != nil {
        t.Fatalf("Failed to create book: %v", err)
    }
    
    // Verify
    retrieved, err := repo.GetByID(context.Background(), newBook.ID)
    if err != nil {
        t.Fatalf("Failed to retrieve book: %v", err)
    }
    
    if retrieved.Name != newBook.Name {
        t.Errorf("Expected name %s, got %s", newBook.Name, retrieved.Name)
    }
}
```

### Test Database Setup

```go
// test/testdb/setup.go
package testdb

import (
    "database/sql"
    "testing"
    
    "github.com/jmoiron/sqlx"
    _ "github.com/lib/pq"
)

func Setup(t *testing.T) *sqlx.DB {
    t.Helper()
    
    dsn := "postgres://library:library123@localhost:5432/library_test?sslmode=disable"
    db, err := sqlx.Connect("postgres", dsn)
    if err != nil {
        t.Fatalf("Failed to connect to test database: %v", err)
    }
    
    // Run migrations or setup schema
    // ...
    
    return db
}

func Teardown(t *testing.T, db *sqlx.DB) {
    t.Helper()
    
    // Clean up test data
    db.Exec("TRUNCATE books CASCADE")
    db.Close()
}
```

## HTTP Handler Tests

```go
// internal/infrastructure/pkg/handler/book_test.go
package handlers_test

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    
    "library-service/internal/infrastructure/pkg/handlers"
    "library-service/internal/usecase/bookops"
    ucmocks "library-service/internal/usecase/mocks"
)

func TestBookHandler_Create(t *testing.T) {
    mockUC := new(ucmocks.MockCreateBookUseCase)
    handler := handlers.NewBookHandler(mockUC)
    
    // Prepare request
    reqBody := map[string]interface{}{
        "name":  "Test Book",
        "isbn":  "9780132350884",
        "genre": "Test",
    }
    body, _ := json.Marshal(reqBody)
    
    req := httptest.NewRequest(http.MethodPost, "/books", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    
    // Mock use case
    expectedBook := &book.Entity{ID: "123", Name: "Test Book"}
    mockUC.On("Execute", mock.Anything, mock.Anything).
        Return(expectedBook, nil)
    
    // Execute
    handler.Create(w, req)
    
    // Assert
    assert.Equal(t, http.StatusCreated, w.Code)
    
    var response map[string]interface{}
    json.Unmarshal(w.Body.Bytes(), &response)
    assert.Equal(t, "Test Book", response["name"])
}
```

## Test Fixtures

### Shared Test Data

```go
// test/fixtures/books.go
package fixtures

import "library-service/internal/domain/book"

func ValidBook() book.Entity {
    return book.NewEntity("Clean Code", "9780132350884", "Programming")
}

func BooksSlice() []book.Entity {
    return []book.Entity{
        ValidBook(),
        book.NewEntity("The Pragmatic Programmer", "9780135957059", "Programming"),
        book.NewEntity("Design Patterns", "9780201633610", "Programming"),
    }
}

func InvalidISBN() string {
    return "invalid-isbn-123"
}
```

### Using Fixtures

```go
import "library-service/test/fixtures"

func TestSomething(t *testing.T) {
    book := fixtures.ValidBook()
    // Use in test...
}
```

## Benchmarks

```go
// internal/domain/book/service_benchmark_test.go
package book_test

import (
    "testing"
    "library-service/internal/domain/book"
)

func BenchmarkValidateISBN(b *testing.B) {
    svc := book.NewService()
    isbn := "978-0-306-40615-7"
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = svc.ValidateISBN(isbn)
    }
}

func BenchmarkValidateISBNParallel(b *testing.B) {
    svc := book.NewService()
    isbn := "978-0-306-40615-7"
    
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            _ = svc.ValidateISBN(isbn)
        }
    })
}
```

```bash
# Run benchmarks
go test -bench=. -benchmem ./internal/domain/book/

# Output:
# BenchmarkValidateISBN-8         5000000    250 ns/op    0 B/op    0 allocs/op
```

## Coverage

```bash
# Generate coverage
go test -coverprofile=coverage.out ./...

# View in browser
go tool cover -html=coverage.out

# Get coverage percentage
go test -cover ./...

# Detailed coverage by function
go tool cover -func=coverage.out
```

## Test Best Practices

### DO

✅ **Use table-driven tests**
```go
tests := []struct {
    name string
    // test cases
}{...}
```

✅ **Test one thing per test**
```go
func TestCreateBook_ValidInput(t *testing.T) {}
func TestCreateBook_InvalidISBN(t *testing.T) {}
```

✅ **Use descriptive test names**
```go
func TestBookService_ValidateISBN_RejectsInvalidChecksum(t *testing.T) {}
```

✅ **Use t.Helper() for test helpers**
```go
func createTestBook(t *testing.T) book.Entity {
    t.Helper()
    // ...
}
```

✅ **Clean up after tests**
```go
defer testdb.Teardown(t, db)
```

### DON'T

❌ **Don't test implementation details**
```go
// Bad: Testing private function
func TestPrivateHelper(t *testing.T) {}
```

❌ **Don't use sleeps for synchronization**
```go
// Bad
time.Sleep(100 * time.Millisecond)

// Good: Use proper synchronization
done := make(chan bool)
```

❌ **Don't share state between tests**
```go
// Bad: Global variable modified in tests
var testDB *sql.DB

// Good: Create fresh state per test
func TestSomething(t *testing.T) {
    db := testdb.Setup(t)
    // ...
}
```

## Continuous Integration

Tests run automatically on:
- Every commit (via pre-commit hook)
- Pull requests (via GitHub Actions)
- Main branch merges

See `.github/workflows/ci.yml` for CI configuration.

## Next Steps

- Review [Development Guide](./development-guide.md) for workflow
- Check [Architecture Documentation](./architecture.md) for patterns and endpoint testing
- See [Standards](./standards.md) for code quality
