# Use Case Layer

**Application business logic and orchestration - the application's brain.**

## Purpose

This layer contains:
- **Use Cases**: Application-specific business flows (one per file)
- **Use Case DTOs**: Input/output data structures
- **Orchestration**: Coordinates domain services and repositories
- **Transaction Management**: Ensures data consistency
- **Cross-Cutting Concerns**: Caching, validation, error handling

## Dependency Rule

Use cases depend **only on the domain layer** (inward dependency).

```
Use Case (this layer)
  ↓ depends on
Domain (business logic)

  ✗ NO dependency on
Adapters (HTTP, DB implementations)
```

## Directory Structure

```
usecase/
├── book/
│   ├── create_book.go        # Create book use case
│   ├── create_book_test.go   # Unit tests
│   ├── update_book.go        # Update book use case
│   ├── delete_book.go        # Delete book use case
│   ├── list_books.go         # List books use case
│   ├── get_book.go           # Get single book use case
│   └── dto.go                # Use case DTOs
│
├── member/
│   ├── create_member.go
│   ├── update_member.go
│   └── dto.go
│
├── subscription/
│   ├── subscribe_member.go   # Subscription flow
│   └── dto.go
│
└── author/
    ├── create_author.go
    └── dto.go
```

## Use Case Pattern

### Structure

```go
// 1. Define use case struct with dependencies
type CreateBookUseCase struct {
    bookRepo    book.Repository    // Interface from domain
    bookService *book.Service      // Domain service
    bookCache   book.Cache         // Interface from domain
}

// 2. Constructor with dependency injection
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

// 3. Execute method (standard name)
func (uc *CreateBookUseCase) Execute(ctx context.Context, input CreateBookInput) (*book.Entity, error) {
    // Step 1: Map DTO to entity
    entity := book.Entity{
        ID:      uuid.New().String(),
        Name:    input.Name,
        ISBN:    input.ISBN,
        Authors: input.Authors,
        Genre:   input.Genre,
    }

    // Step 2: Validate with domain service
    if err := uc.bookService.ValidateBook(entity); err != nil {
        return nil, fmt.Errorf("validation failed: %w", err)
    }

    // Step 3: Check business constraints
    existing, _ := uc.bookRepo.GetByISBN(ctx, entity.ISBN)
    if existing != nil {
        return nil, errors.New("book with this ISBN already exists")
    }

    // Step 4: Persist
    if err := uc.bookRepo.Create(ctx, entity); err != nil {
        return nil, fmt.Errorf("failed to create book: %w", err)
    }

    // Step 5: Cache (optional, fire-and-forget)
    _ = uc.bookCache.Set(ctx, entity)

    return &entity, nil
}
```

### DTOs

```go
// Input DTO
type CreateBookInput struct {
    Name    string   `json:"name"`
    ISBN    string   `json:"isbn"`
    Authors []string `json:"authors"`
    Genre   string   `json:"genre"`
}

// Helper method
func (dto CreateBookInput) ToEntity() book.Entity {
    return book.Entity{
        ID:      uuid.New().String(),
        Name:    dto.Name,
        ISBN:    dto.ISBN,
        Authors: dto.Authors,
        Genre:   dto.Genre,
    }
}
```

## Typical Use Case Flow

```
1. Accept Input DTO
   ↓
2. Map DTO → Domain Entity
   ↓
3. Validate with Domain Service
   ↓
4. Check Business Rules (uniqueness, etc.)
   ↓
5. Execute Domain Logic
   ↓
6. Persist via Repository
   ↓
7. Update Cache (if needed)
   ↓
8. Return Result
```

## Testing

Use cases should have **80%+ test coverage** with mocked dependencies.

```go
func TestCreateBookUseCase_Execute(t *testing.T) {
    // Arrange
    mockRepo := &MockRepository{}
    mockService := book.NewService()
    mockCache := &MockCache{}

    uc := NewCreateBookUseCase(mockRepo, mockService, mockCache)

    input := CreateBookInput{
        Name: "Clean Code",
        ISBN: "978-0-13-235088-4",
    }

    // Setup mocks
    mockRepo.On("GetByISBN", mock.Anything, input.ISBN).Return(nil, nil)
    mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

    // Act
    result, err := uc.Execute(context.Background(), input)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, "Clean Code", result.Name)
    mockRepo.AssertExpectations(t)
}
```

## Examples

### Create Book Use Case
**File**: `book/create_book.go`
**Purpose**: Create a new book with validation
**Steps**: Validate → Check duplicates → Persist → Cache

### Subscribe Member Use Case
**File**: `subscription/subscribe_member.go`
**Purpose**: Create member subscription
**Steps**: Validate member → Check active subscription → Calculate dates → Create subscription

### Delete Book Use Case
**File**: `book/delete_book.go`
**Purpose**: Delete book safely
**Steps**: Check if borrowed → Validate deletion → Remove from DB → Invalidate cache

## Best Practices

1. **One Use Case per File**: Single Responsibility Principle
2. **Thin Orchestration**: Delegate logic to domain services
3. **Error Wrapping**: Add context to errors (`fmt.Errorf("context: %w", err)`)
4. **Dependency Injection**: Constructor injection only
5. **DTO Separation**: Don't expose domain entities to outer layers
6. **Testability**: Mock all dependencies
7. **Idempotency**: Design use cases to be safely retryable
8. **Transaction Boundaries**: Use cases define transaction scope

## Common Patterns

### Caching Pattern
```go
// Check cache first
if cached, err := uc.cache.Get(ctx, id); err == nil {
    return cached, nil
}

// Fetch from DB
entity, err := uc.repo.GetByID(ctx, id)
if err != nil {
    return nil, err
}

// Update cache (fire-and-forget)
_ = uc.cache.Set(ctx, entity)

return entity, nil
```

### Validation Pattern
```go
// Domain validation
if err := uc.service.ValidateEntity(entity); err != nil {
    return nil, fmt.Errorf("validation failed: %w", err)
}

// Business constraints
if exists, _ := uc.repo.Exists(ctx, entity.ID); exists {
    return nil, errors.New("entity already exists")
}
```

### Error Handling Pattern
```go
if err := uc.repo.Create(ctx, entity); err != nil {
    // Wrap with context
    return nil, fmt.Errorf("failed to create %s: %w", entity.Name, err)
}
```

## Adding New Use Case

1. **Create file**: `internal/usecase/{domain}/{action}_{entity}.go`
2. **Define struct**: With repository and service dependencies
3. **Add constructor**: `New{Action}{Entity}UseCase(...)`
4. **Implement Execute**: `func (uc *UC) Execute(ctx, input) (output, error)`
5. **Write tests**: Mock dependencies, test flows
6. **Update DTOs**: Add input/output DTOs to `dto.go`

## References

- [Domain Layer](../domain/README.md)
- [Development Guide](../../docs/guides/DEVELOPMENT.md)
