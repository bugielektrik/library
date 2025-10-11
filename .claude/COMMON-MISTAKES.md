# Common Mistakes and How to Avoid Them

This guide documents common mistakes made when working with this codebase, why they happen, and how to fix them. Designed for both human developers and AI assistants to prevent repeated errors.

---

## Table of Contents

1. [Architecture Violations](#architecture-violations)
2. [Dependency Injection Errors](#dependency-injection-errors)
3. [Repository Pattern Mistakes](#repository-pattern-mistakes)
4. [Use Case Anti-Patterns](#use-case-anti-patterns)
5. [HTTP Handler Issues](#http-handler-issues)
6. [Testing Mistakes](#testing-mistakes)
7. [Database and Migration Errors](#database-and-migration-errors)
8. [Error Handling Problems](#error-handling-problems)
9. [Validation Issues](#validation-issues)
10. [Logging and Context Problems](#logging-and-context-problems)

---

## Architecture Violations

### Mistake 1: Importing from Outer Layers in Domain

**Wrong:**
```go
// internal/domain/book/service.go
package book

import (
    "library-service/internal/adapters/repository/postgres" // ❌ WRONG!
    "library-service/internal/usecase/bookops"             // ❌ WRONG!
)

func (s *Service) ValidateISBN(isbn string) error {
    // Domain importing from adapters/use cases = architecture violation
}
```

**Why it's wrong:**
- Violates Clean Architecture dependency rule: Domain → Use Case → Adapters
- Domain layer must have ZERO external dependencies
- Makes domain logic untestable and tightly coupled

**Correct:**
```go
// internal/domain/book/service.go
package book

import (
    "context"
    "fmt"
    "regexp"
    // ONLY standard library and internal domain packages allowed
)

func (s *Service) ValidateISBN(isbn string) error {
    // Pure business logic with no external dependencies
    if !isValidISBN13(isbn) {
        return fmt.Errorf("invalid ISBN-13 format")
    }
    return nil
}
```

**Detection:**
- Build will fail with "import cycle not allowed"
- Run `go list -deps ./internal/domain/...` - should show NO internal/usecase or internal/adapters

---

### Mistake 2: Business Logic in Use Cases Instead of Domain Services

**Wrong:**
```go
// internal/usecase/bookops/create_book.go
func (uc *CreateBookUseCase) Execute(ctx context.Context, req CreateBookRequest) error {
    // ❌ Business logic in use case!
    if len(req.ISBN) != 13 {
        return errors.ErrValidation.WithDetails("reason", "ISBN must be 13 digits")
    }

    checksum := calculateISBN13Checksum(req.ISBN)
    if checksum != req.ISBN[12] {
        return errors.ErrValidation.WithDetails("reason", "invalid ISBN checksum")
    }

    // ... persistence logic
}
```

**Why it's wrong:**
- Business rules scattered across multiple use cases
- Impossible to reuse validation logic
- Use cases should orchestrate, NOT implement business rules

**Correct:**
```go
// internal/domain/book/service.go
func (s *Service) ValidateISBN(isbn string) error {
    // ✅ Business logic in domain service
    if len(isbn) != 13 {
        return fmt.Errorf("ISBN must be 13 digits")
    }

    if !s.isValidChecksum(isbn) {
        return fmt.Errorf("invalid ISBN checksum")
    }

    return nil
}

// internal/usecase/bookops/create_book.go
func (uc *CreateBookUseCase) Execute(ctx context.Context, req CreateBookRequest) error {
    // ✅ Use case orchestrates, calls domain service for validation
    if err := uc.bookService.ValidateISBN(req.ISBN); err != nil {
        return errors.ErrValidation.WithDetails("reason", err.Error())
    }

    // Orchestration: call repository, cache, etc.
    return uc.repo.Create(ctx, book)
}
```

**Rule of thumb:**
- If logic is a BUSINESS RULE → Domain Service
- If logic is ORCHESTRATION (calling multiple services/repos) → Use Case

---

## Dependency Injection Errors

### Mistake 3: Forgetting to Wire New Use Cases in Container

**Wrong:**
```go
// Created new use case file, but forgot to add to container.go
// internal/usecase/bookops/archive_book.go
type ArchiveBookUseCase struct {
    repo book.Repository
}

// ❌ Forgot to add to container.go!
```

**Symptom:**
- `nil pointer dereference` when calling use case from handler
- Handler tries to call `useCases.ArchiveBook.Execute()` but ArchiveBook is nil

**Correct:**
```go
// 1. Add to Container struct (internal/usecase/container.go)
type Container struct {
    // ... existing use cases
    ArchiveBook *bookops.ArchiveBookUseCase  // ✅ Add this
}

// 2. Wire in NewContainer() function
func NewContainer(repos *Repositories, ...) *Container {
    bookService := book.NewService()

    return &Container{
        // ... existing use cases
        ArchiveBook: bookops.NewArchiveBookUseCase(repos.Book, bookService),  // ✅ Wire it
    }
}

// 3. Update handler to use it
func (h *BookHandler) ArchiveBook(w http.ResponseWriter, r *http.Request) {
    err := h.useCases.ArchiveBook.Execute(r.Context(), req)  // ✅ Now works
}
```

**Checklist when adding new use case:**
1. ✅ Create use case file in `internal/usecase/{entity}ops/`
2. ✅ Add to `Container` struct in `container.go`
3. ✅ Wire in `NewContainer()` function
4. ✅ Use in HTTP handler

---

### Mistake 4: Creating Infrastructure Services in NewContainer Instead of app.go

**Wrong:**
```go
// internal/usecase/container.go
func NewContainer(repos *Repositories, ...) *Container {
    // ❌ Creating infrastructure services here!
    jwtService := auth.NewJWTService(config.JWTSecret, config.JWTExpiry)
    passwordService := auth.NewPasswordService()

    // This is wrong because:
    // 1. Requires config access (infrastructure concern)
    // 2. Creates multiple instances (should be singleton)
    // 3. Violates separation of concerns
}
```

**Correct:**
```go
// internal/infrastructure/app/app.go
func New() (*App, error) {
    // ✅ Infrastructure services created during app bootstrap
    jwtService := auth.NewJWTService(cfg.JWT.Secret, cfg.JWT.Expiry)
    passwordService := auth.NewPasswordService()

    authSvcs := &usecase.AuthServices{
        JWTService:      jwtService,
        PasswordService: passwordService,
    }

    // ✅ Pass to container
    useCases := usecase.NewContainer(repos, caches, authSvcs, gatewaySvcs)
}

// internal/usecase/container.go
func NewContainer(repos *Repositories, caches *Caches, authSvcs *AuthServices, ...) *Container {
    // ✅ Domain services created here (lightweight, no config)
    bookService := book.NewService()
    memberService := member.NewService()

    return &Container{
        // ✅ Use infrastructure services from authSvcs
        RegisterMember: authops.NewRegisterUseCase(repos.Member, authSvcs.PasswordService, authSvcs.JWTService, memberService),
    }
}
```

**Rule:**
- **Infrastructure services** (JWT, Password, PaymentGateway) → Created in `app.go`
- **Domain services** (Book, Member, Reservation, Payment) → Created in `container.go`

---

## Repository Pattern Mistakes

### Mistake 5: Returning DTOs from Repositories Instead of Domain Entities

**Wrong:**
```go
// internal/adapters/repository/postgres/book.go
func (r *PostgresBookRepository) Get(ctx context.Context, id string) (*dto.BookResponse, error) {
    var book dto.BookResponse  // ❌ Repository returning DTO!
    err := r.db.GetContext(ctx, &book, "SELECT * FROM books WHERE id = $1", id)
    return &book, err
}
```

**Why it's wrong:**
- Repository is an adapter implementing domain interface
- Domain interface specifies domain entities, NOT DTOs
- DTOs are HTTP layer concern, repository shouldn't know about them

**Correct:**
```go
// internal/domain/book/repository.go
type Repository interface {
    Get(ctx context.Context, id string) (*Book, error)  // ✅ Domain entity
}

// internal/adapters/repository/postgres/book.go
func (r *PostgresBookRepository) Get(ctx context.Context, id string) (*book.Book, error) {
    var b book.Book  // ✅ Domain entity
    err := r.db.GetContext(ctx, &b, "SELECT * FROM books WHERE id = $1", id)
    return &b, err
}

// internal/adapters/http/handlers/book.go
func (h *BookHandler) GetBook(w http.ResponseWriter, r *http.Request) {
    book, err := h.useCases.GetBook.Execute(ctx, id)  // Returns domain entity
    if err != nil {
        h.RespondError(w, r, err)
        return
    }

    response := dto.BookFromDomain(book)  // ✅ Convert to DTO in handler
    h.RespondJSON(w, http.StatusOK, response)
}
```

**Data flow:**
```
Repository → Domain Entity → Use Case → Domain Entity → Handler → DTO → JSON
```

---

### Mistake 6: Not Using BaseRepository for Common Operations

**Wrong:**
```go
// internal/adapters/repository/postgres/book.go
type PostgresBookRepository struct {
    db *sqlx.DB
}

func (r *PostgresBookRepository) Create(ctx context.Context, book book.Book) (string, error) {
    // ❌ Reimplementing common CRUD logic!
    query := `INSERT INTO books (id, name, genre, isbn) VALUES ($1, $2, $3, $4) RETURNING id`
    var id string
    err := r.db.GetContext(ctx, &id, query, book.ID, book.Name, book.Genre, book.ISBN)
    return id, err
}

func (r *PostgresBookRepository) Get(ctx context.Context, id string) (*book.Book, error) {
    // ❌ More boilerplate!
    var b book.Book
    query := `SELECT * FROM books WHERE id = $1`
    err := r.db.GetContext(ctx, &b, query, id)
    return &b, err
}
// ... repeating for every repository
```

**Correct:**
```go
// internal/adapters/repository/postgres/book.go
type PostgresBookRepository struct {
    *BaseRepository[book.Book]  // ✅ Embed BaseRepository
}

func NewPostgresBookRepository(db *sqlx.DB) *PostgresBookRepository {
    return &PostgresBookRepository{
        BaseRepository: NewBaseRepository[book.Book](db, "books"),  // ✅ Generic CRUD
    }
}

// ✅ Get Create(), Get(), Update(), Delete() for free from BaseRepository!

// Only implement custom methods
func (r *PostgresBookRepository) ListByGenre(ctx context.Context, genre string) ([]book.Book, error) {
    var books []book.Book
    query := `SELECT * FROM books WHERE genre = $1`
    err := r.db.SelectContext(ctx, &books, query, genre)
    return books, err
}
```

**Benefits:**
- DRY: CRUD logic written once, reused everywhere
- Type-safe with generics
- Consistent error handling

---

## Use Case Anti-Patterns

### Mistake 7: Use Cases Calling Other Use Cases Directly

**Wrong:**
```go
// internal/usecase/reservationops/create_reservation.go
type CreateReservationUseCase struct {
    repo           reservation.Repository
    getMemberUC    *memberops.GetMemberProfileUseCase  // ❌ Use case depending on use case!
}

func (uc *CreateReservationUseCase) Execute(ctx context.Context, req CreateReservationRequest) error {
    // ❌ Use case calling another use case
    member, err := uc.getMemberUC.Execute(ctx, req.MemberID)
    if err != nil {
        return err
    }
    // ...
}
```

**Why it's wrong:**
- Creates tight coupling between use cases
- Makes testing harder (must mock use cases)
- Violates Single Responsibility Principle

**Correct (Option 1 - Call repository directly):**
```go
// internal/usecase/reservationops/create_reservation.go
type CreateReservationUseCase struct {
    reservationRepo reservation.Repository
    memberRepo      member.Repository       // ✅ Depend on repositories
    service         *reservation.Service
}

func (uc *CreateReservationUseCase) Execute(ctx context.Context, req CreateReservationRequest) error {
    // ✅ Call repository directly
    member, err := uc.memberRepo.Get(ctx, req.MemberID)
    if err != nil {
        return fmt.Errorf("getting member: %w", err)
    }

    // Use domain service for business logic
    if err := uc.service.ValidateReservation(member, req.BookID); err != nil {
        return err
    }

    // ... create reservation
}
```

**Correct (Option 2 - Extract to domain service):**
```go
// internal/domain/reservation/service.go
func (s *Service) CanMemberReserveBook(ctx context.Context, memberRepo member.Repository, memberID, bookID string) error {
    // ✅ Domain service orchestrates domain logic
    member, err := memberRepo.Get(ctx, memberID)
    if err != nil {
        return fmt.Errorf("getting member: %w", err)
    }

    if !member.HasActiveSubscription() {
        return errors.ErrValidation.WithDetails("reason", "member has no active subscription")
    }

    return nil
}
```

**Exception:** ProcessCallbackRetries depends on HandleCallback (line 166 in container.go). This is acceptable because:
- ProcessCallbackRetries is a worker job, HandleCallback is the core logic
- HandleCallback is reusable (webhook calls + manual retries)
- Created intentionally in specific order to handle dependency

---

### Mistake 8: Not Logging Use Case Execution

**Wrong:**
```go
// internal/usecase/bookops/create_book.go
func (uc *CreateBookUseCase) Execute(ctx context.Context, req CreateBookRequest) error {
    // ❌ No logging!
    book := book.New(req.Name, req.Genre, req.ISBN)

    if err := uc.repo.Create(ctx, book); err != nil {
        return err  // ❌ Error disappears without trace
    }

    return nil
}
```

**Why it's wrong:**
- Impossible to debug production issues
- No visibility into execution flow
- Can't track performance or identify bottlenecks

**Correct:**
```go
// internal/usecase/bookops/create_book.go
func (uc *CreateBookUseCase) Execute(ctx context.Context, req CreateBookRequest) error {
    logger := logutil.UseCaseLogger(ctx, "create_book_usecase", "execute")
    logger.Info("creating book",
        zap.String("name", strutil.SafeString(req.Name)),
        zap.String("isbn", strutil.SafeString(req.ISBN)),
    )

    book := book.New(req.Name, req.Genre, req.ISBN)

    if err := uc.bookService.ValidateISBN(strutil.SafeString(req.ISBN)); err != nil {
        logger.Error("ISBN validation failed", zap.Error(err))
        return errors.ErrValidation.WithDetails("reason", err.Error())
    }

    if err := uc.repo.Create(ctx, book); err != nil {
        logger.Error("failed to create book in repository", zap.Error(err))
        return fmt.Errorf("creating book: %w", err)
    }

    logger.Info("book created successfully", zap.String("book_id", book.ID))
    return nil
}
```

**Pattern:**
1. Create logger at start: `logutil.UseCaseLogger(ctx, "usecase_name", "method_name")`
2. Log start with input parameters
3. Log errors with context
4. Log success with result identifiers

---

## HTTP Handler Issues

### Mistake 9: Not Using Validator for Input Validation

**Wrong:**
```go
// internal/adapters/http/handlers/book.go
func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
    var req bookops.CreateBookRequest
    if err := httputil.DecodeJSON(r, &req); err != nil {
        h.RespondError(w, r, err)
        return
    }

    // ❌ Manual validation instead of validator
    if req.Name == nil || *req.Name == "" {
        h.RespondError(w, r, errors.ErrValidation.WithDetails("reason", "name is required"))
        return
    }

    if req.ISBN == nil || *req.ISBN == "" {
        h.RespondError(w, r, errors.ErrValidation.WithDetails("reason", "ISBN is required"))
        return
    }

    // ... more manual checks
}
```

**Correct:**
```go
// Define validation tags on DTO
type CreateBookRequest struct {
    Name  *string `json:"name" validate:"required,min=1,max=200"`
    Genre *string `json:"genre" validate:"required"`
    ISBN  *string `json:"isbn" validate:"required,len=13"`
}

// Use validator in handler
func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
    var req bookops.CreateBookRequest
    if err := httputil.DecodeJSON(r, &req); err != nil {
        h.RespondError(w, r, err)
        return
    }

    // ✅ Validator handles all validation
    if !h.validator.ValidateStruct(w, req) {
        return  // Validator already responded with error
    }

    response, err := h.useCases.CreateBook.Execute(r.Context(), req)
    // ...
}
```

**All handlers must:**
1. Have `validator *middleware.Validator` field
2. Accept validator in constructor
3. Call `validator.ValidateStruct(w, req)` after decoding JSON

---

### Mistake 10: Using String Literals for Content-Type Headers

**Wrong:**
```go
// internal/adapters/http/handlers/payment_page.go
func (h *PaymentPageHandler) ServePaymentPage(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html; charset=utf-8")  // ❌ String literal
    // ...
}

func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")  // ❌ Inconsistent, missing charset
    // ...
}
```

**Correct:**
```go
// internal/adapters/http/handlers/payment_page.go
import "library-service/pkg/httputil"

func (h *PaymentPageHandler) ServePaymentPage(w http.ResponseWriter, r *http.Request) {
    w.Header().Set(httputil.HeaderContentType, httputil.ContentTypeHTML)  // ✅ Constants
    // ...
}

// internal/adapters/http/handlers/base.go (or use httputil.RespondJSON)
func (h *BaseHandler) RespondJSON(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set(httputil.HeaderContentType, httputil.ContentTypeJSON)  // ✅ Constants
    w.WriteStatus(status)
    json.NewEncoder(w).Encode(data)
}
```

**Available constants:**
- `httputil.HeaderContentType` = "Content-Type"
- `httputil.ContentTypeJSON` = "application/json; charset=utf-8"
- `httputil.ContentTypeHTML` = "text/html; charset=utf-8"

---

## Testing Mistakes

### Mistake 11: Mocking Domain Services in Tests

**Wrong:**
```go
func TestCreateBook(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockRepo := mocks.NewMockBookRepository(ctrl)
    mockCache := mocks.NewMockBookCache(ctrl)
    mockService := mocks.NewMockBookService(ctrl)  // ❌ Mocking domain service!

    mockService.EXPECT().ValidateISBN(gomock.Any()).Return(nil)  // ❌ Testing mock, not real logic!

    useCase := bookops.NewCreateBookUseCase(mockRepo, mockCache, mockService)
    // ...
}
```

**Why it's wrong:**
- Domain services are pure business logic with no external dependencies
- Mocking them means you're NOT testing the business rules
- Makes tests fragile and meaningless

**Correct:**
```go
func TestCreateBook(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockRepo := mocks.NewMockBookRepository(ctrl)
    mockCache := mocks.NewMockBookCache(ctrl)
    realService := book.NewService()  // ✅ Use real domain service!

    mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return("book-id", nil)
    mockCache.EXPECT().Invalidate(gomock.Any(), "book-id").Return(nil)

    useCase := bookops.NewCreateBookUseCase(mockRepo, mockCache, realService)

    // ✅ This actually tests ISBN validation logic
    err := useCase.Execute(ctx, bookops.CreateBookRequest{
        ISBN: strutil.SafeStringPtr("978-invalid"),  // Should fail validation
    })

    assert.Error(t, err)
}
```

**Rule:**
- **Mock:** Repositories, Caches, External Services (JWT, PaymentGateway)
- **Real:** Domain Services (pure business logic)

---

### Mistake 12: Not Using Build Tags for Integration Tests

**Wrong:**
```go
// test/integration/book_repository_test.go
package integration

// ❌ No build tag!

func TestBookRepository_Create(t *testing.T) {
    // Integration test requiring real PostgreSQL
}
```

**Why it's wrong:**
- Integration tests run with `go test ./...` even when DB is unavailable
- CI/CD fails if database not ready
- Slows down unit test execution

**Correct:**
```go
// test/integration/book_repository_test.go
//go:build integration  // ✅ Build tag

package integration

func TestBookRepository_Create(t *testing.T) {
    // Integration test requiring real PostgreSQL
}
```

**Usage:**
```bash
# Run unit tests only (fast, no database required)
go test ./...

# Run integration tests only (requires PostgreSQL)
go test -tags=integration ./test/integration/...

# Run all tests
go test -tags=integration ./...
```

---

## Database and Migration Errors

### Mistake 13: Not Creating Both .up.sql and .down.sql Migrations

**Wrong:**
```bash
# Only created up migration
$ ls migrations/postgres/
000007_add_payments_table.up.sql  # ✅
# ❌ Missing .down.sql
```

**Why it's wrong:**
- Can't rollback migration if something goes wrong
- Breaks `make migrate-down` command
- CI/CD fails on rollback tests

**Correct:**
```bash
$ make migrate-create name=add_payments_table

# Creates both:
migrations/postgres/000007_add_payments_table.up.sql    # ✅
migrations/postgres/000007_add_payments_table.down.sql  # ✅
```

**Migration file structure:**
```sql
-- 000007_add_payments_table.up.sql
CREATE TABLE payments (
    id UUID PRIMARY KEY,
    amount INTEGER NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 000007_add_payments_table.down.sql
DROP TABLE IF EXISTS payments;
```

---

### Mistake 14: Forgetting to Run Migrations After Schema Changes

**Wrong:**
```bash
# Made changes to migrations/postgres/000007_add_payments_table.up.sql
# Ran the server immediately without applying migration

$ make run
# ❌ Error: relation "payments" does not exist
```

**Correct:**
```bash
# After creating or modifying migrations, ALWAYS run:
$ make migrate-up

# Verify migration applied:
$ psql -h localhost -U library -d library -c "\dt"
# Should show payments table

# Then run server:
$ make run
```

**Checklist for schema changes:**
1. ✅ Create migration: `make migrate-create name=...`
2. ✅ Write both .up.sql and .down.sql
3. ✅ Apply migration: `make migrate-up`
4. ✅ Verify in database: `psql ...` or `make db-cli`
5. ✅ Test rollback: `make migrate-down && make migrate-up`
6. ✅ Update domain entities and repositories if needed

---

## Error Handling Problems

### Mistake 15: Swallowing Errors Without Context

**Wrong:**
```go
func (uc *CreateBookUseCase) Execute(ctx context.Context, req CreateBookRequest) error {
    book := book.New(req.Name, req.Genre, req.ISBN)

    if err := uc.repo.Create(ctx, book); err != nil {
        return err  // ❌ No context about WHAT failed or WHERE
    }

    if err := uc.cache.Invalidate(ctx, book.ID); err != nil {
        return err  // ❌ Same error type, impossible to distinguish
    }

    return nil
}
```

**Why it's wrong:**
- Error "database connection timeout" could come from repo OR cache
- No way to know which operation failed
- Debugging production issues becomes nightmare

**Correct:**
```go
func (uc *CreateBookUseCase) Execute(ctx context.Context, req CreateBookRequest) error {
    logger := logutil.UseCaseLogger(ctx, "create_book_usecase", "execute")

    book := book.New(req.Name, req.Genre, req.ISBN)

    if err := uc.repo.Create(ctx, book); err != nil {
        logger.Error("failed to create book in repository", zap.Error(err))
        return fmt.Errorf("creating book in repository: %w", err)  // ✅ Context + wrapped error
    }

    if err := uc.cache.Invalidate(ctx, book.ID); err != nil {
        logger.Warn("failed to invalidate cache", zap.Error(err))  // ✅ Log but don't fail
        // Cache failure is not critical, continue
    }

    logger.Info("book created successfully", zap.String("book_id", book.ID))
    return nil
}
```

**Error wrapping pattern:**
```go
fmt.Errorf("descriptive context: %w", originalError)
```

**Benefits:**
- Full error chain preserved
- Can unwrap with `errors.Is()` and `errors.As()`
- Clear error messages in logs

---

### Mistake 16: Not Using Domain Errors for Common Cases

**Wrong:**
```go
func (r *PostgresBookRepository) Get(ctx context.Context, id string) (*book.Book, error) {
    var b book.Book
    err := r.db.GetContext(ctx, &b, "SELECT * FROM books WHERE id = $1", id)
    if err != nil {
        return nil, err  // ❌ Returning raw sql.ErrNoRows to caller
    }
    return &b, nil
}

// Handler has to check for sql.ErrNoRows
if err == sql.ErrNoRows {  // ❌ Coupling handler to database errors
    respondError(w, r, 404)
}
```

**Correct:**
```go
func (r *PostgresBookRepository) Get(ctx context.Context, id string) (*book.Book, error) {
    var b book.Book
    err := r.db.GetContext(ctx, &b, "SELECT * FROM books WHERE id = $1", id)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, errors.ErrNotFound  // ✅ Convert to domain error
        }
        return nil, fmt.Errorf("querying book: %w", err)
    }
    return &b, nil
}

// Handler checks domain errors (no database knowledge needed)
if errors.Is(err, errors.ErrNotFound) {  // ✅ Domain error
    h.RespondError(w, r, err)  // ErrorHandler middleware converts to 404
}
```

**Available domain errors** (`pkg/errors/domain.go`):
- `ErrNotFound` → 404
- `ErrAlreadyExists` → 409
- `ErrValidation` → 400
- `ErrUnauthorized` → 401
- `ErrForbidden` → 403
- `ErrConflict` → 409

---

## Validation Issues

### Mistake 17: Validating Business Rules in HTTP Layer

**Wrong:**
```go
// internal/adapters/http/handlers/reservation.go
func (h *ReservationHandler) CreateReservation(w http.ResponseWriter, r *http.Request) {
    var req reservationops.CreateReservationRequest
    if err := httputil.DecodeJSON(r, &req); err != nil {
        h.RespondError(w, r, err)
        return
    }

    // ❌ Business rule validation in HTTP handler!
    if req.ReservationDuration > 30 {
        h.RespondError(w, r, errors.ErrValidation.WithDetails("reason", "max reservation duration is 30 days"))
        return
    }

    response, err := h.useCases.CreateReservation.Execute(r.Context(), req)
    // ...
}
```

**Why it's wrong:**
- Business rules should be in domain layer, not HTTP layer
- If you add a gRPC/GraphQL API later, you have to duplicate validation
- Makes business rules scattered and hard to maintain

**Correct:**
```go
// internal/domain/reservation/service.go
const MaxReservationDurationDays = 30

func (s *Service) ValidateReservationDuration(days int) error {
    if days > MaxReservationDurationDays {
        return fmt.Errorf("max reservation duration is %d days", MaxReservationDurationDays)
    }
    if days < 1 {
        return fmt.Errorf("reservation duration must be at least 1 day")
    }
    return nil
}

// internal/usecase/reservationops/create_reservation.go
func (uc *CreateReservationUseCase) Execute(ctx context.Context, req CreateReservationRequest) error {
    // ✅ Business validation in use case via domain service
    if err := uc.service.ValidateReservationDuration(req.ReservationDuration); err != nil {
        return errors.ErrValidation.WithDetails("reason", err.Error())
    }

    // ... persistence
}

// internal/adapters/http/handlers/reservation.go
func (h *ReservationHandler) CreateReservation(w http.ResponseWriter, r *http.Request) {
    var req reservationops.CreateReservationRequest
    if err := httputil.DecodeJSON(r, &req); err != nil {
        h.RespondError(w, r, err)
        return
    }

    // ✅ Only structural validation (required fields, types)
    if !h.validator.ValidateStruct(w, req) {
        return
    }

    // ✅ Business validation happens in use case
    response, err := h.useCases.CreateReservation.Execute(r.Context(), req)
    // ...
}
```

**Validation layer responsibilities:**
- **HTTP Handler (validator):** Structural validation (required, min/max length, format)
- **Domain Service:** Business rule validation (ISBN checksum, subscription pricing, reservation duration)

---

## Logging and Context Problems

### Mistake 18: Not Propagating Context Through Layers

**Wrong:**
```go
// Handler
func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
    // ❌ Using background context instead of request context
    err := h.useCases.CreateBook.Execute(context.Background(), req)
}

// Use case
func (uc *CreateBookUseCase) Execute(ctx context.Context, req CreateBookRequest) error {
    // ❌ Not passing context to repository
    return uc.repo.Create(context.Background(), book)
}
```

**Why it's wrong:**
- Request cancellation not propagated (client disconnects but query continues)
- Request ID lost (can't trace request through logs)
- Timeout not enforced
- Breaks distributed tracing

**Correct:**
```go
// Handler
func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
    // ✅ Use request context
    err := h.useCases.CreateBook.Execute(r.Context(), req)
}

// Use case
func (uc *CreateBookUseCase) Execute(ctx context.Context, req CreateBookRequest) error {
    logger := logutil.UseCaseLogger(ctx, "create_book_usecase", "execute")  // ✅ Context contains request ID

    // ✅ Pass context to all downstream calls
    if err := uc.repo.Create(ctx, book); err != nil {
        return fmt.Errorf("creating book: %w", err)
    }

    if err := uc.cache.Invalidate(ctx, book.ID); err != nil {
        logger.Warn("cache invalidation failed", zap.Error(err))
    }

    return nil
}

// Repository
func (r *PostgresBookRepository) Create(ctx context.Context, book book.Book) (string, error) {
    // ✅ Context passed to database query
    err := r.db.GetContext(ctx, &id, query, ...)
    return id, err
}
```

**Context rules:**
1. Always accept `context.Context` as FIRST parameter
2. Always pass context to downstream calls
3. NEVER use `context.Background()` in business logic (only in main, tests)
4. Use `r.Context()` in HTTP handlers

---

### Mistake 19: Not Using logutil Helpers

**Wrong:**
```go
// Handler
func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
    logger := h.logger.Named("book_handler").Named("create_book")  // ❌ Manual logger creation
    logger.Info("creating book")
    // ...
}

// Use case
func (uc *CreateBookUseCase) Execute(ctx context.Context, req CreateBookRequest) error {
    logger := zap.L().Named("create_book_usecase")  // ❌ Lost request context!
    logger.Info("executing use case")
    // ...
}
```

**Why it's wrong:**
- Inconsistent logger naming
- Lost request ID (can't trace logs for single request)
- Manual logger creation is error-prone

**Correct:**
```go
// Handler
func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    logger := logutil.HandlerLogger(ctx, "book_handler", "create_book")  // ✅ Extracts request ID from context
    logger.Info("creating book", zap.String("isbn", req.ISBN))
    // ...
}

// Use case
func (uc *CreateBookUseCase) Execute(ctx context.Context, req CreateBookRequest) error {
    logger := logutil.UseCaseLogger(ctx, "create_book_usecase", "execute")  // ✅ Request ID preserved
    logger.Info("executing use case", zap.String("book_id", book.ID))
    // ...
}
```

**logutil helpers:**
- `logutil.HandlerLogger(ctx, "handler_name", "method_name")` - For HTTP handlers
- `logutil.UseCaseLogger(ctx, "usecase_name", "method_name")` - For use cases

**Benefits:**
- Automatic request ID injection
- Consistent naming convention
- All logs for single request can be traced

---

## Prevention Strategies

### Before Committing

Run full CI pipeline locally:
```bash
make ci  # fmt → vet → lint → test → build
```

This catches:
- Import cycle errors
- Unused variables
- Missing error checks
- Test failures

### Adding New Features

Follow this checklist (from `.claude/development-workflows.md`):

1. ✅ Domain layer (entity + service + repository interface + tests)
2. ✅ Repository implementation (PostgreSQL)
3. ✅ Use case (orchestration + tests)
4. ✅ HTTP handler (DTO + validation + Swagger)
5. ✅ Wire in `container.go` (add to Container struct + NewContainer function)
6. ✅ Add routes in `router.go`
7. ✅ Migration if needed (`make migrate-create`)
8. ✅ Run `make gen-docs` to update Swagger
9. ✅ Run `make test` to verify everything works
10. ✅ Test manually with `curl` or Swagger UI

### Code Review Checklist

When reviewing PRs or your own code:

- [ ] Domain layer has no external dependencies
- [ ] Business logic in domain services, NOT use cases
- [ ] All use cases wired in `container.go`
- [ ] Repositories return domain entities, NOT DTOs
- [ ] HTTP handlers use validator for input validation
- [ ] Content-Type constants used (no string literals)
- [ ] Context propagated through all layers
- [ ] Errors wrapped with context
- [ ] Domain errors used (ErrNotFound, ErrValidation, etc.)
- [ ] Logging with logutil helpers
- [ ] Integration tests have `//go:build integration` tag
- [ ] Both .up.sql and .down.sql migrations created
- [ ] Swagger annotations added for new endpoints

---

## Summary

**Top 5 Most Common Mistakes:**

1. **Architecture violation:** Domain importing from outer layers → Breaks Clean Architecture
2. **Missing wiring:** Forgot to add use case to `container.go` → Nil pointer panic
3. **Wrong error handling:** Not wrapping errors with context → Impossible to debug
4. **No validation:** Missing validator calls in handlers → Bad data reaches database
5. **Lost context:** Using `context.Background()` instead of request context → Can't trace requests

**Quick References:**
- Architecture rules: CLAUDE.md#Architecture
- Container wiring: internal/usecase/container.go (comprehensive guide)
- Domain errors: pkg/errors/domain.go
- Logging helpers: pkg/logutil/
- Validation: internal/adapters/http/middleware/validator.go

For development workflows, see: `.claude/development-workflows.md`
