# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 🚨 NEW CLAUDE CODE INSTANCE? **[START HERE → .claude/CLAUDE-START.md](./.claude/CLAUDE-START.md)**

> **📚 Full Documentation:** See [`.claude/`](./.claude/) directory for comprehensive guides
>
> **🎯 Not sure what to read?** Check [Context Guide](./.claude/context-guide.md) for task-specific reading lists
>
> **⚡ Quick Navigation:**
> - **New to codebase?** Read `.claude/CLAUDE-START.md` (60-second boot sequence) ⭐
> - **Adding features?** See `.claude/development-workflows.md` for step-by-step workflows
> - **Refactoring?** Check `.claude/quick-wins.md` for safe improvements
> - **Architecture questions?** See `.claude/architecture.md`
> - **Testing?** See `.claude/testing.md` for patterns and strategies
> - **Latest changes?** See `.claude/LEGACY_CODE_REMOVAL.md` and `.claude/HANDLER_REFACTORING_FINAL.md` ⭐

## Project Overview

Library Management System - A Go-based REST API following Clean Architecture principles, optimized for vibecoding with Claude Code. The system manages books, authors, members, subscriptions, reservations, and payments with JWT authentication and epayment.kz payment gateway integration.

**Key Technologies:** Go 1.25, PostgreSQL 15+, Redis 7+, Chi router, JWT, Docker, Swagger/OpenAPI

## Architecture (Clean Architecture)

The codebase follows strict dependency rules: **Domain → Use Case → Adapters → Infrastructure**

```
internal/
├── domain/              # Business logic (ZERO external dependencies)
│   ├── book/           # Book entity, service, repository interface
│   ├── member/         # Member entity, service (subscriptions)
│   ├── author/         # Author entity
│   ├── reservation/    # Reservation entity, service
│   └── payment/        # Payment entity, service (payments, receipts)
├── usecase/            # Application orchestration (depends on domain)
│   ├── bookops/        # CreateBook, UpdateBook, etc. ("ops" suffix to avoid naming conflicts)
│   ├── authops/        # Register, Login, RefreshToken ("ops" suffix)
│   ├── subops/         # SubscribeMember ("ops" suffix)
│   ├── reservationops/ # CreateReservation, CancelReservation ("ops" suffix)
│   └── paymentops/     # Payment operations (18 use cases) ("ops" suffix)
├── adapters/           # External interfaces (HTTP, DB, cache, payment)
│   ├── http/           # Chi handlers, middleware, DTOs
│   ├── repository/     # PostgreSQL implementations
│   ├── cache/          # Redis/Memory implementations
│   └── payment/        # Payment gateway adapters (epayment.kz)
└── infrastructure/     # Technical concerns
    ├── auth/           # JWT token generation/validation
    ├── store/          # Database connections
    └── server/         # HTTP server configuration
```

**Critical Rules:**
- Domain layer must NEVER import from outer layers
- Use case packages use **"ops" suffix** (e.g., `bookops`) to avoid naming conflicts with domain packages (e.g., `book`)
- Use cases define behavior via interfaces, adapters provide implementations
- Business logic lives in **domain services**, NOT in use cases
- Infrastructure services (JWT, Password) created in `app.go`, domain services (Book, Member, Payment) created in `container.go`

## Request Flow

This flowchart shows the complete lifecycle of an HTTP request through the Clean Architecture layers:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           HTTP REQUEST (Client)                              │
│                    GET /api/v1/books?genre=fiction                          │
│                    Authorization: Bearer <jwt_token>                         │
└──────────────────────────────────┬──────────────────────────────────────────┘
                                   │
                                   ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  LAYER 1: HTTP SERVER (Chi Router)                                          │
│  Location: internal/adapters/http/router.go                                 │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────┐       │
│  │  Middleware Chain (executed in order):                          │       │
│  │  1. RequestID        → Generate unique request ID               │       │
│  │  2. RealIP           → Extract real client IP                   │       │
│  │  3. RequestLogger    → Log request details                      │       │
│  │  4. ErrorHandler     → Catch panics, standardize errors         │       │
│  │  5. Recoverer        → Recover from panics                      │       │
│  │  6. Timeout          → Enforce request timeout (30s default)    │       │
│  │  7. AuthMiddleware   → Validate JWT, extract member ID          │       │
│  │                         (Protected routes only)                  │       │
│  └─────────────────────────────────────────────────────────────────┘       │
│                                                                              │
│  Route matching: /api/v1/books → BookHandler.Routes()                       │
└──────────────────────────────────┬──────────────────────────────────────────┘
                                   │
                                   ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  LAYER 2: HTTP HANDLER (Adapter)                                            │
│  Location: internal/adapters/http/handlers/book.go                          │
│                                                                              │
│  BookHandler.ListBooks(w http.ResponseWriter, r *http.Request)              │
│  ┌──────────────────────────────────────────────────────────────┐          │
│  │  1. Extract query parameters (genre, limit, offset)          │          │
│  │  2. Validate input using validator.ValidateStruct()          │          │
│  │  3. Build DTO request object (bookops.ListBooksRequest)      │          │
│  │  4. Call use case: useCases.ListBooks.Execute(ctx, req)      │          │
│  │  5. Convert domain entities to DTOs (book.ToDTO())           │          │
│  │  6. Respond with JSON: RespondJSON(w, 200, response)         │          │
│  └──────────────────────────────────────────────────────────────┘          │
│                                                                              │
│  Error handling: All errors converted to HTTP status codes                  │
│  - ErrNotFound → 404                                                        │
│  - ErrValidation → 400                                                      │
│  - ErrUnauthorized → 401                                                    │
│  - ErrAlreadyExists → 409                                                   │
│  - Unknown errors → 500                                                     │
└──────────────────────────────────┬──────────────────────────────────────────┘
                                   │
                                   ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  LAYER 3: USE CASE (Application Logic)                                      │
│  Location: internal/usecase/bookops/list_books.go                           │
│                                                                              │
│  ListBooksUseCase.Execute(ctx context.Context, req ListBooksRequest)        │
│  ┌──────────────────────────────────────────────────────────────┐          │
│  │  Dependencies (injected via container.go):                    │          │
│  │  - bookRepo: book.Repository (PostgreSQL implementation)      │          │
│  │                                                                │          │
│  │  Execution flow:                                               │          │
│  │  1. Log use case start (logutil.UseCaseLogger)               │          │
│  │  2. Build repository query filters (genre, limit, offset)    │          │
│  │  3. Call repository: bookRepo.List(ctx, filters)             │          │
│  │  4. Return domain entities ([]book.Book)                     │          │
│  │  5. Log use case completion                                   │          │
│  └──────────────────────────────────────────────────────────────┘          │
│                                                                              │
│  Note: Use cases orchestrate operations but contain NO business logic       │
│        Business rules live in domain services (e.g., book.Service)          │
└──────────────────────────────────┬──────────────────────────────────────────┘
                                   │
                                   ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  LAYER 4: REPOSITORY (Data Access Adapter)                                  │
│  Location: internal/adapters/repository/postgres/book.go                    │
│                                                                              │
│  PostgresBookRepository.List(ctx, filters) ([]book.Book, error)             │
│  ┌──────────────────────────────────────────────────────────────┐          │
│  │  1. Build SQL query with filters (WHERE genre = $1)          │          │
│  │  2. Execute query: db.SelectContext(ctx, &books, query)      │          │
│  │  3. Map database rows to domain entities                     │          │
│  │  4. Return []book.Book (domain objects, NOT DTOs)            │          │
│  └──────────────────────────────────────────────────────────────┘          │
│                                                                              │
│  Uses BaseRepository[Book] for common CRUD operations                       │
│  Custom methods for complex queries (e.g., search, filters)                 │
└──────────────────────────────────┬──────────────────────────────────────────┘
                                   │
                                   ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│  LAYER 5: DATABASE (PostgreSQL)                                             │
│  Location: Docker container (localhost:5432)                                │
│                                                                              │
│  Query execution:                                                            │
│  SELECT id, name, genre, isbn, created_at, updated_at                       │
│  FROM books                                                                  │
│  WHERE genre = 'fiction'                                                     │
│  LIMIT 20 OFFSET 0;                                                          │
│                                                                              │
│  Returns rows to repository layer                                           │
└──────────────────────────────────┬──────────────────────────────────────────┘
                                   │
                                   ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         RESPONSE PATH (Unwinding)                            │
│                                                                              │
│  Repository → Use Case → Handler → HTTP Server → Client                     │
│                                                                              │
│  Data transformations:                                                       │
│  - DB rows        → Domain entities (book.Book)                             │
│  - Domain entities → DTOs (dto.BookResponse)                                │
│  - DTOs           → JSON response                                           │
│                                                                              │
│  HTTP Response:                                                              │
│  Status: 200 OK                                                              │
│  Content-Type: application/json; charset=utf-8                              │
│  Body:                                                                       │
│  {                                                                           │
│    "books": [                                                                │
│      {                                                                       │
│        "id": "uuid-1",                                                       │
│        "name": "The Great Gatsby",                                           │
│        "genre": "fiction",                                                   │
│        "isbn": "978-0-7432-7356-5"                                           │
│      },                                                                      │
│      ...                                                                     │
│    ],                                                                        │
│    "total": 42,                                                              │
│    "limit": 20,                                                              │
│    "offset": 0                                                               │
│  }                                                                           │
└─────────────────────────────────────────────────────────────────────────────┘


KEY OBSERVATIONS:

1. **Dependency Direction (Clean Architecture)**:
   - HTTP layer depends on Use Case layer (imports usecase package)
   - Use Case layer depends on Domain layer (imports domain interfaces)
   - Repository layer depends on Domain layer (implements domain interfaces)
   - Domain layer depends on NOTHING (pure business logic)

2. **Data Flow**:
   - Inward: HTTP Request → DTO → Domain Entity
   - Outward: Domain Entity → DTO → HTTP Response

3. **Error Handling**:
   - Domain errors (e.g., ErrNotFound) propagate up unchanged
   - Each layer adds context with fmt.Errorf("...: %w", err)
   - HTTP handler converts domain errors to HTTP status codes

4. **Cross-Cutting Concerns**:
   - Logging: Context-based (RequestLogger middleware + logutil helpers)
   - Authentication: JWT validation in middleware, member ID in context
   - Validation: Input validation in handlers (validator), business validation in domain services

5. **Caching (Optional Path)**:
   - For GET operations, use cases may check cache before repository
   - Cache hit → return cached domain entity
   - Cache miss → query repository → store in cache → return

6. **Domain Service Usage** (Example: CreateBook):
   ```
   Handler → CreateBookUseCase → bookService.ValidateISBN()
                                → bookRepo.Create()
                                → cache.Invalidate()
   ```
   Domain service called for business rule validation BEFORE persistence

For detailed layer documentation:
- Entry point: cmd/api/main.go (boot sequence)
- Dependency wiring: internal/usecase/container.go (comprehensive guide)
- HTTP routing: internal/adapters/http/router.go
- Middleware: internal/adapters/http/middleware/
```

## Common Commands

### Building
```bash
make build              # Build all binaries (api, worker, migrate)
make build-api          # Build API server only → bin/library-api
make build-worker       # Build worker only → bin/library-worker
make build-migrate      # Build migration tool → bin/library-migrate
```

### Running Locally
```bash
# Full stack (recommended for development)
make dev                # Starts docker services + migrations + API server

# Individual services
make run                # Run API server (requires PostgreSQL/Redis running)
make run-worker         # Run background worker
make up                 # Start docker-compose (PostgreSQL + Redis)
make down               # Stop docker services

# Quick start (5 minutes)
make init && make up && make migrate-up && make run
```

### Testing
```bash
make test               # All tests with race detection + coverage
make test-unit          # Unit tests only (fast, no database)
make test-integration   # Integration tests (requires database)
make test-coverage      # Generate HTML coverage report

# Run specific package tests
go test -v ./internal/domain/book/...
go test -v -run TestCreateBook ./internal/usecase/bookops/
```

### Code Quality
```bash
make ci                 # Full CI pipeline: fmt → vet → lint → test → build
make lint               # Run golangci-lint (25+ linters enabled)
make fmt                # Format code with gofmt + goimports
make vet                # Run go vet for suspicious constructs
```

### Database Migrations
```bash
make migrate-up         # Apply all pending migrations
make migrate-down       # Rollback last migration
make migrate-create name=add_book_ratings  # Create new migration

# Direct usage (requires POSTGRES_DSN env var)
go run cmd/migrate/main.go up
go run cmd/migrate/main.go down
POSTGRES_DSN="postgres://library:library123@localhost:5432/library?sslmode=disable" go run cmd/migrate/main.go up
```

### Development Tools
```bash
make install-tools      # Install golangci-lint, mockgen, swag
make gen-mocks          # Generate test mocks
make gen-docs           # Generate Swagger/OpenAPI docs
make benchmark          # Run performance benchmarks
```

## API Documentation

**Swagger UI:** http://localhost:8080/swagger/index.html (when server is running)

**Regenerating Swagger Documentation:**
```bash
# Full regeneration (recommended)
make gen-docs

# Manual regeneration with dependency parsing
swag init -g cmd/api/main.go -o api/openapi --parseDependency --parseInternal
```

**Important Swagger Annotations:**
- `@Summary` - Brief description (required)
- `@Description` - Detailed explanation
- `@Tags` - Group endpoints together
- `@Security BearerAuth` - **Required for all protected endpoints**
- `@Param` - Request parameters (body, path, query, header)
- `@Success` / `@Failure` - Response codes with schemas
- `@Router` - Endpoint path and HTTP method

## Development Workflow

### Adding a New Feature

**Follow this order:** Domain → Use Case → Adapters → Wiring → Migration → Documentation

See `.claude/common-tasks.md` for complete step-by-step guides.

**Quick Example: Adding a "Loan" domain**

1. **Domain Layer** (`internal/domain/loan/`):
   - Create entity, service, repository interface
   - Write unit tests with 100% coverage

2. **Use Case Layer** (`internal/usecase/loanops/`):
   - Note: "ops" suffix to avoid naming conflicts
   - Create use cases that orchestrate domain services
   - Mock repositories in tests

3. **Adapter Layer**:
   - Implement repository (`internal/adapters/repository/postgres/loan.go`)
   - Create HTTP handlers (`internal/adapters/http/handlers/loan.go`)
   - Add DTOs and Swagger annotations

4. **Wire Dependencies** (`internal/usecase/container.go`):
   - Add repository to `Repositories` struct
   - Add use cases to `Container` struct
   - Wire in `NewContainer()` function

5. **Database Migration**:
   ```bash
   make migrate-create name=create_loans_table
   make migrate-up
   ```

6. **Update Documentation**:
   ```bash
   make gen-docs
   ```

## Code Consistency Patterns (CRITICAL)

### Context Value Access - ALWAYS Use Helper Functions

**❌ WRONG - Direct context access (causes type safety issues):**
```go
memberID, ok := ctx.Value("member_id").(string)
if !ok || memberID == "" {
    h.respondError(w, r, errors.ErrUnauthorized.WithDetails("reason", "member_id not found"))
    return
}
```

**✅ CORRECT - Use helper functions:**
```go
import "library-service/internal/adapters/http/middleware"

memberID, ok := middleware.GetMemberIDFromContext(ctx)
if !ok {
    h.respondError(w, r, errors.ErrUnauthorized)
    return
}
```

**Available helper functions:**
- `middleware.GetMemberIDFromContext(ctx)` - Extract authenticated member ID
- `middleware.GetMemberEmailFromContext(ctx)` - Extract member email
- `middleware.GetMemberRoleFromContext(ctx)` - Extract member role
- `middleware.GetClaimsFromContext(ctx)` - Extract full JWT claims

**Why this matters:** Type safety, less code, consistent error handling, easier to refactor.

### Status Code Checks - ALWAYS Use httputil

**❌ WRONG - Magic numbers:**
```go
if status >= 500 {
    logger.Error("server error")
}
if status >= 400 {
    logger.Warn("client error")
}
```

**✅ CORRECT - Self-documenting:**
```go
import "library-service/pkg/httputil"

if httputil.IsServerError(status) {
    logger.Error("server error")
}
if httputil.IsClientError(status) {
    logger.Warn("client error")
}
```

**Available functions:**
- `httputil.IsServerError(code)` - 5xx status codes
- `httputil.IsClientError(code)` - 4xx status codes
- `httputil.IsSuccess(code)` - 2xx status codes
- `httputil.IsRedirect(code)` - 3xx status codes

### Validator - ALWAYS Inject as Dependency

**❌ WRONG - Creating validator in handler:**
```go
func NewBookHandler(createBookUC *bookops.CreateBookUseCase) *BookHandler {
    return &BookHandler{
        createBookUC: createBookUC,
        validator:    middleware.NewValidator(), // ❌ Created here
    }
}
```

**✅ CORRECT - Inject as dependency:**
```go
func NewBookHandler(
    createBookUC *bookops.CreateBookUseCase,
    validator *middleware.Validator, // ✅ Injected
) *BookHandler {
    return &BookHandler{
        createBookUC: createBookUC,
        validator:    validator,
    }
}
```

**Why this matters:** Testability (can mock validator), follows DI pattern, single instance.

---

## Key Implementation Patterns

### 1. Package Naming Convention

**Use Case Packages Use "ops" Suffix:**
- Domain: `internal/domain/book` (package `book`)
- Use Case: `internal/usecase/bookops` (package `bookops`)

**Rationale:**
- Avoids naming conflicts when importing both domain and use case packages
- No need for import aliases (cleaner, more idiomatic Go)
- Clear distinction: domain = entities/business rules, use cases = operations/orchestration

```go
import (
    "library-service/internal/domain/book"      // package book
    "library-service/internal/usecase/bookops"  // package bookops
)

// Clean references without aliases
bookEntity := book.NewEntity(...)
useCase := bookops.NewCreateBookUseCase(...)
```

### 2. Dependency Injection (Grouped Container Structure)

**Container Organization** (`internal/usecase/container.go`)

The container uses **domain-grouped structure** (refactored October 2025):

```go
type Container struct {
    Book         BookUseCases        // CreateBook, GetBook, ListBooks, UpdateBook, DeleteBook, ListBookAuthors
    Author       AuthorUseCases      // ListAuthors
    Auth         AuthUseCases        // RegisterMember, LoginMember, RefreshToken, ValidateToken
    Member       MemberUseCases      // ListMembers, GetMemberProfile
    Subscription SubscriptionUseCases // SubscribeMember
    Reservation  ReservationUseCases  // CreateReservation, CancelReservation, GetReservation, ListMemberReservations
    Payment      PaymentUseCases      // InitiatePayment, VerifyPayment, CancelPayment, RefundPayment, etc. (9 use cases)
    SavedCard    SavedCardUseCases    // SaveCard, ListSavedCards, DeleteSavedCard, SetDefaultCard
    Receipt      ReceiptUseCases      // GenerateReceipt, GetReceipt, ListReceipts
}
```

**Handler Access Pattern:**
```go
type BookHandler struct {
    useCases *usecase.Container  // Grouped container
}

func (h *BookHandler) create(...) {
    h.useCases.Book.CreateBook.Execute(...)  // Domain.UseCase.Execute
}
```

**When adding new features:**
1. Add repository interface to `Repositories` struct
2. Add use cases to appropriate domain group in `Container` struct
3. Create domain factory function (e.g., `newBookUseCases()`)
4. Wire in `NewContainer()` function

**Critical Distinction:**
- **Infrastructure Services** (JWT, Password, Gateway): Created in `app.go`, passed to container
- **Domain Services** (Book, Member, Payment): Created in domain factories within `container.go`

### 3. Domain Services vs Use Cases

**Domain Service** (`internal/domain/book/service.go`):
- Pure business rules (ISBN validation, constraints)
- NO external dependencies (no database, HTTP, frameworks)
- Pure functions when possible
- 100% test coverage (easy to achieve)

**Use Case** (`internal/usecase/bookops/create_book.go`):
- Orchestrates domain entities and services
- Calls domain service for validation
- Persists to repository
- Updates cache
- Returns domain entities (not DTOs)

### 4. Repository Pattern

**Interface:** Defined in `internal/domain/{entity}/repository.go`
**Implementation:** In `internal/adapters/repository/{type}/{entity}.go`

**Benefits:**
- Domain is independent of database technology
- Easy to swap PostgreSQL for MongoDB (just change adapter)
- Easy to mock for testing

**Modern Implementation (Phase 3 Refactoring):**

#### 4a. Generic Repository Helpers (ADR 008)

**Use generic helpers for standard CRUD operations:**

```go
import "library-service/internal/adapters/repository/postgres"

func (r *AuthorRepository) GetByID(ctx context.Context, id string) (author.Author, error) {
    return postgres.GetByID[author.Author](ctx, r.db, "authors", id)
}

func (r *AuthorRepository) List(ctx context.Context) ([]author.Author, error) {
    return postgres.List[author.Author](ctx, r.db, "authors", "id")
}

func (r *AuthorRepository) Delete(ctx context.Context, id string) error {
    return postgres.DeleteByID(ctx, r.db, "authors", id)
}
```

**Available helpers:**
- `GetByID[T]` - Retrieve single entity
- `GetByIDWithColumns[T]` - Retrieve with specific columns
- `List[T]` - List all entities
- `ListWithColumns[T]` - List with specific columns
- `DeleteByID` - Delete by ID (with RETURNING verification)
- `ExistsByID` - Check existence
- `CountAll` - Count entities

**Benefits:**
- ✅ 80% code reduction for standard operations
- ✅ Type-safe with Go generics
- ✅ Consistent error handling
- ✅ Tested once, used everywhere

#### 4b. BaseRepository Pattern (ADR 011)

**Use embeddable BaseRepository for minimal boilerplate:**

```go
type AuthorRepository struct {
    postgres.BaseRepository[author.Author]  // Embed base repository
}

func NewAuthorRepository(db *sqlx.DB) *AuthorRepository {
    return &AuthorRepository{
        BaseRepository: postgres.NewBaseRepository[author.Author](db, "authors"),
    }
}

// ✅ Inherited methods (no implementation needed):
// - Get, List, ListWithOrder, Delete
// - Exists, Count, BatchGet
// - GenerateID, Transaction
// - GetDB (for custom queries)

// ✅ Only implement entity-specific methods:
func (r *AuthorRepository) Add(ctx context.Context, a author.Author) (string, error) {
    id := r.GenerateID()  // Use inherited method
    query := `INSERT INTO authors (id, full_name, pseudonym) VALUES ($1, $2, $3)`
    _, err := r.GetDB().ExecContext(ctx, query, id, a.FullName, a.Pseudonym)
    return id, postgres.HandleSQLError(err)
}
```

**Benefits:**
- ✅ 86% code reduction for standard CRUD
- ✅ Built-in transaction support
- ✅ Utility methods (ID generation, existence checks, counting)
- ✅ Override any method when needed
- ✅ Focus on business logic, not boilerplate

**When to use which:**
- **Generic Helpers:** One-off queries, manual control over each call
- **BaseRepository:** New repositories, want maximum code reduction
- **Both:** BaseRepository uses generic helpers internally

### 5. Error Handling

```go
// Wrap errors with context (use %w for unwrapping)
if err := s.repo.Create(ctx, book); err != nil {
    return fmt.Errorf("creating book in repository: %w", err)
}

// Domain errors (defined in pkg/errors/domain.go)
return errors.ErrNotFound          // 404
return errors.ErrAlreadyExists     // 409
return errors.ErrValidation        // 400
```

### 6. Utility Packages (Created in Phase 1-5 Refactoring)

**String Utilities** (`pkg/strutil`):
```go
import "library-service/pkg/strutil"

// Safe string pointer handling
name := strutil.SafeString(book.Name)      // *string → string
ptr := strutil.SafeStringPtr("value")      // string → *string
```

**HTTP Utilities** (`pkg/httputil`):
```go
import "library-service/pkg/httputil"

// Self-documenting status checks instead of magic numbers
if httputil.IsServerError(status) {  // Instead of: status >= 500
    logger.Error("internal error")
}
```

**Logger Utilities** (`pkg/logutil`):
```go
import "library-service/pkg/logutil"

// Use case layer - 3 lines reduced to 1 line
logger := logutil.UseCaseLogger(ctx, "create_book",
    zap.String("isbn", req.ISBN),
)

// Handler layer - automatic structured fields
logger := logutil.HandlerLogger(ctx, "book_handler", "create")

// Repository layer
logger := logutil.RepositoryLogger(ctx, "book", "create")

// Gateway layer
logger := logutil.GatewayLogger(ctx, "epayment", "initiate_payment")
```

**Base Handler** (`internal/adapters/http/handlers`):
```go
// Embed BaseHandler to inherit RespondError and RespondJSON methods
type BookHandler struct {
    BaseHandler  // Provides RespondError() and RespondJSON()
    createBookUC *bookops.CreateBookUseCase
    // ... other fields
}

// Use inherited methods
h.RespondError(w, r, err)
h.RespondJSON(w, http.StatusOK, response)
```

**Constants for Self-Documenting Code:**
```go
// ISBN prefixes (internal/domain/book/service.go)
book.ISBN13PrefixBookland    // "978" - standard ISBN-13 prefix
book.ISBN13PrefixMusicland   // "979" - alternative ISBN-13 prefix

// Payment gateway timeouts (internal/adapters/payment/epayment/gateway.go)
epayment.DefaultHTTPTimeout   // 30 seconds
epayment.TokenExpiryBuffer    // 5 minutes before token refresh
```

## Authentication System

**JWT-based authentication with access/refresh tokens:**

```bash
# Register user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"Test123!@#","full_name":"John Doe"}'

# Login (returns access_token + refresh_token)
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"Test123!@#"}'

# Use access token for protected endpoints
curl -X GET http://localhost:8080/api/v1/books \
  -H "Authorization: Bearer <access_token>"
```

**Token Configuration:**
- Access token: 24h (configurable via `JWT_EXPIRY`)
- Refresh token: 7 days
- Secret key: `JWT_SECRET` environment variable (MUST change in production)

**Protected Endpoints:**
All endpoints under `/api/v1/books/*`, `/api/v1/reservations/*`, `/api/v1/payments/*`, `/api/v1/receipts/*`, and `/api/v1/auth/me` require JWT authentication.

## Payment System

**Integration:** epayment.kz (Kazakhstan payment gateway)

### Payment Flow
1. **Initiate Payment** - Create payment with invoice ID and get widget URL
2. **User Completes Payment** - User redirects to payment gateway widget
3. **Callback Processing** - Gateway sends callback with payment result
4. **Verification** - Verify payment status with gateway
5. **Receipt Generation** - Generate receipt for completed payments

### Key Features
- **Payment Types:** Fines, subscriptions, book purchases
- **Supported Currencies:** KZT, USD, EUR, RUB
- **Payment Methods:** Card, saved card
- **Refunds:** Full and partial refunds supported
- **Receipts:** Auto-generated with unique numbers (RCP-YYYY-NNNNN format)
- **Webhook Retry:** Exponential backoff (1min, 5min, 15min, 1h, 6h) with max 5 retries
- **Payment Expiry:** Automatic expiration of pending payments after timeout
- **Saved Cards:** Store and reuse cards for faster checkout

### Background Worker
The worker (`cmd/worker/main.go`) processes:
- **Payment Expiry Job** - Marks expired pending payments as failed (every 5 minutes)
- **Callback Retry Job** - Retries failed webhook callbacks with exponential backoff (every 1 minute)

Start worker: `make run-worker` or `go run cmd/worker/main.go`

### Payment Gateway Configuration
```bash
# Required environment variables
EPAYMENT_BASE_URL="https://api.epayment.kz"
EPAYMENT_CLIENT_ID="your-client-id"
EPAYMENT_CLIENT_SECRET="your-client-secret"
EPAYMENT_TERMINAL="your-terminal-id"
EPAYMENT_WIDGET_URL="https://widget.epayment.kz"
```

## Environment Configuration

**Setup:**
```bash
cp .env.example .env
# Edit .env with your settings (especially JWT_SECRET, DB credentials)
```

**Critical Variables:**
- `POSTGRES_DSN`: Database connection string
- `JWT_SECRET`: Token signing key (REQUIRED)
- `REDIS_HOST`: Cache server (optional, uses memory cache if unavailable)
- `APP_MODE`: `dev` (verbose logs) or `prod` (JSON logs)
- `EPAYMENT_*`: Payment gateway credentials (required for payment features)

**Docker Development:**
```bash
cd deployments/docker
docker-compose up -d  # PostgreSQL on :5432, Redis on :6379
```

## Testing Guidelines

**Coverage Requirements:**
- Domain layer: 100% (critical business logic)
- Use cases: 80%+
- Overall: 60%+

**Unit Tests (Domain/Use Cases):**
```go
// Table-driven tests (Go standard)
func TestBookService_ValidateISBN(t *testing.T) {
    tests := []struct {
        name    string
        isbn    string
        wantErr bool
    }{
        {"valid ISBN-13", "978-0-306-40615-7", false},
        {"invalid checksum", "978-0-306-40615-8", true},
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

**Integration Tests:**
- Use build tags: `//go:build integration`
- Test against real PostgreSQL (docker-compose)
- Run with: `make test-integration`

See [Testing Guide](./.claude/testing.md) for comprehensive testing strategies.

## Important Files

### Core Architecture Files
- `internal/usecase/container.go` - **CRITICAL**: Dependency injection wiring
- `internal/infrastructure/app/app.go` - **CRITICAL**: Application bootstrap sequence
- `internal/adapters/http/router.go` - HTTP route configuration
- `cmd/api/main.go` - API entry point and Swagger metadata
- `cmd/worker/main.go` - Background worker (payment expiry, callback retries)
- `cmd/migrate/main.go` - Migration tool entry point

### Utility Packages (Refactoring Phases 1-5)
- `pkg/strutil/` - Safe string pointer utilities
- `pkg/httputil/` - HTTP status code constants and helpers
- `pkg/logutil/` - Logger initialization utilities (UseCaseLogger, HandlerLogger, etc.)
- `pkg/errors/` - Custom error types and domain errors
- `internal/adapters/http/handlers/base.go` - Shared response methods for handlers

### Payment System
- `internal/adapters/payment/epayment/gateway.go` - Payment gateway adapter (537 lines)
- `internal/domain/payment/` - Payment entities and business logic
- `internal/usecase/paymentops/` - Payment use cases (18 files)

### Configuration & Tools
- `Makefile` - All common commands (30+ targets)
- `.golangci.yml` - Linter configuration (25+ linters)
- `migrations/postgres/` - Database schema changes
- `api/openapi/` - Generated Swagger documentation

### Documentation (.claude/ directory)
- `.claude/README.md` - Quick start guide (30-second overview)
- `.claude/context-guide.md` - Task-specific reading lists
- `.claude/ANALYSIS-2025-10-07.md` - **LATEST fresh codebase analysis** ⭐
- `.claude/architecture.md` - Detailed architecture guide
- `.claude/common-tasks.md` - Step-by-step development workflows
- `.claude/testing.md` - Testing patterns and strategies
- `.claude/REFACTORING-EXECUTIVE-SUMMARY.md` - Complete refactoring overview (Phases 1-8)
- `.claude/REFACTORING-PHASE-8.md` - **LATEST consistency improvements** ⭐
- `.claude/REFACTORING-ROADMAP.md` - Implementation roadmap
- `.claude/QUICK-WINS.md` - Actionable improvements (<30 min each)

## Quick Reference

```bash
# Start coding (first time - RECOMMENDED)
./scripts/dev-setup.sh  # Automated setup: deps, docker, migrations, seeds, hooks

# OR manual setup (first time)
make init && make up && make migrate-up
make install-hooks      # Install pre-commit quality checks

# Daily development
make dev                # Start everything
make run                # Run API server only

# Before commit (pre-commit hooks run automatically)
make ci                 # Run full CI pipeline locally
make test               # Run tests
make lint               # Run linters

# Development data
./scripts/seed-data.sh  # Seed test users and books
# Test accounts: admin@library.com / Admin123!@#
#                user@library.com / User123!@#

# Add new feature (follow this order)
# 1. Domain (entity + service + tests)       → internal/domain/{entity}/
# 2. Use case (orchestration + tests)        → internal/usecase/{entity}ops/  (note "ops" suffix!)
# 3. Adapter (HTTP handler + repository)     → internal/adapters/
# 4. Add Swagger annotations to handlers     → @Security, @Summary, @Param, etc.
# 5. Wire in container.go                    → internal/usecase/container.go
# 6. Migration (if needed)                   → make migrate-create name=...
# 7. Regenerate API docs                     → make gen-docs
```

## Troubleshooting

**"connection refused" errors:**
```bash
make up
docker-compose -f deployments/docker/docker-compose.yml ps
```

**Migration errors:**
```bash
# Check database connection
psql -h localhost -U library -d library

# Reset database (destructive!)
make migrate-down && make migrate-up
```

**Port 8080 already in use:**
```bash
lsof -ti:8080 | xargs kill -9
```

**Tests fail randomly:**
```bash
go clean -testcache && make test
```

**Build errors after refactoring:**
```bash
# Ensure vendor is up to date
go mod tidy && go mod vendor

# Rebuild everything
make clean && make build
```

**Payment gateway timeout errors:**
```bash
# Check epayment.kz environment variables
env | grep EPAYMENT

# Verify gateway connectivity
curl -X POST $EPAYMENT_BASE_URL/oauth2/token
```

## Refactoring Status & Opportunities

**✅ Completed (October 2025 - Phases 1-6 + Pattern Refactoring):**

**Latest Updates (October 11, 2025):** ⭐
- ✅ **Use Case Pattern Refactoring** - All 34 use cases follow unified Execute(ctx, req) pattern
- ✅ **HTTP Handler Pattern Refactoring** - All 8 handlers follow consistent structure (100% compliance)
- ✅ **Legacy Code Removal** - Removed 904 lines, 5 files (LegacyContainer, .unused files)
- ✅ **Container Migration** - Migrated to grouped Container structure (9 domain groups)
- ✅ **Handler Methods** - All handler methods now private (lowercase), consistent with Go idioms
- ✅ **Validation Standardization** - All handlers use `validator.ValidateStruct()` consistently

**Phase 5 Status: VERIFIED COMPLETE** ⭐ *October 9, 2025*
- ✅ All handlers use container injection (8/8 handlers)
- ✅ All DTOs have conversion helpers (zero manual loops)
- ✅ 100% HTTP status code consistency (all use http.Status* constants)
- ✅ Large files split (payment.go → 3 focused files)

### Phase 3: Structural Improvements
- ✅ **Generic Repository Patterns** (ADR 008)
  - Created 7 reusable helpers: GetByID, List, Delete, Exists, Count, etc.
  - Refactored 3 repositories (author, book, member)
  - **Impact:** ~45 lines saved, projected 150+ across all repositories
  - Tests: 8 test functions, all passing

- ✅ **Payment Gateway Modularization** (ADR 009)
  - Split 546-line monolithic `gateway.go` into 4 focused files
  - Organized by responsibility: core (107 lines), auth (118 lines), payment (348 lines), types (61 lines)
  - **Impact:** Better maintainability, single responsibility principle
  - Tests: 14 tests passing, surfaced and fixed 5 hidden bugs

- ✅ **Domain Service for Payment Status** (ADR 010)
  - Extracted payment status logic from use case to domain layer
  - Added 3 methods: MapGatewayStatus(), IsFinalStatus(), UpdateStatusFromCallback()
  - **Impact:** Clean Architecture compliance restored, business logic in domain
  - Tests: All use case tests passing + new domain service tests

- ✅ **BaseRepository Pattern** (ADR 011)
  - Created embeddable BaseRepository[T] with 10 methods
  - Provides: CRUD operations, transactions, utilities (GenerateID, Exists, Count, BatchGet)
  - **Impact:** 86% code reduction for standard operations
  - Tests: 9 comprehensive test functions

**Phase 3 Total Impact:**
- ✅ 4 Architecture Decision Records created
- ✅ ~60 lines removed, ~200 new generic/base code (net positive for maintainability)
- ✅ 40+ tests passing (generic helpers + base repository + domain service + use cases)
- ✅ Clean Architecture compliance improved
- ✅ Foundation laid for rapid repository development

**Prior Work (Phases 1-2):**
- ✅ Package documentation (14 doc.go files)
- ✅ String utilities (`pkg/strutil`) - SafeString, SafeStringPtr
- ✅ HTTP utilities (`pkg/httputil`) - IsServerError, IsClientError, etc.
- ✅ Logger utilities (`pkg/logutil`) - UseCaseLogger, HandlerLogger, etc.
- ✅ Base handler for shared response methods
- ✅ ISBN and timeout constants
- ✅ Critical test coverage (JWT, Password, Payment gateway, Domain services)

**✅ Phase 5 Completion Verified (October 9, 2025):**
- ✅ Handler consistency: 100% (all 8 handlers use container injection)
- ✅ DTO conversion: 100% (zero manual loops, all use helpers)
- ✅ HTTP status codes: 100% (all use http.Status* constants)
- ✅ File organization: DONE (payment.go split into 3 focused files)

**📊 Overall Progress: ~75% Complete (Phases 1-5 Done)**

**Phase 5 Total Impact:**
- ✅ 410 lines eliminated
- ✅ 280+ tests added (all passing)
- ✅ 11 ADRs documenting decisions
- ✅ 100% handler consistency

**Remaining Optional Work (Phases 6-8):**

**🟡 Medium Priority (~20 hours):**
- ⏭️ Migrate remaining 7 repositories to generic patterns (projected: +105 lines saved)
- ⏭️ Add test infrastructure (fixtures, integration tests)
- ⏭️ Logger adoption in final ~15% of locations

**🟢 Low Priority (~3 hours):**
- ⏭️ Final polish (Content-Type constants, etc.)
- ⏭️ Documentation examples for complex domains

**Note:** Remaining work is **optional optimization**, not critical fixes. The codebase is in excellent condition for ongoing development.

**Latest Documentation (October 2025):**
- **Pattern Refactoring:** `.claude/HANDLER_REFACTORING_FINAL.md` ⭐ **100% handler pattern compliance**
- **Use Case Refactoring:** `.claude/COMPLETE_USECASE_REFACTORING.md` ⭐ **All 34 use cases unified**
- **Legacy Removal:** `.claude/LEGACY_CODE_REMOVAL.md` ⭐ **904 lines removed, zero legacy code**
- **Status Report:** `.claude/REFACTORING-STATUS-2025-10-09.md` (comprehensive Phase 5 status)
- **ADRs:** `.claude/adrs/008-011` (Generic Repos, Gateway Split, Domain Services, Base Repo)
- **Migration Guide:** `.claude/MIGRATION-GUIDE-REPOSITORIES.md` (step-by-step repository patterns)

## Pre-approved Commands

These commands are safe to run without asking:
- `make test`, `make test-unit`, `make test-coverage`
- `make fmt`, `make vet`, `make lint`, `make ci`
- `go test ./internal/domain/...`
- `go run cmd/api/main.go` (local development)
- `make gen-docs` (regenerate Swagger documentation)
