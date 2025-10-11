# Current Code Patterns

**Purpose:** Quick reference for active code patterns. Claude Code should check this file first before searching codebase for pattern examples.

**Last Updated:** October 11, 2025

---

## File Organization

### Directory Structure
```
internal/
├── domain/              # Business logic (zero dependencies)
│   ├── book/
│   ├── member/
│   ├── payment/
│   └── reservation/
├── usecase/             # Application logic
│   ├── authops/        # Note: "ops" suffix
│   ├── bookops/
│   ├── paymentops/
│   └── reservationops/
├── adapters/            # External interfaces
│   ├── http/
│   │   ├── handlers/
│   │   │   ├── book/   # Organized by domain
│   │   │   ├── auth/
│   │   │   └── payment/
│   │   ├── dto/
│   │   └── middleware/
│   └── repository/
│       ├── postgres/
│       ├── mongo/
│       └── mocks/
└── infrastructure/      # Technical concerns
    ├── auth/
    ├── config/
    └── log/
```

### Naming Conventions

**Use case packages:** Add "ops" suffix
- ✅ `internal/usecase/bookops`
- ✅ `internal/usecase/authops`
- ❌ `internal/usecase/book` (conflicts with domain)

**Handler files:** Organized by operation type
- `handler.go` - struct and Routes()
- `crud.go` - create, update, delete
- `query.go` - get, list
- `manage.go` - specialized operations

**Test files:** Same directory as code
- ✅ `book_test.go` next to `book.go`
- ❌ Separate `tests/` directory

---

## Code Templates

### 1. Handler Template
```go
type BookHandler struct {
    handlers.BaseHandler
    useCases  *usecase.Container  // Grouped container
    validator *middleware.Validator
}

func NewBookHandler(
    useCases *usecase.Container,
    validator *middleware.Validator,
) *BookHandler {
    return &BookHandler{
        useCases:  useCases,
        validator: validator,
    }
}

func (h *BookHandler) Routes() chi.Router {
    r := chi.NewRouter()
    r.Post("/", h.create)
    r.Get("/{id}", h.get)
    return r
}

func (h *BookHandler) create(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    logger := logutil.HandlerLogger(ctx, "book", "create")

    memberID, ok := h.GetMemberID(w, r)
    if !ok {
        return
    }

    var req dto.CreateBookRequest
    if err := httputil.DecodeJSON(r, &req); err != nil {
        h.RespondError(w, r, err)
        return
    }

    if !h.validator.ValidateStruct(w, req) {
        return
    }

    result, err := h.useCases.Book.CreateBook.Execute(ctx, bookops.CreateBookRequest{
        Name: req.Name,
    })

    if err != nil {
        h.RespondError(w, r, err)
        return
    }

    logger.Info("book created", zap.String("id", result.ID))
    h.RespondJSON(w, http.StatusCreated, dto.FromBookEntity(result.Book))
}
```

### 2. Use Case Template
```go
type CreateBookUseCase struct {
    bookRepo book.Repository
}

func NewCreateBookUseCase(bookRepo book.Repository) *CreateBookUseCase {
    return &CreateBookUseCase{bookRepo: bookRepo}
}

type CreateBookRequest struct {
    Name     string
    MemberID string
}

type CreateBookResponse struct {
    Book book.Book
}

func (uc *CreateBookUseCase) Execute(
    ctx context.Context,
    req CreateBookRequest,
) (CreateBookResponse, error) {
    logger := logutil.UseCaseLogger(ctx, "book", "create")

    // Validation
    if req.Name == "" {
        return CreateBookResponse{}, errors.NewError(errors.CodeValidation).
            WithField("name", "required").
            Build()
    }

    // Create entity
    bookEntity := book.Book{
        Name:     req.Name,
        MemberID: req.MemberID,
    }

    // Repository call
    created, err := uc.bookRepo.Create(ctx, bookEntity)
    if err != nil {
        logger.Error("failed to create book", zap.Error(err))
        return CreateBookResponse{}, fmt.Errorf("creating book: %w", err)
    }

    logger.Info("book created", zap.String("id", created.ID))

    return CreateBookResponse{Book: created}, nil
}
```

### 3. Repository Template
```go
type BookRepository struct {
    db *sqlx.DB
}

func NewBookRepository(db *sqlx.DB) *BookRepository {
    return &BookRepository{db: db}
}

func (r *BookRepository) Create(ctx context.Context, bookEntity book.Book) (book.Book, error) {
    logger := logutil.RepositoryLogger(ctx, "book", "create")

    query := `
        INSERT INTO books (id, name, created_at)
        VALUES (:id, :name, NOW())
        RETURNING id, name, created_at
    `

    params := map[string]interface{}{
        "id":   bookEntity.ID,
        "name": bookEntity.Name,
    }

    rows, err := r.db.NamedQueryContext(ctx, query, params)
    if err != nil {
        logger.Error("failed to insert book", zap.Error(err))
        return book.Book{}, r.handleError(err)
    }
    defer rows.Close()

    var result book.Book
    if rows.Next() {
        if err := rows.StructScan(&result); err != nil {
            return book.Book{}, fmt.Errorf("scanning book: %w", err)
        }
    }

    logger.Info("book created", zap.String("id", result.ID))
    return result, nil
}
```

### 4. Table-Driven Test Template
```go
func TestCreateBookUseCase_Execute(t *testing.T) {
    tests := []struct {
        name        string
        request     CreateBookRequest
        setupMocks  func(*mocks.MockBookRepository)
        expectError bool
        validate    func(*testing.T, CreateBookResponse)
    }{
        {
            name: "successful creation",
            request: CreateBookRequest{
                Name: "Test Book",
            },
            setupMocks: func(repo *mocks.MockBookRepository) {
                repo.On("Create", mock.Anything, mock.Anything).
                    Return(book.Book{ID: "book-123"}, nil).
                    Once()
            },
            expectError: false,
            validate: func(t *testing.T, resp CreateBookResponse) {
                assert.Equal(t, "book-123", resp.Book.ID)
            },
        },
        {
            name: "validation error",
            request: CreateBookRequest{
                Name: "",
            },
            setupMocks: func(repo *mocks.MockBookRepository) {},
            expectError: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo := new(mocks.MockBookRepository)
            tt.setupMocks(mockRepo)

            uc := NewCreateBookUseCase(mockRepo)
            ctx := helpers.TestContext(t)

            result, err := uc.Execute(ctx, tt.request)

            if tt.expectError {
                require.Error(t, err)
            } else {
                require.NoError(t, err)
                if tt.validate != nil {
                    tt.validate(t, result)
                }
            }

            mockRepo.AssertExpectations(t)
        })
    }
}
```

---

## Common Operations

### Adding a New Feature

**1. Domain Layer** (`internal/domain/{entity}/`)
```bash
# Create entity, repository interface, service
- entity.go
- repository.go
- service.go (optional)
- service_test.go
```

**2. Use Case Layer** (`internal/usecase/{entity}ops/`)
```bash
# Create use cases
- create_{entity}.go
- get_{entity}.go
- update_{entity}.go
- delete_{entity}.go
- list_{entities}.go
```

**3. Adapter Layer** (`internal/adapters/`)
```bash
# HTTP handlers
- http/handlers/{entity}/handler.go
- http/handlers/{entity}/crud.go
- http/handlers/{entity}/query.go

# DTOs
- http/dto/{entity}.go

# Repository implementation
- repository/postgres/{entity}.go
```

**4. Wire in Container** (`internal/usecase/container.go`)
```go
// Add to Container struct
type Container struct {
    // ...
    Entity EntityUseCases
}

// Add to NewContainer function
entityUseCases := newEntityUseCases(repos.Entity)
```

**5. Add Router** (`internal/adapters/http/router.go`)
```go
entityHandler := entity.NewEntityHandler(cfg.Usecases, validator)

r.Group(func(r chi.Router) {
    r.Use(authMiddleware.Authenticate)
    r.Mount("/entities", entityHandler.Routes())
})
```

---

## Validation Patterns

### DTO Validation
```go
type CreateRequest struct {
    Name  string `json:"name" validate:"required,min=3,max=100"`
    Email string `json:"email" validate:"required,email"`
    Age   int    `json:"age" validate:"required,min=18,max=120"`
}

// In handler
if !h.validator.ValidateStruct(w, req) {
    return  // Error already written
}
```

### Business Logic Validation
```go
// In domain service
func (s *BookService) ValidateBook(b Book) error {
    if len(b.ISBN) != 13 {
        return errors.NewError(errors.CodeValidation).
            WithField("isbn", "must be 13 digits").
            Build()
    }
    return nil
}
```

---

## Error Patterns

### Creating Errors
```go
// Validation error
errors.NewError(errors.CodeValidation).
    WithField("name", "too long").
    WithDetail("max_length", 100).
    Build()

// Not found
errors.NotFound("book not found")

// Unauthorized
errors.Unauthorized("invalid credentials")

// Internal error
errors.Internal("database connection failed")
```

### Wrapping Errors
```go
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}
```

### Checking Errors
```go
if errors.Is(err, store.ErrorNotFound) {
    // Handle not found
}
```

---

## Logging Patterns

### Handler Logging
```go
logger := logutil.HandlerLogger(ctx, "book_handler", "create")
logger.Info("book created", zap.String("id", result.ID))
logger.Error("failed to create", zap.Error(err))
```

### Use Case Logging
```go
logger := logutil.UseCaseLogger(ctx, "book", "create")
logger.Info("creating book", zap.String("name", req.Name))
```

### Repository Logging
```go
logger := logutil.RepositoryLogger(ctx, "book", "create")
logger.Info("book saved", zap.String("id", id))
```

---

## Database Patterns

### Transactions
```go
tx, err := r.db.BeginTxx(ctx, nil)
if err != nil {
    return fmt.Errorf("beginning transaction: %w", err)
}
defer tx.Rollback()

// Operations...

if err := tx.Commit(); err != nil {
    return fmt.Errorf("committing transaction: %w", err)
}
```

### Named Parameters
```go
query := `INSERT INTO books (id, name) VALUES (:id, :name)`
params := map[string]interface{}{
    "id":   book.ID,
    "name": book.Name,
}
_, err := r.db.NamedExecContext(ctx, query, params)
```

---

## File Size Limits

**Enforce these limits:**
- ✅ Handlers: 100-200 lines
- ✅ Use cases: 200-400 lines
- ✅ Repositories: 150-300 lines
- ✅ Domain entities: 100-200 lines
- ✅ Tests: 200-400 lines

**When file exceeds limit:**
- Split into multiple files (crud.go, query.go, manage.go)
- Extract helper methods
- Create separate use cases

---

## Quick Reference

### Most Common Operations
1. **Add endpoint:** Handler → DTO → Use case → Repository
2. **Add validation:** DTO validate tags or domain service
3. **Add test:** Create table-driven test with mocks
4. **Add migration:** `make migrate-create name=description`

### Most Common Imports
```go
// Logging
"go.uber.org/zap"
"library-service/pkg/logutil"

// Errors
"library-service/pkg/errors"

// HTTP
"github.com/go-chi/chi/v5"
"library-service/pkg/httputil"

// Database
"github.com/jmoiron/sqlx"

// Testing
"github.com/stretchr/testify/assert"
"github.com/stretchr/testify/mock"
"github.com/stretchr/testify/require"
```

---

**Note:** See `examples/` directory for complete working examples of each pattern.
