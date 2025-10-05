# Codebase Map

> **Find any file in under 10 seconds**

## Purpose

This map shows you EXACTLY where to find code for specific tasks. No more grepping through hundreds of files.

**Use this when:** You need to find similar code to reference or understand where new code should go.

---

## 🗺️ Quick Navigation

**Jump to:**
- [By Domain Entity](#by-domain-entity) - "Where is book/member/loan code?"
- [By Layer](#by-layer) - "Where are handlers/use cases/repositories?"
- [By Task](#by-task) - "I want to add X, where do I look?"
- [By File Type](#by-file-type) - "Where are tests/migrations/configs?"

---

## 📦 By Domain Entity

### Book

```
Domain Layer:
├── internal/domain/book/entity.go              ← Book struct, NewEntity()
├── internal/domain/book/service.go             ← ValidateISBN(), business logic
├── internal/domain/book/repository.go          ← Repository interface
├── internal/domain/book/errors.go              ← Domain-specific errors
└── internal/domain/book/*_test.go              ← Domain tests (100% coverage)

Use Case Layer:
├── internal/usecase/bookops/create_book.go     ← CreateBookUseCase
├── internal/usecase/bookops/update_book.go     ← UpdateBookUseCase
├── internal/usecase/bookops/delete_book.go     ← DeleteBookUseCase
├── internal/usecase/bookops/get_book.go        ← GetBookUseCase
├── internal/usecase/bookops/list_books.go      ← ListBooksUseCase
└── internal/usecase/bookops/*_test.go          ← Use case tests

Adapter Layer (HTTP):
├── internal/adapters/http/handlers/book.go     ← HTTP handlers
├── internal/adapters/http/dto/book.go          ← Request/Response DTOs
└── internal/adapters/http/handlers/book_test.go ← Handler tests

Adapter Layer (Repository):
├── internal/adapters/repository/postgres/book.go       ← PostgreSQL implementation
├── internal/adapters/repository/postgres/book_test.go  ← Repository integration tests
└── internal/adapters/repository/memory/book.go         ← In-memory (for tests)

Migrations:
├── migrations/postgres/000001_create_books.up.sql
└── migrations/postgres/000001_create_books.down.sql
```

---

### Member

```
Domain Layer:
├── internal/domain/member/entity.go            ← Member struct
├── internal/domain/member/service.go           ← Subscription logic, eligibility
├── internal/domain/member/repository.go        ← Repository interface
└── internal/domain/member/*_test.go

Use Case Layer:
├── internal/usecase/memberops/create_member.go
├── internal/usecase/memberops/update_member.go
├── internal/usecase/memberops/get_member.go
└── internal/usecase/memberops/*_test.go

Adapter Layer (HTTP):
├── internal/adapters/http/handlers/member.go
├── internal/adapters/http/dto/member.go
└── internal/adapters/http/handlers/member_test.go

Adapter Layer (Repository):
├── internal/adapters/repository/postgres/member.go
└── internal/adapters/repository/memory/member.go

Migrations:
├── migrations/postgres/000002_create_members.up.sql
└── migrations/postgres/000002_create_members.down.sql
```

---

### Author

```
Domain Layer:
├── internal/domain/author/entity.go
├── internal/domain/author/repository.go
└── internal/domain/author/*_test.go

Use Case Layer:
├── internal/usecase/authorops/create_author.go
├── internal/usecase/authorops/get_author.go
└── internal/usecase/authorops/*_test.go

Adapter Layer (HTTP):
├── internal/adapters/http/handlers/author.go
├── internal/adapters/http/dto/author.go

Adapter Layer (Repository):
├── internal/adapters/repository/postgres/author.go

Migrations:
├── migrations/postgres/000003_create_authors.up.sql
└── migrations/postgres/000003_create_authors.down.sql
```

---

### Loan (Future - Not Yet Implemented)

**When implementing, follow this structure:**
```
Domain Layer:
├── internal/domain/loan/entity.go              ← Loan struct, NewLoan()
├── internal/domain/loan/service.go             ← CalculateLateFee(), IsOverdue()
├── internal/domain/loan/repository.go          ← Repository interface
└── internal/domain/loan/*_test.go

Use Case Layer:
├── internal/usecase/loanops/create_loan.go     ← Borrow book
├── internal/usecase/loanops/return_book.go     ← Return book, calculate fees
├── internal/usecase/loanops/list_overdue.go    ← Get overdue loans
└── internal/usecase/loanops/*_test.go

Adapter Layer:
├── internal/adapters/http/handlers/loan.go
├── internal/adapters/http/dto/loan.go
├── internal/adapters/repository/postgres/loan.go

Migrations:
├── migrations/postgres/XXXXXX_create_loans.up.sql
└── migrations/postgres/XXXXXX_create_loans.down.sql
```

**See:** [examples/README.md](./examples/README.md#adding-a-new-domain-loan) for complete implementation example

---

### Subscription

```
Domain Layer:
├── internal/domain/subscription/entity.go      ← Subscription struct
├── internal/domain/subscription/service.go     ← CalculateProRatedCost()
├── internal/domain/subscription/repository.go
└── internal/domain/subscription/*_test.go

Use Case Layer:
├── internal/usecase/subops/subscribe_member.go ← Upgrade subscription
└── internal/usecase/subops/*_test.go

Adapter Layer:
├── internal/adapters/http/handlers/subscription.go
├── internal/adapters/repository/postgres/subscription.go
```

---

## 🏗️ By Layer

### Domain Layer (Pure Business Logic)

```
internal/domain/
├── book/
│   ├── entity.go           ← Book struct, constructor
│   ├── service.go          ← Business rules (ValidateISBN, etc.)
│   ├── repository.go       ← Repository interface
│   ├── errors.go           ← Domain errors
│   └── *_test.go           ← Tests (100% coverage, NO mocks)
├── member/
│   └── ... (same structure)
├── author/
│   └── ... (same structure)
└── errors/
    └── errors.go           ← Common domain errors (ErrNotFound, etc.)
```

**Rule:** Domain has ZERO external dependencies (only stdlib and other domain packages)

---

### Use Case Layer (Orchestration)

```
internal/usecase/
├── bookops/                ← "ops" suffix to avoid conflict with domain/book
│   ├── create_book.go      ← One file per use case
│   ├── update_book.go
│   ├── get_book.go
│   ├── list_books.go
│   └── *_test.go           ← Tests (80%+ coverage, mock repositories)
├── memberops/
│   └── ...
├── authops/                ← Authentication use cases
│   ├── register.go
│   ├── login.go
│   └── refresh_token.go
└── subops/                 ← Subscription use cases
    └── subscribe_member.go
```

**Rule:** Use cases orchestrate, don't implement business logic. Business logic → domain service.

---

### Adapter Layer - HTTP

```
internal/adapters/http/
├── handlers/
│   ├── book.go             ← BookHandler (CreateBook, GetBook, etc.)
│   ├── member.go           ← MemberHandler
│   ├── author.go           ← AuthorHandler
│   ├── auth.go             ← AuthHandler (login, register)
│   └── *_test.go           ← Handler tests (mock use cases)
├── dto/
│   ├── book.go             ← Request/Response structs for Book
│   ├── member.go
│   ├── error.go            ← ErrorResponse struct
│   └── response.go         ← Standard response wrapper
├── middleware/
│   ├── auth.go             ← JWT authentication middleware
│   ├── cors.go             ← CORS configuration
│   ├── logging.go          ← Request/response logging
│   └── recovery.go         ← Panic recovery
└── routes/
    └── router.go           ← Route definitions (chi router)
```

---

### Adapter Layer - Repository

```
internal/adapters/repository/
├── postgres/
│   ├── book.go             ← PostgreSQL implementation of book.Repository
│   ├── member.go
│   ├── author.go
│   └── *_test.go           ← Integration tests (use real Postgres)
└── memory/
    ├── book.go             ← In-memory implementation (for tests)
    ├── member.go
    └── author.go
```

**Rule:** Repository implements domain interface. One file per entity.

---

### Infrastructure Layer

```
internal/infrastructure/
├── app/
│   └── app.go              ← Bootstrap infrastructure (DB, Redis, JWT)
├── container/
│   └── container.go        ← Wire use cases, handlers (Dependency Injection)
├── auth/
│   ├── jwt.go              ← JWT token generation/validation
│   └── password.go         ← Password hashing (bcrypt)
├── store/
│   └── postgres.go         ← PostgreSQL connection pool
├── cache/
│   └── redis.go            ← Redis client (future)
└── config/
    └── config.go           ← Configuration loading (env vars)
```

**Key Files:**
- `app/app.go` → Initialize DB, Redis, external services
- `container/container.go` → Wire everything together (repos, services, use cases, handlers)

---

## 🎯 By Task

### "I want to add a new API endpoint"

**Order of files to touch:**

1. **Use Case** (if doesn't exist):
   ```
   internal/usecase/bookops/search_books.go
   internal/usecase/bookops/search_books_test.go
   ```

2. **DTO** (request/response):
   ```
   internal/adapters/http/dto/book.go
   Add: SearchBooksRequest, SearchBooksResponse
   ```

3. **Handler**:
   ```
   internal/adapters/http/handlers/book.go
   Add method: func (h *BookHandler) SearchBooks(w http.ResponseWriter, r *http.Request)
   Add Swagger annotations
   ```

4. **Routes**:
   ```
   internal/adapters/http/routes/router.go
   Add: r.Get("/books/search", handlers.Book.SearchBooks)
   ```

5. **Swagger**:
   ```bash
   make gen-docs
   ```

6. **Tests**:
   ```
   internal/adapters/http/handlers/book_test.go
   Add: TestBookHandler_SearchBooks
   ```

**See:** [examples/README.md](./examples/README.md#adding-a-new-api-endpoint) for complete code

---

### "I want to add business logic"

**Go to:**
```
internal/domain/{entity}/service.go
```

**Example: Add ISBN validation for Book**
```
File: internal/domain/book/service.go

func (s *Service) ValidateISBN(isbn string) error {
    // Business logic here
}
```

**Then write tests:**
```
File: internal/domain/book/service_test.go

func TestService_ValidateISBN(t *testing.T) {
    // Test with NO mocks (pure logic)
}
```

---

### "I want to add a database query"

**Two steps:**

1. **Define in domain interface:**
   ```
   File: internal/domain/book/repository.go

   type Repository interface {
       GetByISBN(ctx context.Context, isbn string) (Entity, error)
   }
   ```

2. **Implement in adapter:**
   ```
   File: internal/adapters/repository/postgres/book.go

   func (r *BookRepository) GetByISBN(ctx context.Context, isbn string) (book.Entity, error) {
       query := "SELECT * FROM books WHERE isbn = $1"
       // ...
   }
   ```

**Test:**
```
File: internal/adapters/repository/postgres/book_test.go

func TestBookRepository_GetByISBN(t *testing.T) {
    // Integration test with real database
}
```

---

### "I want to add authentication to an endpoint"

**Two steps:**

1. **Add middleware to route:**
   ```
   File: internal/adapters/http/routes/router.go

   r.Route("/books", func(r chi.Router) {
       r.Use(authMiddleware)  // ← Add this
       r.Post("/", handlers.Book.CreateBook)
   })
   ```

2. **Add Swagger annotation:**
   ```
   File: internal/adapters/http/handlers/book.go

   // @Security BearerAuth  ← Add this
   // @Router /books [post]
   func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request)
   ```

**Get user from request:**
```go
claims := auth.GetClaimsFromContext(r.Context())
memberID := claims.MemberID
```

---

### "I want to add a database migration"

```bash
# Create migration files
make migrate-create name=add_loans_table
```

**Files created:**
```
migrations/postgres/
├── XXXXXX_add_loans_table.up.sql    ← Edit this (CREATE TABLE)
└── XXXXXX_add_loans_table.down.sql  ← Edit this (DROP TABLE)
```

**Apply:**
```bash
make migrate-up
```

**Rollback:**
```bash
make migrate-down
```

---

## 📂 By File Type

### Tests

```
Unit Tests (domain, no mocks):
internal/domain/book/*_test.go
internal/domain/member/*_test.go

Use Case Tests (mock repositories):
internal/usecase/bookops/*_test.go
internal/usecase/authops/*_test.go

Integration Tests (real database):
internal/adapters/repository/postgres/*_test.go

Handler Tests (mock use cases):
internal/adapters/http/handlers/*_test.go
```

**Run specific tests:**
```bash
# Domain tests
go test ./internal/domain/book/

# Use case tests
go test ./internal/usecase/bookops/

# Integration tests (requires DB)
go test ./internal/adapters/repository/postgres/

# All tests
make test
```

---

### Migrations

```
migrations/postgres/
├── 000001_create_books.up.sql
├── 000001_create_books.down.sql
├── 000002_create_members.up.sql
├── 000002_create_members.down.sql
├── 000003_create_authors.up.sql
├── 000003_create_authors.down.sql
└── ... (numbered sequentially)
```

**Naming:** `{number}_{description}.{up|down}.sql`

---

### Configuration

```
Root:
├── .env.example            ← Example environment variables
├── docker-compose.yml      ← Local development (Postgres, Redis)
├── Makefile                ← All commands (make help)
└── go.mod                  ← Go dependencies

Config:
├── internal/infrastructure/config/config.go  ← Load config from env

Scripts:
├── .claude/scripts/review.sh  ← Pre-commit checks
```

---

### Documentation

```
.claude/
├── README.md               ← Start here
├── context-guide.md        ← What to read for each task
├── glossary.md             ← Domain terms
├── codebase-map.md         ← This file
├── examples/               ← Code examples
├── adrs/                   ← Architecture decisions
└── ... (20+ other guides)
```

---

## 🔍 Finding Code Examples

### "How do I create a use case?"

**Look at:**
```
internal/usecase/bookops/create_book.go
internal/usecase/authops/login.go
```

**Pattern:**
```go
type CreateBookUseCase struct {
    repo    book.Repository
    service *book.Service
}

func NewCreateBookUseCase(repo book.Repository, svc *book.Service) *CreateBookUseCase {
    return &CreateBookUseCase{repo: repo, service: svc}
}

func (uc *CreateBookUseCase) Execute(ctx context.Context, req Request) (*book.Entity, error) {
    // Orchestration logic
}
```

---

### "How do I write handler tests?"

**Look at:**
```
internal/adapters/http/handlers/book_test.go
```

**Pattern:**
```go
func TestBookHandler_CreateBook(t *testing.T) {
    mockUC := &mocks.MockCreateBookUseCase{}
    handler := handlers.NewBookHandler(mockUC, ...)

    req := httptest.NewRequest("POST", "/books", body)
    w := httptest.NewRecorder()

    handler.CreateBook(w, req)

    assert.Equal(t, http.StatusCreated, w.Code)
}
```

---

### "How do I implement a repository?"

**Look at:**
```
internal/adapters/repository/postgres/book.go
```

**Pattern:**
```go
type BookRepository struct {
    db *pgxpool.Pool
}

func NewBookRepository(db *pgxpool.Pool) book.Repository {
    return &BookRepository{db: db}
}

func (r *BookRepository) Create(ctx context.Context, book book.Entity) error {
    query := `INSERT INTO books (id, title, isbn) VALUES ($1, $2, $3)`
    _, err := r.db.Exec(ctx, query, book.ID, book.Title, book.ISBN)
    return err
}
```

---

## 🛠️ Common File Patterns

### Entry Points

```
cmd/
├── api/main.go             ← API server entry point
├── migrate/main.go         ← Database migration tool
└── worker/main.go          ← Background worker (future)
```

**Start API:**
```bash
go run cmd/api/main.go
```

---

### Dependency Wiring

```
internal/infrastructure/container/container.go
```

**This file wires everything:**
- Repositories (PostgreSQL implementations)
- Domain services
- Use cases
- Handlers

**When to edit:** Adding new use case or handler

---

### Routes

```
internal/adapters/http/routes/router.go
```

**All HTTP routes defined here:**
```go
r.Route("/api/v1", func(r chi.Router) {
    r.Route("/books", func(r chi.Router) {
        r.Post("/", handlers.Book.CreateBook)
        r.Get("/{id}", handlers.Book.GetBook)
    })
})
```

---

## 💡 Pro Tips

1. **Use file structure to navigate:**
   ```bash
   # Find all use cases for books
   find internal/usecase/bookops/ -name "*.go"

   # Find all handlers
   ls internal/adapters/http/handlers/
   ```

2. **Grep for examples:**
   ```bash
   # Find how we create entities
   grep -r "NewEntity" internal/domain/

   # Find all use case constructors
   grep -r "UseCase struct" internal/usecase/
   ```

3. **Check tests for usage:**
   ```bash
   # See how BookRepository is used
   grep -r "BookRepository" internal/usecase/bookops/*_test.go
   ```

4. **Follow the pattern:**
   - Look at `book/` implementation
   - Copy structure for new entity (e.g., `loan/`)
   - Adapt to your needs

---

**Last Updated:** 2025-01-19

**Next Review:** When adding new domain entities or restructuring code
