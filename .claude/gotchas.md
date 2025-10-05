# Common Pitfalls & Gotchas

> **Real mistakes to avoid in this codebase - learn from others' errors!**

## 🚨 Critical Mistakes

### 1. Package Naming Conflicts

**❌ WRONG:**
```go
// Using same package name for domain and use case
// internal/domain/book/entity.go
package book

// internal/usecase/book/create_book.go
package book  // ← CONFLICT!

// When importing both:
import (
    "library-service/internal/domain/book"
    "library-service/internal/usecase/book"  // ← Name collision!
)
```

**✅ CORRECT:**
```go
// internal/usecase/bookops/create_book.go
package bookops  // ← Note "ops" suffix!

// Now you can import both cleanly:
import (
    "library-service/internal/domain/book"
    "library-service/internal/usecase/bookops"
)

book.Entity{}           // Domain entity
bookops.CreateBookUseCase{}  // Use case
```

**Why:** Go doesn't allow two packages with the same name in the same file. Using "ops" suffix avoids import aliases and makes code cleaner.

### 2. Domain Layer Importing from Outer Layers

**❌ WRONG:**
```go
// internal/domain/book/service.go
package book

import (
    "library-service/internal/adapters/repository/postgres"  // ← FORBIDDEN!
)

func (s *Service) SaveBook(b Entity) error {
    repo := postgres.NewBookRepository()  // ← Violates Clean Architecture!
    return repo.Create(b)
}
```

**✅ CORRECT:**
```go
// internal/domain/book/service.go
package book

// Domain service has NO dependencies, only pure logic
func (s *Service) ValidateISBN(isbn string) error {
    // Pure business logic only
    if len(isbn) != 13 {
        return ErrInvalidISBN
    }
    return nil
}

// internal/usecase/bookops/create_book.go
package bookops

type CreateBookUseCase struct {
    repo book.Repository  // ← Interface from domain
}

func (uc *CreateBookUseCase) Execute(req Request) error {
    // Validate using domain service
    if err := uc.bookService.ValidateISBN(req.ISBN); err != nil {
        return err
    }

    // Persist using repository interface
    return uc.repo.Create(ctx, entity)
}
```

**Why:** Domain must be pure. Dependencies must point INWARD (Domain ← Use Case ← Adapters).

### 3. Business Logic in Wrong Layer

**❌ WRONG - Business logic in handler:**
```go
// internal/adapters/http/handlers/book.go
func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
    // ❌ Business logic in handler!
    if len(req.ISBN) != 13 {
        respondError(w, "invalid ISBN", http.StatusBadRequest)
        return
    }

    // ❌ Checksum validation in handler!
    if !validateISBNChecksum(req.ISBN) {
        respondError(w, "invalid checksum", http.StatusBadRequest)
        return
    }
}
```

**❌ WRONG - Business logic in use case:**
```go
// internal/usecase/bookops/create_book.go
func (uc *CreateBookUseCase) Execute(req Request) error {
    // ❌ Business rules in use case!
    if len(req.ISBN) != 13 {
        return errors.New("invalid ISBN")
    }
}
```

**✅ CORRECT - Business logic in domain service:**
```go
// internal/domain/book/service.go
func (s *Service) ValidateISBN(isbn string) error {
    // ✓ Business rule belongs here
    if len(isbn) != 13 {
        return ErrInvalidISBN
    }

    // ✓ Complex validation logic
    if !s.validateChecksum(isbn) {
        return ErrInvalidChecksum
    }

    return nil
}

// internal/usecase/bookops/create_book.go
func (uc *CreateBookUseCase) Execute(req Request) error {
    // ✓ Use case orchestrates
    if err := uc.bookService.ValidateISBN(req.ISBN); err != nil {
        return err
    }
    // ... rest of orchestration
}

// internal/adapters/http/handlers/book.go
func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
    // ✓ Handler just calls use case
    result, err := h.createBookUC.Execute(req)
    if err != nil {
        respondError(w, err, http.StatusInternalServerError)
        return
    }
    respondJSON(w, result, http.StatusCreated)
}
```

**Why:** Business logic in domain = testable, reusable. Business logic in handlers = untestable, scattered.

## 🐛 Common Bugs

### 4. Missing Error Wrapping

**❌ WRONG:**
```go
func (uc *CreateBookUseCase) Execute(req Request) error {
    if err := uc.repo.Create(ctx, book); err != nil {
        return err  // ❌ Lost context!
    }
}
```

**✅ CORRECT:**
```go
func (uc *CreateBookUseCase) Execute(req Request) error {
    if err := uc.repo.Create(ctx, book); err != nil {
        return fmt.Errorf("creating book in repository: %w", err)  // ✓ Context + wrapping
    }
}
```

**Error message chain:**
```
creating book in repository: duplicate key value: books_isbn_key
                                                  ↑
                           wrapped error          original error from DB
```

**Why:** `%w` allows `errors.Is()` and `errors.As()` to work. Adding context helps debugging.

### 5. Forgetting @Security Annotation

**❌ WRONG:**
```go
// @Summary Create a book
// @Tags books
// @Accept json
// @Param request body dto.CreateBookRequest true "Book details"
// @Success 201 {object} dto.BookResponse
// @Router /books [post]
func (h *BookHandler) CreateBook(...) {
    // Handler requires authentication but no @Security!
}
```

**✅ CORRECT:**
```go
// @Summary Create a book
// @Tags books
// @Accept json
// @Security BearerAuth  // ← REQUIRED!
// @Param request body dto.CreateBookRequest true "Book details"
// @Success 201 {object} dto.BookResponse
// @Router /books [post]
func (h *BookHandler) CreateBook(...) {
    // Now Swagger UI knows this needs authentication
}
```

**Why:** Without `@Security BearerAuth`, Swagger UI won't prompt for authentication and requests will fail with 401.

### 6. SQL Injection Risk

**❌ DANGEROUS:**
```go
func (r *Repository) GetByISBN(isbn string) (Entity, error) {
    query := "SELECT * FROM books WHERE isbn = '" + isbn + "'"  // ← SQL INJECTION!
    row := r.db.QueryRow(query)
}
```

**❌ ALSO DANGEROUS:**
```go
query := fmt.Sprintf("SELECT * FROM books WHERE isbn = '%s'", isbn)  // ← Still vulnerable!
```

**✅ CORRECT:**
```go
func (r *Repository) GetByISBN(ctx context.Context, isbn string) (Entity, error) {
    query := "SELECT * FROM books WHERE isbn = $1"  // ← Parameterized!
    row := r.db.QueryRowContext(ctx, query, isbn)
}
```

**Why:** Parameterized queries prevent SQL injection. Always use `$1, $2, $3...` placeholders.

### 7. Missing Migration Down File

**❌ WRONG:**
```bash
# Only create .up.sql
migrations/postgres/000004_add_loans.up.sql  ✓
migrations/postgres/000004_add_loans.down.sql  ✗ Missing!
```

**✅ CORRECT:**
```bash
# Always create both
migrations/postgres/000004_add_loans.up.sql  ✓
migrations/postgres/000004_add_loans.down.sql  ✓
```

**Why:** Down migrations are needed to rollback. CI will fail without them.

## 🧪 Testing Pitfalls

### 8. Using Real Database in Unit Tests

**❌ WRONG:**
```go
// internal/domain/book/service_test.go
func TestValidateISBN(t *testing.T) {
    db := setupTestDB()  // ❌ Database in unit test!
    repo := postgres.NewBookRepository(db)
    svc := NewService(repo)

    err := svc.ValidateISBN("invalid")
}
```

**✅ CORRECT:**
```go
// internal/domain/book/service_test.go
func TestValidateISBN(t *testing.T) {
    svc := NewService()  // ✓ No dependencies!

    err := svc.ValidateISBN("invalid")
    assert.Error(t, err)
}
```

**Why:** Unit tests should be FAST (< 1 second). Database = integration test. Domain services should be pure (no dependencies).

### 9. Not Using Table-Driven Tests

**❌ WRONG:**
```go
func TestValidateISBN_Valid(t *testing.T) {
    err := ValidateISBN("9780132350884")
    assert.NoError(t, err)
}

func TestValidateISBN_Invalid(t *testing.T) {
    err := ValidateISBN("invalid")
    assert.Error(t, err)
}

func TestValidateISBN_Empty(t *testing.T) {
    err := ValidateISBN("")
    assert.Error(t, err)
}
// Lots of duplication!
```

**✅ CORRECT:**
```go
func TestValidateISBN(t *testing.T) {
    tests := []struct {
        name    string
        isbn    string
        wantErr bool
    }{
        {"valid ISBN-13", "9780132350884", false},
        {"invalid format", "invalid", true},
        {"empty string", "", true},
        {"invalid checksum", "9780132350881", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateISBN(tt.isbn)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

**Why:** Table-driven tests reduce duplication and make it easy to add test cases.

### 10. Forgetting to Clean Test Cache

**❌ SYMPTOM:**
```
Tests pass locally but fail in CI
Tests pass first time but fail on second run
```

**✅ FIX:**
```bash
# Clear test cache before running
go clean -testcache
go test ./...
```

**Why:** Go caches test results. Stale cache can cause false positives.

## 🔐 Security Pitfalls

### 11. Hardcoding Secrets

**❌ WRONG:**
```go
// internal/infrastructure/auth/jwt.go
const jwtSecret = "my-super-secret-key"  // ❌ NEVER!

func GenerateToken() string {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(jwtSecret))
}
```

**✅ CORRECT:**
```go
// internal/infrastructure/auth/jwt.go
type JWTService struct {
    secret string  // ✓ From config
}

func NewJWTService(secret string) *JWTService {
    return &JWTService{secret: secret}
}

// cmd/api/main.go
secret := os.Getenv("JWT_SECRET")  // ✓ From environment
if secret == "" {
    log.Fatal("JWT_SECRET is required")
}
jwtService := auth.NewJWTService(secret)
```

**Why:** Secrets in code = security breach. Always use environment variables.

### 12. Not Validating DTO Input

**❌ WRONG:**
```go
func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
    var req dto.CreateBookRequest
    json.NewDecoder(r.Body).Decode(&req)

    // ❌ No validation! Could have empty name, invalid ISBN, etc.
    book, _ := h.createBookUC.Execute(req)
}
```

**✅ CORRECT:**
```go
type CreateBookRequest struct {
    Name   string   `json:"name" validate:"required,max=255"`
    ISBN   string   `json:"isbn" validate:"required,isbn"`
    Genre  string   `json:"genre" validate:"required"`
    Authors []string `json:"authors" validate:"required,min=1"`
}

func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
    var req dto.CreateBookRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, errors.ErrValidation, http.StatusBadRequest)
        return
    }

    // ✓ Validate before processing
    if err := validate.Struct(req); err != nil {
        respondError(w, err, http.StatusBadRequest)
        return
    }

    book, err := h.createBookUC.Execute(req)
    // ...
}
```

**Why:** Always validate input at the boundary. Never trust user input.

## 📊 Performance Pitfalls

### 13. N+1 Query Problem

**❌ WRONG:**
```go
func (uc *ListBooksUseCase) Execute() ([]BookWithAuthors, error) {
    books, _ := uc.repo.ListBooks()

    result := make([]BookWithAuthors, len(books))
    for i, book := range books {
        // ❌ N+1! One query per book!
        authors, _ := uc.authorRepo.GetByBookID(book.ID)
        result[i] = BookWithAuthors{Book: book, Authors: authors}
    }
    return result, nil
}
```

**Query count:** 1 (list books) + N (for each book) = N+1 queries

**✅ CORRECT:**
```go
func (r *Repository) ListBooksWithAuthors() ([]BookWithAuthors, error) {
    query := `
        SELECT b.id, b.name, b.isbn, a.id, a.name
        FROM books b
        LEFT JOIN book_authors ba ON b.id = ba.book_id
        LEFT JOIN authors a ON ba.author_id = a.id
    `  // ✓ Single query with JOIN!

    rows, _ := r.db.Query(query)
    // Map results...
}
```

**Query count:** 1

**Why:** N+1 queries kill performance. Use JOINs or batch loading.

### 14. Missing Database Indexes

**❌ WRONG:**
```sql
CREATE TABLE loans (
    id UUID PRIMARY KEY,
    book_id UUID NOT NULL,      -- ❌ No index!
    member_id UUID NOT NULL,    -- ❌ No index!
    status VARCHAR(20)           -- ❌ No index!
);
```

**Query performance:** Slow! Full table scan for every query.

**✅ CORRECT:**
```sql
CREATE TABLE loans (
    id UUID PRIMARY KEY,
    book_id UUID NOT NULL,
    member_id UUID NOT NULL,
    status VARCHAR(20)
);

-- ✓ Index foreign keys
CREATE INDEX idx_loans_book_id ON loans(book_id);
CREATE INDEX idx_loans_member_id ON loans(member_id);

-- ✓ Index frequently queried columns
CREATE INDEX idx_loans_status ON loans(status);

-- ✓ Composite index for common query patterns
CREATE INDEX idx_loans_member_status ON loans(member_id, status);
```

**Query performance:** Fast! Index lookup instead of table scan.

**Why:** Always index foreign keys and frequently queried columns.

## 🎯 Workflow Pitfalls

### 15. Not Running `make ci` Before Commit

**❌ WRONG:**
```bash
git add .
git commit -m "feat: add loans"
git push
# CI fails! 😞
```

**✅ CORRECT:**
```bash
make ci  # Runs: fmt, vet, lint, test, build
# Wait for it to pass...
git add .
git commit -m "feat: add loans"
git push
# CI passes! 😊
```

**Why:** Catch issues locally before pushing. Save CI time and avoid embarrassment.

### 16. Forgetting to Regenerate Swagger

**❌ SYMPTOM:**
```
Added new endpoint but it doesn't show in Swagger UI
Changed response DTO but Swagger shows old structure
```

**✅ FIX:**
```bash
# After ANY handler changes:
make gen-docs

# Then restart API
make run
```

**Why:** Swagger docs are generated from code annotations. Must regenerate after changes.

### 17. Not Using Correct Git Commit Format

**❌ WRONG:**
```bash
git commit -m "fixed stuff"
git commit -m "WIP"
git commit -m "asdfasdf"
```

**✅ CORRECT:**
```bash
git commit -m "feat: add loan management system"
git commit -m "fix: resolve ISBN validation bug"
git commit -m "refactor: extract validation to domain service"
git commit -m "docs: update API documentation"
git commit -m "test: add integration tests for loans"
```

**Format:** `type: description`
- `feat:` new feature
- `fix:` bug fix
- `refactor:` code refactoring
- `docs:` documentation
- `test:` tests
- `chore:` maintenance

**Why:** Clear history, semantic versioning, automated changelogs.

## 🤔 Conceptual Misunderstandings

### 18. Confusing Domain Service with Use Case

**Common Question:** "Why do I need both BookService and CreateBookUseCase?"

**Answer:**

**Domain Service** = Business Rules (What can/cannot happen)
```go
// internal/domain/book/service.go
func (s *Service) ValidateISBN(isbn string) error {
    // PURE business logic
    // No database, no HTTP, no framework
    // Just: "Is this ISBN valid according to our business rules?"
}

func (s *Service) CanBeDeleted(book Entity) error {
    // Business constraint: "Can we delete this book?"
}
```

**Use Case** = Workflow (How to do something)
```go
// internal/usecase/bookops/create_book.go
func (uc *CreateBookUseCase) Execute(req Request) error {
    // Orchestration:
    // 1. Validate (using domain service)
    // 2. Create entity
    // 3. Save to database
    // 4. Update cache
    // 5. Log event
}
```

**Rule:** If it's a business rule → Domain Service. If it's a workflow → Use Case.

### 19. Thinking DTOs and Entities Are the Same

**❌ WRONG ASSUMPTION:**
"I can just use domain.Entity everywhere"

**Reality:** They serve different purposes!

**Domain Entity** = Business object
```go
// internal/domain/book/entity.go
type Entity struct {
    ID        string
    Name      string
    ISBN      string
    CreatedAt time.Time    // Internal field
    UpdatedAt time.Time    // Internal field
}
```

**DTO (Data Transfer Object)** = External representation
```go
// internal/adapters/http/dto/book.go
type BookResponse struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    ISBN  string `json:"isbn"`
    // No timestamps in response!
}

type CreateBookRequest struct {
    Name    string `json:"name" validate:"required"`
    ISBN    string `json:"isbn" validate:"required"`
    // No ID (server generates it)
}
```

**Why separate:** Domain entity can change without breaking API. DTO shields internal structure.

## 📝 Quick Gotchas Checklist

Before committing, check for these common mistakes:

- [ ] Package naming: Use cases use "ops" suffix
- [ ] No domain → outer layer imports
- [ ] Business logic in domain service, not handler
- [ ] Errors wrapped with `fmt.Errorf("context: %w", err)`
- [ ] `@Security BearerAuth` on protected endpoints
- [ ] SQL queries use `$1` parameters, not string concat
- [ ] Both `.up.sql` and `.down.sql` migration files
- [ ] Unit tests have no database
- [ ] Table-driven tests for multiple scenarios
- [ ] No hardcoded secrets
- [ ] DTOs have validation tags
- [ ] Foreign keys have indexes
- [ ] Ran `make ci` before committing
- [ ] Swagger regenerated after handler changes
- [ ] Commit message follows format (`type: description`)

---

**Learn from these mistakes so you don't repeat them!** 🎓
