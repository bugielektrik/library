# Quick Wins Guide

> **Safe, high-value improvements Claude Code can suggest immediately**

## Purpose

This guide lists improvements that are:
- ✅ **Safe** - Won't break existing functionality
- ✅ **High-value** - Noticeable improvement in code quality/performance
- ✅ **Quick** - Can be implemented in 5-30 minutes
- ✅ **Always beneficial** - No downside

**Use this when:** You want to proactively improve the codebase beyond just completing the user's task.

---

## 🎯 When to Suggest Quick Wins

**DO suggest when:**
- User asks "anything else I should improve?"
- You notice the issue while working on nearby code
- User is refactoring/touching the same area
- You're adding a new feature and see related issues

**DON'T suggest when:**
- User is fixing an urgent bug (stay focused)
- User explicitly said "just do X" (respect scope)
- You haven't verified the improvement is safe
- It would delay the primary task significantly

**Always ask first:** "I noticed X while working on this. Would you like me to fix it, or should I stay focused on the current task?"

---

## 🔍 Code Quality Quick Wins

### 1. Add Missing Error Wrapping

**Look for:**
```go
// ❌ Missing context
if err := repo.Create(ctx, book); err != nil {
    return err
}
```

**Suggest:**
```go
// ✅ Context added
if err := repo.Create(ctx, book); err != nil {
    return fmt.Errorf("creating book in repository: %w", err)
}
```

**Why:** Makes debugging 10x easier (know where error originated)
**Time:** 2 minutes per file
**Risk:** Zero (only adds context)

---

### 2. Add Missing Validation

**Look for:**
```go
// ❌ No validation
func NewEntity(title, isbn string) Entity {
    return Entity{
        ID:    uuid.New().String(),
        Title: title,
        ISBN:  isbn,
    }
}
```

**Suggest:**
```go
// ✅ Validation added
func NewEntity(title, isbn string) (Entity, error) {
    if title == "" {
        return Entity{}, errors.New("title cannot be empty")
    }
    if len(isbn) != 13 {
        return Entity{}, errors.New("ISBN must be 13 digits")
    }

    return Entity{
        ID:    uuid.New().String(),
        Title: title,
        ISBN:  isbn,
    }, nil
}
```

**Why:** Prevents invalid entities from being created
**Time:** 5 minutes
**Risk:** Low (might need to update calling code)

---

### 3. Replace Magic Numbers with Constants

**Look for:**
```go
// ❌ Magic numbers
if len(isbn) != 13 {
    return errors.New("invalid ISBN")
}

fee := daysLate * 0.50
if fee > 25.0 {
    fee = 25.0
}
```

**Suggest:**
```go
// ✅ Named constants
const (
    ISBNLength      = 13
    LateFeePerDay   = 0.50
    MaxLateFee      = 25.0
)

if len(isbn) != ISBNLength {
    return errors.New("invalid ISBN")
}

fee := float64(daysLate) * LateFeePerDay
if fee > MaxLateFee {
    fee = MaxLateFee
}
```

**Why:** Self-documenting code, easier to change business rules
**Time:** 5 minutes
**Risk:** Zero

---

### 4. Add Missing Tests

**Look for:**
- Functions with no `_test.go` file
- Test coverage < 80% for use cases
- Test coverage < 100% for domain services

**Suggest:**
```go
// Add tests for uncovered edge cases
func TestService_CalculateLateFee_NegativeDays_ReturnsZero(t *testing.T) {
    svc := NewService()
    fee := svc.CalculateLateFee(-5)
    assert.Equal(t, 0.0, fee)
}

func TestService_CalculateLateFee_ExceedsMax_ReturnsMax(t *testing.T) {
    svc := NewService()
    fee := svc.CalculateLateFee(100)  // Would be $50
    assert.Equal(t, 25.0, fee)  // Capped at $25
}
```

**Why:** Prevents regressions, documents expected behavior
**Time:** 10-20 minutes
**Risk:** Zero (only adds tests)

---

### 5. Extract Repeated Logic into Helper Functions

**Look for:**
```go
// ❌ Repeated code
func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
    var req dto.CreateBookRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(dto.ErrorResponse{Error: err.Error()})
        return
    }
}

func (h *BookHandler) UpdateBook(w http.ResponseWriter, r *http.Request) {
    var req dto.UpdateBookRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(dto.ErrorResponse{Error: err.Error()})
        return
    }
}
```

**Suggest:**
```go
// ✅ Extracted helper
func decodeJSON(w http.ResponseWriter, r *http.Request, v interface{}) error {
    if err := json.NewDecoder(r.Body).Decode(v); err != nil {
        respondError(w, http.StatusBadRequest, "invalid JSON")
        return err
    }
    return nil
}

func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
    var req dto.CreateBookRequest
    if err := decodeJSON(w, r, &req); err != nil {
        return
    }
    // ...
}
```

**Why:** DRY principle, easier to maintain
**Time:** 10 minutes
**Risk:** Low (test thoroughly)

---

## 🚀 Performance Quick Wins

### 6. Add Missing Database Indexes

**Look for:**
```sql
-- Query in logs is slow
SELECT * FROM books WHERE isbn = '1234567890';  -- 150ms

-- Check if index exists
\d books
-- No index on isbn column
```

**Suggest:**
```bash
# Create migration
make migrate-create name=add_books_isbn_index
```

```sql
-- up migration
CREATE INDEX IF NOT EXISTS idx_books_isbn ON books(isbn);

-- down migration
DROP INDEX IF EXISTS idx_books_isbn;
```

**Why:** 10-100x faster queries
**Time:** 5 minutes
**Risk:** Very low (indexes only improve reads, slightly slower writes)

**When to suggest:**
- Foreign key columns without indexes
- Columns used in WHERE clauses
- Columns used in ORDER BY

---

### 7. Pre-allocate Slices with Known Capacity

**Look for:**
```go
// ❌ Slice grows dynamically (multiple allocations)
var results []Book
for _, item := range items {  // 1000 items
    results = append(results, process(item))
}
// Allocates: ~14 times (capacity: 0→1→2→4→8→16→...→1024)
```

**Suggest:**
```go
// ✅ Pre-allocated (single allocation)
results := make([]Book, 0, len(items))
for _, item := range items {
    results = append(results, process(item))
}
// Allocates: 1 time
```

**Why:** 30-50% fewer allocations, less GC pressure
**Time:** 2 minutes
**Risk:** Zero

---

### 8. Use Context Timeouts

**Look for:**
```go
// ❌ No timeout
ctx := context.Background()
book, err := repo.GetByID(ctx, id)
```

**Suggest:**
```go
// ✅ Timeout added
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

book, err := repo.GetByID(ctx, id)
```

**Why:** Prevents hanging requests, better resource management
**Time:** 2 minutes
**Risk:** Low (ensure timeout is reasonable)

---

## 🛡️ Security Quick Wins

### 9. Add Input Validation to DTOs

**Look for:**
```go
// ❌ No validation tags
type CreateBookRequest struct {
    Title string `json:"title"`
    ISBN  string `json:"isbn"`
}
```

**Suggest:**
```go
// ✅ Validation added
type CreateBookRequest struct {
    Title string `json:"title" validate:"required,min=1,max=255"`
    ISBN  string `json:"isbn" validate:"required,len=13,numeric"`
}
```

**Why:** Prevents invalid data from entering system
**Time:** 5 minutes
**Risk:** Low (might need to handle validation errors)

---

### 10. Check for Hardcoded Secrets

**Look for:**
```go
// ❌ Hardcoded secret
jwtManager := auth.NewJWTManager("my-secret-key-123")
```

**Suggest:**
```go
// ✅ From environment
jwtSecret := os.Getenv("JWT_SECRET")
if jwtSecret == "" {
    log.Fatal("JWT_SECRET environment variable not set")
}
jwtManager := auth.NewJWTManager(jwtSecret)
```

**Why:** Critical security issue
**Time:** 5 minutes
**Risk:** Zero (assuming .env is configured)

---

### 11. Add Missing @Security Annotations

**Look for:**
```go
// ❌ Protected endpoint, no Swagger annotation
// @Summary Get book by ID
// @Router /books/{id} [get]
func (h *BookHandler) GetBook(w http.ResponseWriter, r *http.Request)
```

**Suggest:**
```go
// ✅ Security annotation added
// @Summary Get book by ID
// @Security BearerAuth
// @Router /books/{id} [get]
func (h *BookHandler) GetBook(w http.ResponseWriter, r *http.Request)
```

**Why:** Swagger UI knows endpoint requires auth
**Time:** 1 minute
**Risk:** Zero (documentation only)

---

## 📝 Documentation Quick Wins

### 12. Add Missing GoDoc Comments

**Look for:**
```go
// ❌ No documentation
type Service struct {
    // ...
}

func (s *Service) ValidateISBN(isbn string) error {
    // ...
}
```

**Suggest:**
```go
// ✅ GoDoc added
// Service contains business logic for book operations.
type Service struct {
    // ...
}

// ValidateISBN checks if an ISBN-13 is valid using checksum validation.
// Returns an error if the ISBN is not 13 digits or has an invalid checksum.
func (s *Service) ValidateISBN(isbn string) error {
    // ...
}
```

**Why:** Better IDE tooltips, godoc.org documentation
**Time:** 10 minutes
**Risk:** Zero

---

### 13. Add Missing Swagger Examples

**Look for:**
```go
// @Param request body dto.CreateBookRequest true "Book data"
```

**Suggest:**
```go
// @Param request body dto.CreateBookRequest true "Book data" example({"title":"The Great Gatsby","isbn":"9780743273565"})
```

**Why:** Better API documentation in Swagger UI
**Time:** 2 minutes per endpoint
**Risk:** Zero

---

## 🧪 Testing Quick Wins

### 14. Add Table-Driven Tests

**Look for:**
```go
// ❌ Repetitive tests
func TestValidateISBN_Valid(t *testing.T) {
    svc := NewService()
    err := svc.ValidateISBN("9780743273565")
    assert.NoError(t, err)
}

func TestValidateISBN_TooShort(t *testing.T) {
    svc := NewService()
    err := svc.ValidateISBN("123")
    assert.Error(t, err)
}
```

**Suggest:**
```go
// ✅ Table-driven
func TestService_ValidateISBN(t *testing.T) {
    tests := []struct {
        name    string
        isbn    string
        wantErr bool
    }{
        {"valid ISBN-13", "9780743273565", false},
        {"too short", "123", true},
        {"too long", "12345678901234", true},
        {"invalid checksum", "9780743273564", true},
    }

    svc := NewService()
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := svc.ValidateISBN(tt.isbn)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

**Why:** Easier to add new test cases, more comprehensive
**Time:** 15 minutes
**Risk:** Zero

---

### 15. Add Benchmark Tests for Domain Logic

**Look for:**
- Domain services with no benchmarks
- Performance-critical code

**Suggest:**
```go
func BenchmarkService_ValidateISBN(b *testing.B) {
    svc := NewService()
    isbn := "9780743273565"

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _ = svc.ValidateISBN(isbn)
    }
}
```

**Why:** Catch performance regressions early
**Time:** 5 minutes
**Risk:** Zero

---

## 🗄️ Database Quick Wins

### 16. Add Indexes on Foreign Keys

**Check migrations for:**
```sql
-- ❌ Foreign key without index
CREATE TABLE loans (
    id UUID PRIMARY KEY,
    book_id UUID NOT NULL REFERENCES books(id),
    member_id UUID NOT NULL REFERENCES members(id)
);
-- No indexes on book_id or member_id
```

**Suggest:**
```sql
-- ✅ Indexes added
CREATE TABLE loans (
    id UUID PRIMARY KEY,
    book_id UUID NOT NULL REFERENCES books(id),
    member_id UUID NOT NULL REFERENCES members(id)
);

CREATE INDEX idx_loans_book_id ON loans(book_id);
CREATE INDEX idx_loans_member_id ON loans(member_id);
```

**Why:** JOINs will be 10-100x faster
**Time:** 2 minutes
**Risk:** Very low

---

### 17. Add NOT NULL Constraints

**Look for:**
```sql
-- ❌ Missing NOT NULL
CREATE TABLE books (
    id UUID PRIMARY KEY,
    title VARCHAR(255),  -- Should this be nullable?
    isbn VARCHAR(13)     -- Probably not
);
```

**Suggest:**
```sql
-- ✅ Constraints added
CREATE TABLE books (
    id UUID PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    isbn VARCHAR(13) NOT NULL UNIQUE
);
```

**Why:** Database enforces data integrity
**Time:** 5 minutes (requires migration)
**Risk:** Medium (ensure no existing NULL values)

---

## 🏗️ Architecture Quick Wins

### 18. Move Business Logic from Use Case to Domain Service

**Look for:**
```go
// ❌ Business logic in use case
// internal/usecase/bookops/create_book.go
func (uc *CreateBookUseCase) Execute(req Request) error {
    // Business logic here (should be in domain!)
    if len(req.ISBN) != 13 {
        return errors.New("ISBN must be 13 digits")
    }

    sum := 0
    for i, r := range req.ISBN[:12] {
        // Complex ISBN checksum validation...
    }

    // More business logic...
}
```

**Suggest:**
```go
// ✅ Moved to domain service
// internal/domain/book/service.go
func (s *Service) ValidateISBN(isbn string) error {
    if len(isbn) != 13 {
        return errors.New("ISBN must be 13 digits")
    }
    // Checksum validation logic...
}

// internal/usecase/bookops/create_book.go
func (uc *CreateBookUseCase) Execute(req Request) error {
    // Use case orchestrates
    if err := uc.bookService.ValidateISBN(req.ISBN); err != nil {
        return err
    }
    // ...
}
```

**Why:** Proper separation of concerns, easier to test
**Time:** 15 minutes
**Risk:** Low (move logic, update tests)

---

### 19. Replace Import Aliases with "ops" Suffix

**Look for:**
```go
// ❌ Import alias needed
import (
    "library-service/internal/domain/book"
    usecaseBook "library-service/internal/usecase/book"  // Alias required
)
```

**Suggest:**
```go
// ✅ Rename package to bookops
// Rename: internal/usecase/book/ → internal/usecase/bookops/

import (
    "library-service/internal/domain/book"
    "library-service/internal/usecase/bookops"  // No alias needed
)
```

**Why:** Follows project convention, clearer code
**Time:** 10 minutes (rename package, update imports)
**Risk:** Low (IDE can help refactor)

---

## 📋 Quick Win Checklist

**Before suggesting a quick win:**
- [ ] I've verified this is a real improvement (not just different style)
- [ ] I've checked it doesn't break existing tests
- [ ] I've estimated the time accurately (< 30 minutes)
- [ ] I've asked the user if they want me to do it
- [ ] I've explained the benefit clearly

**When suggesting:**
```
"I noticed while working on this that [X]. Would you like me to [fix/improve] it?

It would take about [Y minutes] and would [benefit].

Should I:
1. Fix it now as part of this change
2. Leave it for later
3. Add it to a TODO for future work"
```

---

## 🎯 Prioritization Guide

**High Priority (Always suggest if you see them):**
1. Hardcoded secrets (security critical)
2. SQL injection vulnerabilities (security critical)
3. Missing error wrapping (makes debugging 10x harder)
4. Missing indexes on foreign keys (huge performance impact)
5. Missing NOT NULL on required fields (data integrity)

**Medium Priority (Suggest if working nearby):**
6. Missing validation
7. Magic numbers → constants
8. Missing tests (especially domain layer)
9. Pre-allocate slices
10. Add context timeouts

**Low Priority (Mention if user asks for improvements):**
11. GoDoc comments
12. Table-driven tests
13. Swagger examples
14. Benchmark tests

---

## 💡 Pro Tips

1. **One at a time:** Suggest one quick win at a time, not a list of 10
2. **Provide the fix:** Don't just identify the issue, show the code
3. **Explain the why:** Users learn when you explain the benefit
4. **Respect scope:** If user said "just fix the bug," stay focused
5. **Test your suggestion:** Make sure the improvement actually works

---

## ⚠️ When NOT to Suggest Quick Wins

**Don't suggest if:**
- User is in a hurry (urgent bug fix, production issue)
- User explicitly said "minimal changes only"
- Your suggestion would require > 30 minutes
- You're not 100% sure the change is safe
- It's a matter of style preference (tabs vs spaces, etc.)
- The "improvement" is debatable (not clearly better)

**Remember:** Your goal is to be helpful, not to rewrite the entire codebase.

---

**Last Updated:** 2025-01-19
**See Also:**
- [checklist.md](./checklist.md) - Pre-commit review checklist
- [gotchas.md](./gotchas.md) - Mistakes to avoid
- [security.md](./security.md) - Security best practices
