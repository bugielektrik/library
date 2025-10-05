# Code Standards & Conventions

> **Go best practices, style guide, and project conventions**

## Go Style Guide

This project follows the [Google Go Style Guide](https://google.github.io/styleguide/go/) and [Effective Go](https://go.dev/doc/effective_go).

## Naming Conventions

### Packages

```go
// ✅ Good: lowercase, single word, no underscores
package book
package repository
package http

// ❌ Bad: camelCase, snake_case, multiple words
package bookManager
package book_manager
package bookmanager
```

### Files

```go
// ✅ Good: snake_case
book_service.go
create_book.go
book_repository.go

// ❌ Bad: camelCase, kebab-case
bookService.go
createBook.go
book-repository.go
```

### Variables & Functions

```go
// ✅ Good: Exported (CamelCase), Unexported (camelCase)
type BookService struct {}
func NewBookService() *BookService {}
func (s *BookService) ValidateISBN(isbn string) error {}

var DefaultTimeout = 30 * time.Second
const MaxRetries = 3

// ❌ Bad: Incorrect casing
type bookService struct {}  // Should be unexported or BookService
func New_Book_Service() {}  // Should be NewBookService
```

### Interfaces

```go
// ✅ Good: Verb + "er" suffix
type Reader interface { Read() }
type Writer interface { Write() }
type BookCreator interface { CreateBook() }

// ✅ Good: Single method interfaces
type Repository interface {
    Create(ctx context.Context, book Entity) error
}

// ❌ Bad: "Interface" suffix, noun without "er"
type BookInterface interface {}
type BookRepository interface {}  // Should be just Repository in domain/book package
```

### Constants

```go
// ✅ Good: CamelCase or SCREAMING_SNAKE_CASE for groups
const (
    StatusActive   = "active"
    StatusInactive = "inactive"
)

const (
    MAX_CONNECTIONS = 100
    DEFAULT_TIMEOUT = 30
)

// ❌ Bad: Inconsistent naming
const status_active = "active"
const MaxConnections = 100  // Mixed styles in same group
```

## Code Organization

### File Structure

Each file should have a clear, single purpose:

```go
// book/entity.go - Entity definition
package book

type Entity struct {
    ID   string
    Name string
    ISBN string
}

func NewEntity(name, isbn string) Entity {
    return Entity{
        ID:   generateID(),
        Name: name,
        ISBN: isbn,
    }
}
```

```go
// book/service.go - Business logic
package book

type Service struct {}

func NewService() *Service {
    return &Service{}
}

func (s *Service) ValidateISBN(isbn string) error {
    // Business rule implementation
}
```

### Import Grouping

```go
import (
    // Standard library
    "context"
    "fmt"
    "time"
    
    // External dependencies
    "github.com/google/uuid"
    "go.uber.org/zap"
    
    // Internal packages
    "library-service/internal/domain/book"
    "library-service/pkg/errors"
)
```

### Function Order

```go
// 1. Type definitions
type Service struct {
    repo Repository
}

// 2. Constructor
func NewService(repo Repository) *Service {
    return &Service{repo: repo}
}

// 3. Public methods (alphabetically)
func (s *Service) CreateBook() {}
func (s *Service) DeleteBook() {}
func (s *Service) GetBook() {}

// 4. Private methods
func (s *Service) validateInternal() {}
```

## Error Handling

### Error Wrapping

```go
// ✅ Good: Add context with %w for error wrapping
if err := s.repo.Create(ctx, book); err != nil {
    return fmt.Errorf("creating book in repository: %w", err)
}

// ✅ Good: Check errors immediately
result, err := doSomething()
if err != nil {
    return err
}
useResult(result)

// ❌ Bad: No context, can't unwrap
if err := s.repo.Create(ctx, book); err != nil {
    return err
}

// ❌ Bad: Deferred error check
result, err := doSomething()
useResult(result)
if err != nil {
    return err
}
```

### Custom Errors

```go
// ✅ Good: Define domain errors
var (
    ErrBookNotFound = errors.New("book not found")
    ErrInvalidISBN  = errors.New("invalid ISBN format")
)

// Usage
if !isValid {
    return ErrInvalidISBN
}

// Check
if errors.Is(err, ErrBookNotFound) {
    // Handle not found
}
```

### Error Messages

```go
// ✅ Good: Lowercase, no punctuation
errors.New("invalid ISBN format")
fmt.Errorf("failed to connect to database")

// ❌ Bad: Capitalized, punctuation
errors.New("Invalid ISBN Format.")
fmt.Errorf("Failed to connect to database!")
```

## Function Design

### Return Values

```go
// ✅ Good: Error as last return value
func GetBook(id string) (*Book, error) {}
func CreateBook(book Book) (string, error) {}

// ✅ Good: Context as first parameter
func GetBook(ctx context.Context, id string) (*Book, error) {}

// ❌ Bad: Error not last
func GetBook(id string) (error, *Book) {}

// ❌ Bad: Context not first
func GetBook(id string, ctx context.Context) (*Book, error) {}
```

### Named Returns

```go
// ✅ Good: Named returns for documentation
func GetBookStats(id string) (totalLoans int, averageRating float64, err error) {
    // Implementation
    return
}

// ✅ Good: Simple return
func GetBook(id string) (*Book, error) {
    return book, nil
}

// ❌ Bad: Naked return in long function
func ComplexOperation() (result string, err error) {
    // 100+ lines of code
    return  // Hard to track what's being returned
}
```

### Function Length

```go
// ✅ Good: Functions < 50 lines
func CreateBook(book Book) error {
    if err := validate(book); err != nil {
        return err
    }
    return save(book)
}

// ❌ Bad: 200+ line functions
// Split into smaller, focused functions
```

## Struct Design

### Field Ordering

```go
// ✅ Good: Exported fields first, grouped logically
type Book struct {
    // Core fields
    ID   string
    Name string
    ISBN string
    
    // Metadata
    CreatedAt time.Time
    UpdatedAt time.Time
    
    // Unexported fields
    cached bool
}
```

### Constructor Pattern

```go
// ✅ Good: Explicit constructor
func NewBook(name, isbn string) Book {
    return Book{
        ID:        uuid.New().String(),
        Name:      name,
        ISBN:      isbn,
        CreatedAt: time.Now(),
    }
}

// ❌ Bad: Relying on zero values for complex initialization
book := Book{}
book.ID = uuid.New().String()
```

## Comments

### Package Comments

```go
// ✅ Good: Package comment before package declaration
// Package book provides domain entities and business logic for book management.
//
// The book domain handles ISBN validation, book lifecycle, and business rules
// such as preventing deletion of books with active loans.
package book
```

### Function Comments

```go
// ✅ Good: Starts with function name, explains what and why
// ValidateISBN validates both ISBN-10 and ISBN-13 formats using checksum algorithm.
// It returns an error if the ISBN is invalid or has incorrect checksum.
func ValidateISBN(isbn string) error {}

// ❌ Bad: Doesn't explain purpose
// This function validates ISBN
func ValidateISBN(isbn string) error {}
```

### TODO Comments

```go
// ✅ Good: TODO with username and context
// TODO(john): Implement caching for frequently accessed books
// TODO(sarah): Optimize query to avoid N+1 problem

// ❌ Bad: TODO without owner
// TODO: fix this
```

## Testing Conventions

### Test File Naming

```go
// ✅ Good: _test.go suffix, same package
// book/service.go
// book/service_test.go

// ✅ Good: _benchmark_test.go for benchmarks
// book/service_benchmark_test.go
```

### Test Function Naming

```go
// ✅ Good: Descriptive test names
func TestService_ValidateISBN_ValidISBN13(t *testing.T) {}
func TestService_ValidateISBN_InvalidChecksum(t *testing.T) {}
func TestCreateBook_DuplicateISBN(t *testing.T) {}

// ❌ Bad: Generic names
func TestValidate(t *testing.T) {}
func TestBook(t *testing.T) {}
```

### Table-Driven Tests

```go
// ✅ Good: Use table-driven tests
func TestValidateISBN(t *testing.T) {
    tests := []struct {
        name    string
        isbn    string
        wantErr bool
    }{
        {"valid ISBN-13", "978-0-306-40615-7", false},
        {"invalid checksum", "978-0-306-40615-8", true},
        {"empty string", "", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateISBN(tt.isbn)
            if (err != nil) != tt.wantErr {
                t.Errorf("got error %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## Concurrency

### Goroutines

```go
// ✅ Good: Clear lifecycle, use WaitGroup
var wg sync.WaitGroup
for _, book := range books {
    wg.Add(1)
    go func(b Book) {
        defer wg.Done()
        process(b)
    }(book)
}
wg.Wait()

// ❌ Bad: Unbounded goroutines without synchronization
for _, book := range books {
    go process(book)  // No wait, may reference wrong book
}
```

### Context Usage

```go
// ✅ Good: Pass context for cancellation
func ProcessBooks(ctx context.Context, books []Book) error {
    for _, book := range books {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            if err := process(book); err != nil {
                return err
            }
        }
    }
    return nil
}
```

## Performance Guidelines

### Preallocate Slices

```go
// ✅ Good: Preallocate when size is known
books := make([]Book, 0, expectedSize)

// ❌ Bad: Multiple reallocations
var books []Book  // Will reallocate multiple times
```

### String Building

```go
// ✅ Good: Use strings.Builder for multiple concatenations
var builder strings.Builder
for _, part := range parts {
    builder.WriteString(part)
}
result := builder.String()

// ❌ Bad: Multiple string concatenations (allocations)
var result string
for _, part := range parts {
    result += part  // New allocation each iteration
}
```

### Defer in Loops

```go
// ✅ Good: Manual cleanup in loop
for _, file := range files {
    f, err := os.Open(file)
    if err != nil {
        continue
    }
    process(f)
    f.Close()  // Explicit close
}

// ❌ Bad: Defer in loop (defers accumulate)
for _, file := range files {
    f, err := os.Open(file)
    if err != nil {
        continue
    }
    defer f.Close()  // Won't close until function returns
    process(f)
}
```

## Linting

Project uses `golangci-lint` with 25+ linters enabled.

**Run linters:**
```bash
make lint
golangci-lint run
```

**Auto-fix:**
```bash
golangci-lint run --fix
```

**Key linters enabled:**
- `gofmt` - Code formatting
- `goimports` - Import organization
- `govet` - Suspicious constructs
- `errcheck` - Unchecked errors
- `staticcheck` - Advanced static analysis
- `gosec` - Security issues
- `gocyclo` - Cyclomatic complexity (max: 10)
- `dupl` - Code duplication

## Pre-commit Checklist

Before committing, ensure:

```bash
make ci  # Runs: fmt → vet → lint → test → build
```

- [ ] Code formatted (`make fmt`)
- [ ] No linter errors (`make lint`)
- [ ] All tests pass (`make test`)
- [ ] No race conditions (`go test -race ./...`)
- [ ] Coverage meets requirements
- [ ] Comments added for exported functions
- [ ] TODO comments have owner

## Common Patterns

### Singleton (Avoid)

```go
// ❌ Bad: Global state, hard to test
var GlobalDB *sql.DB

func init() {
    GlobalDB = connect()
}

// ✅ Good: Dependency injection
type Service struct {
    db *sql.DB
}

func NewService(db *sql.DB) *Service {
    return &Service{db: db}
}
```

### Options Pattern

```go
// ✅ Good: For optional parameters
type Config struct {
    Timeout time.Duration
    Retries int
}

type Option func(*Config)

func WithTimeout(d time.Duration) Option {
    return func(c *Config) {
        c.Timeout = d
    }
}

func NewService(opts ...Option) *Service {
    cfg := &Config{
        Timeout: 30 * time.Second,  // defaults
        Retries: 3,
    }
    for _, opt := range opts {
        opt(cfg)
    }
    return &Service{config: cfg}
}

// Usage
svc := NewService(
    WithTimeout(10*time.Second),
    WithRetries(5),
)
```

## References

- [Effective Go](https://go.dev/doc/effective_go)
- [Google Go Style Guide](https://google.github.io/styleguide/go/)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

## Project-Specific Standards

### Layer-Specific Rules

**Domain Layer:**
- ZERO external dependencies
- 100% test coverage
- Pure functions preferred
- No database, HTTP, or framework code

**Use Case Layer:**
- One use case = one file
- Depends only on interfaces
- Returns domain entities, not DTOs

**Adapter Layer:**
- Thin layer, delegate to use cases
- Contains DTOs for external format
- Implements domain interfaces

**Infrastructure Layer:**
- Cross-cutting concerns only
- Can depend on any layer
- Configuration, logging, auth

### File Naming

```
{entity}.go              # Entity definition
{entity}_service.go      # Business logic
{entity}_repository.go   # Repository interface
{operation}.go           # Use case (create_book.go)
```

### Import Aliases

```go
import (
    bookdomain "library-service/internal/domain/book"
    bookusecase "library-service/internal/usecase/bookops"
    bookhandler "library-service/internal/adapters/http/handlers"
)
```

## Next Steps

- Review [Architecture Guide](./architecture.md) for patterns
- Check [Development Guide](./development.md) for workflow
- See [Testing Guide](./testing.md) for test conventions
