# ADR 003: Constructor-Based Dependency Injection

**Date**: 2025-10-04
**Status**: Accepted
**Decision Makers**: Development Team

## Context

The application needs a way to manage dependencies between layers:
- Use cases depend on repositories and domain services
- HTTP handlers depend on use cases
- Repositories depend on database connections

We need a dependency injection approach that:
- Makes dependencies explicit
- Supports testing with mocks
- Works well with vibecoding
- Doesn't require complex DI frameworks

## Decision

We will use **Constructor-Based Dependency Injection** (also known as Constructor Injection) throughout the application.

### Pattern

```go
// Define dependencies in struct
type CreateBookUseCase struct {
    bookRepo    book.Repository     // Interface from domain
    bookService *book.Service       // Domain service
    bookCache   book.Cache          // Interface from domain
}

// Constructor receives all dependencies
func NewCreateBookUseCase(
    repo book.Repository,
    service *book.Service,
    cache book.Cache,
) *CreateBookUseCase {
    return &CreateBookUseCase{
        bookRepo:    repo,
        bookService: service,
        bookCache:   cache,
    }
}

// Methods use injected dependencies
func (uc *CreateBookUseCase) Execute(ctx context.Context, input Input) error {
    // Use uc.bookRepo, uc.bookService, uc.bookCache
}
```

### Wiring in Main

```go
// cmd/api/main.go
func main() {
    // 1. Infrastructure
    db := initDatabase()
    redis := initRedis()

    // 2. Repositories (adapters)
    bookRepo := repository.NewPostgresBookRepository(db)
    bookCache := cache.NewRedisBookCache(redis)

    // 3. Domain services
    bookService := book.NewService()

    // 4. Use cases
    createBookUC := usecase.NewCreateBookUseCase(bookRepo, bookService, bookCache)
    updateBookUC := usecase.NewUpdateBookUseCase(bookRepo, bookService, bookCache)

    // 5. HTTP handlers
    bookHandler := http.NewBookHandler(createBookUC, updateBookUC, ...)

    // 6. Routes
    setupRoutes(router, bookHandler)

    // 7. Start server
    router.Run(":8080")
}
```

## Decisions

### 1. No DI Framework

We will **NOT** use dependency injection frameworks like Wire or Dig.

**Rationale**:
- Constructor injection is simple and explicit
- No magic or generated code
- Easy for Claude Code to understand and modify
- Clear dependency graph visible in code

### 2. Interface-Based Dependencies

Use case dependencies are always **interfaces**, not concrete types:

```go
// ✅ Good: Interface dependency
type CreateBookUseCase struct {
    bookRepo book.Repository  // Interface
}

// ❌ Bad: Concrete dependency
type CreateBookUseCase struct {
    bookRepo *repository.PostgresBookRepository  // Concrete type
}
```

**Rationale**:
- Enables testing with mocks
- Supports Dependency Inversion Principle
- Allows swapping implementations

### 3. Explicit Constructor Parameters

Constructors explicitly list all dependencies (no variadic args, no options pattern):

```go
// ✅ Good: Explicit parameters
func NewCreateBookUseCase(
    repo book.Repository,
    service *book.Service,
    cache book.Cache,
) *CreateBookUseCase

// ❌ Bad: Variadic or options
func NewCreateBookUseCase(deps ...interface{}) *CreateBookUseCase
func NewCreateBookUseCase(opts *Options) *CreateBookUseCase
```

**Rationale**:
- Clear what's required
- Compile-time safety
- Easy to understand

### 4. Dependency Lifecycle

- **Domain Services**: Created once, shared across all use cases
- **Use Cases**: Created once at startup
- **HTTP Handlers**: Created once at startup
- **Repositories**: Created once per database connection

```go
// Singleton pattern for domain services
var (
    bookService   = book.NewService()      // Shared instance
    memberService = member.NewService()    // Shared instance
)

// Use cases use shared service instances
createBookUC := usecase.NewCreateBookUseCase(repo, bookService, cache)
updateBookUC := usecase.NewUpdateBookUseCase(repo, bookService, cache)
```

## Implementation Examples

### Use Case with DI

```go
package usecase

type CreateBookUseCase struct {
    bookRepo    book.Repository
    bookService *book.Service
    bookCache   book.Cache
}

func NewCreateBookUseCase(
    repo book.Repository,
    service *book.Service,
    cache book.Cache,
) *CreateBookUseCase {
    return &CreateBookUseCase{
        bookRepo:    repo,
        bookService: service,
        bookCache:   cache,
    }
}
```

### HTTP Handler with DI

```go
package http

type BookHandler struct {
    createBookUC *usecase.CreateBookUseCase
    updateBookUC *usecase.UpdateBookUseCase
    listBooksUC  *usecase.ListBooksUseCase
}

func NewBookHandler(
    createUC *usecase.CreateBookUseCase,
    updateUC *usecase.UpdateBookUseCase,
    listUC *usecase.ListBooksUseCase,
) *BookHandler {
    return &BookHandler{
        createBookUC: createUC,
        updateBookUC: updateUC,
        listBooksUC:  listUC,
    }
}
```

### Repository with DI

```go
package repository

type PostgresBookRepository struct {
    db *sqlx.DB
}

func NewPostgresBookRepository(db *sqlx.DB) book.Repository {
    return &PostgresBookRepository{db: db}
}
```

## Testing with DI

Constructor injection makes testing trivial:

```go
func TestCreateBookUseCase_Execute(t *testing.T) {
    // Arrange: Inject mocks
    mockRepo := new(MockBookRepository)
    mockCache := new(MockBookCache)
    bookService := book.NewService()

    uc := NewCreateBookUseCase(mockRepo, bookService, mockCache)

    // Mock expectations
    mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

    // Act
    err := uc.Execute(context.Background(), input)

    // Assert
    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
}
```

## Consequences

### Positive

✅ **Explicit Dependencies**: All dependencies visible in constructor
✅ **Testability**: Easy to inject mocks for testing
✅ **Compile Safety**: Missing dependencies caught at compile time
✅ **No Magic**: No generated code or reflection
✅ **Vibecoding**: Claude Code can easily trace dependencies
✅ **Refactoring**: IDE refactoring tools work perfectly
✅ **Documentation**: Constructor serves as documentation

### Negative

❌ **Boilerplate**: More constructor code
❌ **Main Complexity**: Wiring in main.go can get large
❌ **Manual Wiring**: No auto-discovery of dependencies

### Mitigations

- Keep main.go organized with clear sections
- Consider dependency groups for complex apps
- Use helper functions for common dependency sets

## Dependency Graph Example

```
main.go
  ↓
BookHandler
  ↓
CreateBookUseCase
  ↓
┌─────────────┬──────────────┬───────────────┐
│             │              │               │
BookRepository  BookService   BookCache
│               (no deps)     │
│                             │
PostgreSQL                   Redis
```

## Alternatives Considered

### 1. Service Locator Pattern
```go
// ❌ Not chosen
func (uc *CreateBookUseCase) Execute() {
    repo := locator.Get("BookRepository")
}
```
- ❌ Hidden dependencies
- ❌ Runtime errors instead of compile-time
- ❌ Hard to test

### 2. Global Variables
```go
// ❌ Not chosen
var GlobalBookRepo book.Repository

func (uc *CreateBookUseCase) Execute() {
    GlobalBookRepo.Create(...)
}
```
- ❌ Tight coupling
- ❌ Can't test with mocks
- ❌ Race conditions

### 3. DI Framework (Wire/Dig)
```go
// ❌ Not chosen
//+build wireinject

func InitializeApp() *App {
    wire.Build(...)
}
```
- ✅ Less boilerplate
- ❌ Generated code (harder for Claude)
- ❌ Build tags required
- ❌ Learning curve

### 4. Options Pattern
```go
// ❌ Not chosen
type Options struct {
    Repo    book.Repository
    Service *book.Service
}

func NewUseCase(opts Options) *UseCase
```
- ✅ Optional dependencies
- ❌ Less explicit
- ❌ Runtime nil panics

## Migration Example

### Before (Global Variables)
```go
var bookRepo book.Repository  // Global

type CreateBookUseCase struct{}

func (uc *CreateBookUseCase) Execute() {
    bookRepo.Create(...)  // Uses global
}
```

### After (Constructor Injection)
```go
type CreateBookUseCase struct {
    bookRepo book.Repository  // Injected
}

func NewCreateBookUseCase(repo book.Repository) *CreateBookUseCase {
    return &CreateBookUseCase{bookRepo: repo}
}

func (uc *CreateBookUseCase) Execute() {
    uc.bookRepo.Create(...)  // Uses injected
}
```

## Wiring Strategy

### Development
```go
// cmd/api/main.go
func main() {
    // Real implementations
    db := postgres.Connect()
    redis := redis.Connect()

    bookRepo := repository.NewPostgresBookRepository(db)
    // ... wire up
}
```

### Testing
```go
// cmd/api/main_test.go
func TestMain(m *testing.M) {
    // Mock implementations
    bookRepo := new(MockRepository)
    // ... wire up
}
```

## References

- [Dependency Injection in Go](https://blog.drewolson.org/dependency-injection-in-go)
- [Constructor Injection](https://martinfowler.com/articles/injection.html#ConstructorInjectionWithPicocontainer)
- [Wire: Compile-Time DI for Go](https://github.com/google/wire) (considered but not chosen)

## Review Notes

**Why Constructor Injection?**
- Simplicity over framework magic
- Vibecoding-friendly (Claude can trace easily)
- Testability without complexity

**When to Reconsider?**
- If app grows to 100+ use cases, consider Wire
- If wiring in main becomes unmanageable (>500 lines)

**Next Review**: After 50 use cases or 6 months
