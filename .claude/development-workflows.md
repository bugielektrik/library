# Development Workflows

> **Complete step-by-step workflows for common development scenarios**

## Purpose

This guide shows COMPLETE workflows from start to finish, not just individual commands. Follow these flows to ensure you don't miss any steps.

**Difference from other guides:**
- **recipes.md** - Individual commands (copy-paste)
- **examples/** - Code snippets
- **This guide** - Complete end-to-end workflows

---

## 🎯 Workflow Index

- [Adding a New Feature](#workflow-adding-a-new-feature)
- [Fixing a Bug](#workflow-fixing-a-bug)
- [Optimizing Performance](#workflow-optimizing-performance)
- [Refactoring Code](#workflow-refactoring-code)
- [Adding a Database Migration](#workflow-adding-a-database-migration)
- [Responding to Code Review](#workflow-responding-to-code-review)
- [Investigating a Production Issue](#workflow-investigating-a-production-issue)

---

## Workflow: Adding a New Feature

**Scenario:** Product manager says "We need a Loan feature so members can borrow books"

### Phase 1: Planning (5-10 minutes)

```bash
# 1. Create feature branch
git checkout -b feature/add-loans
git push -u origin feature/add-loans

# 2. Read domain documentation
# Read: .claude/glossary.md (understand what a "loan" is)
# Read: .claude/examples/README.md (see similar feature - Book)
# Read: .claude/adrs/001-clean-architecture.md (understand layer separation)
```

**Checklist:**
- [ ] I understand the business domain (what is a loan?)
- [ ] I know which layers I need to touch (domain → use case → adapter → infrastructure)
- [ ] I have a similar feature to reference (Book entity)
- [ ] I created a feature branch

**Output:** Clear understanding of what you're building

---

### Phase 2: Domain Layer (20-30 minutes)

```bash
# 3. Create domain entity
mkdir -p internal/domain/loan
touch internal/domain/loan/entity.go
touch internal/domain/loan/service.go
touch internal/domain/loan/repository.go
touch internal/domain/loan/errors.go
touch internal/domain/loan/entity_test.go
touch internal/domain/loan/service_test.go
```

**In entity.go:**
```go
package loan

import (
    "time"
    "github.com/google/uuid"
)

type Status string

const (
    StatusActive   Status = "active"
    StatusReturned Status = "returned"
    StatusOverdue  Status = "overdue"
)

type Entity struct {
    ID         string
    BookID     string
    MemberID   string
    LoanDate   time.Time
    DueDate    time.Time
    ReturnDate *time.Time
    Status     Status
    LateFee    float64
    CreatedAt  time.Time
    UpdatedAt  time.Time
}

func NewEntity(bookID, memberID string, loanDuration time.Duration) Entity {
    now := time.Now()
    return Entity{
        ID:        uuid.New().String(),
        BookID:    bookID,
        MemberID:  memberID,
        LoanDate:  now,
        DueDate:   now.Add(loanDuration),
        Status:    StatusActive,
        LateFee:   0,
        CreatedAt: now,
        UpdatedAt: now,
    }
}

func (e *Entity) IsOverdue() bool {
    if e.ReturnDate != nil {
        return false // Already returned
    }
    return time.Now().After(e.DueDate)
}
```

**In service.go:**
```go
package loan

import (
    "time"
)

type Service struct{}

func NewService() *Service {
    return &Service{}
}

// Business rule: Calculate late fee ($0.50/day, max $25)
func (s *Service) CalculateLateFee(daysOverdue int) float64 {
    if daysOverdue <= 0 {
        return 0.0
    }

    fee := float64(daysOverdue) * 0.50
    if fee > 25.0 {
        return 25.0 // Max fee
    }
    return fee
}

// Business rule: Calculate days overdue
func (s *Service) CalculateDaysOverdue(dueDate time.Time, returnDate *time.Time) int {
    var endDate time.Time
    if returnDate != nil {
        endDate = *returnDate
    } else {
        endDate = time.Now()
    }

    if endDate.Before(dueDate) {
        return 0
    }

    duration := endDate.Sub(dueDate)
    return int(duration.Hours() / 24)
}
```

**In repository.go:**
```go
package loan

import "context"

type Repository interface {
    Create(ctx context.Context, loan Entity) error
    GetByID(ctx context.Context, id string) (Entity, error)
    GetActiveLoansForMember(ctx context.Context, memberID string) ([]Entity, error)
    Update(ctx context.Context, loan Entity) error
}
```

**Write tests:**
```bash
# Run tests as you write them
go test ./internal/domain/loan/ -v
```

**Checklist:**
- [ ] Entity created with constructor
- [ ] Service contains business logic (CalculateLateFee)
- [ ] Repository interface defined
- [ ] Tests written (100% coverage)
- [ ] All tests pass
- [ ] No external dependencies in domain

---

### Phase 3: Use Case Layer (15-20 minutes)

```bash
# 4. Create use cases
mkdir -p internal/usecase/loanops
touch internal/usecase/loanops/create_loan.go
touch internal/usecase/loanops/return_book.go
touch internal/usecase/loanops/get_loan.go
touch internal/usecase/loanops/create_loan_test.go
```

**In create_loan.go:**
```go
package loanops

import (
    "context"
    "fmt"
    "time"

    "library-service/internal/domain/book"
    "library-service/internal/domain/loan"
    "library-service/internal/domain/member"
)

type CreateLoanUseCase struct {
    loanRepo   loan.Repository
    bookRepo   book.Repository
    memberRepo member.Repository
    loanService *loan.Service
}

func NewCreateLoanUseCase(
    loanRepo loan.Repository,
    bookRepo book.Repository,
    memberRepo member.Repository,
    loanService *loan.Service,
) *CreateLoanUseCase {
    return &CreateLoanUseCase{
        loanRepo:    loanRepo,
        bookRepo:    bookRepo,
        memberRepo:  memberRepo,
        loanService: loanService,
    }
}

type CreateLoanRequest struct {
    BookID   string
    MemberID string
}

func (uc *CreateLoanUseCase) Execute(ctx context.Context, req CreateLoanRequest) (*loan.Entity, error) {
    // 1. Get member (check eligibility)
    member, err := uc.memberRepo.GetByID(ctx, req.MemberID)
    if err != nil {
        return nil, fmt.Errorf("getting member: %w", err)
    }

    // 2. Check member can borrow (no outstanding fees > $10)
    if member.TotalLateFees > 10.0 {
        return nil, fmt.Errorf("member has outstanding late fees: $%.2f", member.TotalLateFees)
    }

    // 3. Get book (check availability)
    book, err := uc.bookRepo.GetByID(ctx, req.BookID)
    if err != nil {
        return nil, fmt.Errorf("getting book: %w", err)
    }

    if book.Status != book.StatusAvailable {
        return nil, fmt.Errorf("book is not available (status: %s)", book.Status)
    }

    // 4. Calculate loan duration based on member subscription
    duration := 14 * 24 * time.Hour // Default: 14 days
    if member.Subscription == "premium" {
        duration = 30 * 24 * time.Hour
    } else if member.Subscription == "vip" {
        duration = 60 * 24 * time.Hour
    }

    // 5. Create loan entity
    loanEntity := loan.NewEntity(req.BookID, req.MemberID, duration)

    // 6. Persist loan
    if err := uc.loanRepo.Create(ctx, loanEntity); err != nil {
        return nil, fmt.Errorf("creating loan: %w", err)
    }

    // 7. Update book status
    book.Status = book.StatusLoaned
    if err := uc.bookRepo.Update(ctx, book); err != nil {
        return nil, fmt.Errorf("updating book status: %w", err)
    }

    return &loanEntity, nil
}
```

**Write use case tests with mocks:**
```bash
go test ./internal/usecase/loanops/ -v
```

**Checklist:**
- [ ] Use cases orchestrate (don't contain business logic)
- [ ] Use cases call domain services for business rules
- [ ] Tests written with mocked repositories
- [ ] All tests pass

---

### Phase 4: Adapter Layer - Repository (10-15 minutes)

```bash
# 5. Implement repository
touch internal/adapters/repository/postgres/loan.go
touch internal/adapters/repository/postgres/loan_test.go
```

**In postgres/loan.go:**
```go
package postgres

import (
    "context"
    "database/sql"
    "fmt"

    "library-service/internal/domain/loan"
)

type LoanRepository struct {
    db *sql.DB
}

func NewLoanRepository(db *sql.DB) loan.Repository {
    return &LoanRepository{db: db}
}

func (r *LoanRepository) Create(ctx context.Context, loanEntity loan.Entity) error {
    query := `
        INSERT INTO loans (id, book_id, member_id, loan_date, due_date, status, late_fee, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    `
    _, err := r.db.ExecContext(ctx, query,
        loanEntity.ID,
        loanEntity.BookID,
        loanEntity.MemberID,
        loanEntity.LoanDate,
        loanEntity.DueDate,
        loanEntity.Status,
        loanEntity.LateFee,
        loanEntity.CreatedAt,
        loanEntity.UpdatedAt,
    )
    if err != nil {
        return fmt.Errorf("inserting loan: %w", err)
    }
    return nil
}

// Implement other methods...
```

**Checklist:**
- [ ] Repository implements domain interface
- [ ] SQL uses parameterized queries ($1, $2, etc.)
- [ ] Integration tests written (optional at this stage)

---

### Phase 5: Database Migration (5 minutes)

```bash
# 6. Create migration
make migrate-create name=create_loans_table
```

**Edit migrations/postgres/XXXXXX_create_loans_table.up.sql:**
```sql
CREATE TABLE IF NOT EXISTS loans (
    id UUID PRIMARY KEY,
    book_id UUID NOT NULL REFERENCES books(id),
    member_id UUID NOT NULL REFERENCES members(id),
    loan_date TIMESTAMP NOT NULL,
    due_date TIMESTAMP NOT NULL,
    return_date TIMESTAMP,
    status VARCHAR(20) NOT NULL,
    late_fee DECIMAL(10, 2) DEFAULT 0.00,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_loans_book_id ON loans(book_id);
CREATE INDEX idx_loans_member_id ON loans(member_id);
CREATE INDEX idx_loans_status ON loans(status);
CREATE INDEX idx_loans_due_date ON loans(due_date) WHERE status = 'active';
```

**Edit migrations/postgres/XXXXXX_create_loans_table.down.sql:**
```sql
DROP INDEX IF EXISTS idx_loans_due_date;
DROP INDEX IF EXISTS idx_loans_status;
DROP INDEX IF EXISTS idx_loans_member_id;
DROP INDEX IF EXISTS idx_loans_book_id;
DROP TABLE IF EXISTS loans;
```

**Apply migration:**
```bash
make migrate-up
```

**Checklist:**
- [ ] Migration created (both up and down)
- [ ] Foreign keys defined
- [ ] Indexes added (especially on foreign keys)
- [ ] Migration applied successfully

---

### Phase 6: Adapter Layer - HTTP (15-20 minutes)

```bash
# 7. Create HTTP layer
touch internal/adapters/http/dto/loan.go
touch internal/adapters/http/handlers/loan.go
touch internal/adapters/http/handlers/loan_test.go
```

**In dto/loan.go:**
```go
package dto

import "time"

type CreateLoanRequest struct {
    BookID   string `json:"book_id" validate:"required,uuid"`
    MemberID string `json:"member_id" validate:"required,uuid"`
}

type LoanResponse struct {
    ID         string     `json:"id"`
    BookID     string     `json:"book_id"`
    MemberID   string     `json:"member_id"`
    LoanDate   time.Time  `json:"loan_date"`
    DueDate    time.Time  `json:"due_date"`
    ReturnDate *time.Time `json:"return_date,omitempty"`
    Status     string     `json:"status"`
    LateFee    float64    `json:"late_fee"`
}
```

**In handlers/loan.go:**
```go
package handlers

import (
    "encoding/json"
    "net/http"

    "library-service/internal/adapters/http/dto"
    "library-service/internal/usecase/loanops"
)

type LoanHandler struct {
    createLoanUC *loanops.CreateLoanUseCase
}

func NewLoanHandler(createLoanUC *loanops.CreateLoanUseCase) *LoanHandler {
    return &LoanHandler{
        createLoanUC: createLoanUC,
    }
}

// CreateLoan creates a new loan
// @Summary      Create a new loan
// @Description  Member borrows a book
// @Tags         loans
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateLoanRequest true "Loan request"
// @Success      201 {object} dto.LoanResponse
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Router       /loans [post]
func (h *LoanHandler) CreateLoan(w http.ResponseWriter, r *http.Request) {
    var req dto.CreateLoanRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, "invalid request body")
        return
    }

    // Validate
    if err := validate.Struct(req); err != nil {
        respondError(w, http.StatusBadRequest, err.Error())
        return
    }

    // Execute use case
    useCaseReq := loanops.CreateLoanRequest{
        BookID:   req.BookID,
        MemberID: req.MemberID,
    }

    loan, err := h.createLoanUC.Execute(r.Context(), useCaseReq)
    if err != nil {
        respondError(w, http.StatusBadRequest, err.Error())
        return
    }

    // Map to response
    response := dto.LoanResponse{
        ID:         loan.ID,
        BookID:     loan.BookID,
        MemberID:   loan.MemberID,
        LoanDate:   loan.LoanDate,
        DueDate:    loan.DueDate,
        ReturnDate: loan.ReturnDate,
        Status:     string(loan.Status),
        LateFee:    loan.LateFee,
    }

    respondJSON(w, response, http.StatusCreated)
}
```

**Checklist:**
- [ ] DTOs created for request/response
- [ ] Handler delegates to use case
- [ ] Swagger annotations added
- [ ] Error handling implemented

---

### Phase 7: Wiring (5 minutes)

**Edit internal/infrastructure/container/container.go:**
```go
// Add to Container struct
type Container struct {
    // ... existing fields
    LoanRepo     loan.Repository
    LoanService  *loan.Service
    CreateLoanUC *loanops.CreateLoanUseCase
    LoanHandler  *handlers.LoanHandler
}

// In New() function
func New(app *app.App) *Container {
    // ... existing code

    // Loan
    loanRepo := postgres.NewLoanRepository(app.Store.DB)
    loanService := loan.NewService()
    createLoanUC := loanops.NewCreateLoanUseCase(loanRepo, bookRepo, memberRepo, loanService)
    loanHandler := handlers.NewLoanHandler(createLoanUC)

    return &Container{
        // ... existing fields
        LoanRepo:     loanRepo,
        LoanService:  loanService,
        CreateLoanUC: createLoanUC,
        LoanHandler:  loanHandler,
    }
}
```

**Edit internal/adapters/http/routes/router.go:**
```go
func NewRouter(container *container.Container) http.Handler {
    r := chi.NewRouter()

    // ... existing routes

    r.Route("/api/v1", func(r chi.Router) {
        // ... existing routes

        // Loans (protected)
        r.Route("/loans", func(r chi.Router) {
            r.Use(authMiddleware)
            r.Post("/", container.LoanHandler.CreateLoan)
        })
    })

    return r
}
```

**Checklist:**
- [ ] Dependencies wired in container.go
- [ ] Routes added in router.go
- [ ] Middleware applied (auth if needed)

---

### Phase 8: Testing & Documentation (10-15 minutes)

```bash
# 8. Generate Swagger docs
make gen-docs

# 9. Run all tests
make test

# 10. Run linter
make lint

# 11. Test manually
make run

# In another terminal:
# Get auth token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test123!@#"}' \
  | jq -r '.tokens.access_token')

# Create a loan
curl -X POST http://localhost:8080/api/v1/loans \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "book_id": "your-book-id",
    "member_id": "your-member-id"
  }'
```

**Checklist:**
- [ ] All tests pass
- [ ] Swagger docs updated
- [ ] Manual testing successful
- [ ] No linter errors

---

### Phase 9: Pre-Commit Checks (5 minutes)

```bash
# 12. Run full CI locally
make ci

# Or use automated review script
.claude/scripts/review.sh

# 13. Review checklist
# Read: .claude/checklist.md
```

**Checklist:**
- [ ] All tests pass (make test)
- [ ] No linter errors (make lint)
- [ ] Code formatted (make fmt)
- [ ] Build successful (make build)
- [ ] Swagger docs updated
- [ ] Database migration has up AND down
- [ ] No hardcoded secrets
- [ ] Errors wrapped with %w

---

### Phase 10: Commit & Push (5 minutes)

```bash
# 14. Stage changes
git add .

# 15. Commit with conventional commit message
git commit -m "feat: add loan management feature

Implemented loan domain for book borrowing:
- Domain layer: Loan entity, service (late fee calculation), repository interface
- Use cases: CreateLoan (orchestrates book availability check, member eligibility)
- Adapters: PostgreSQL repository, HTTP handlers with Swagger docs
- Database: loans table with indexes on foreign keys
- Tests: 100% domain coverage, 85% use case coverage

Business rules implemented:
- Late fee: \$0.50/day, max \$25
- Loan duration: 14/30/60 days based on subscription tier
- Members with >$10 late fees cannot borrow

Co-Authored-By: Claude <noreply@anthropic.com>"

# 16. Push
git push origin feature/add-loans
```

**Checklist:**
- [ ] Commit message follows conventional commits
- [ ] Commit includes Co-Authored-By: Claude
- [ ] Pushed to feature branch

---

### Phase 11: Create Pull Request (5 minutes)

```bash
# 17. Create PR
gh pr create --title "feat: Add loan management feature" --body "$(cat <<'EOF'
## Summary
Implements loan management feature allowing members to borrow books.

## Changes
- ✅ Domain layer: Loan entity with business logic
- ✅ Use cases: CreateLoan, ReturnBook
- ✅ Repository: PostgreSQL implementation
- ✅ API: POST /loans endpoint with JWT auth
- ✅ Database: Migration for loans table
- ✅ Tests: 100% domain, 85% use case coverage

## Business Rules
- Late fee: $0.50/day (max $25)
- Loan duration based on subscription tier
- Members with >$10 late fees blocked

## Testing
- [ ] Unit tests pass (100% domain coverage)
- [ ] Integration tests pass
- [ ] Manual API testing successful
- [ ] Swagger docs updated

## Checklist
- [x] Tests added
- [x] Documentation updated
- [x] Database migration included
- [x] No breaking changes
- [x] make ci passes

🤖 Generated with [Claude Code](https://claude.com/claude-code)
EOF
)"
```

**Total Time:** ~2 hours from idea to PR

---

## Workflow: Fixing a Bug

**Scenario:** "Books can be borrowed even when status is 'maintenance'"

### Phase 1: Reproduce (5-10 minutes)

```bash
# 1. Create bug fix branch
git checkout -b fix/prevent-maintenance-book-borrowing

# 2. Write failing test that reproduces bug
# Edit: internal/usecase/loanops/create_loan_test.go
```

**Add test:**
```go
func TestCreateLoanUseCase_Execute_BookInMaintenance_ReturnsError(t *testing.T) {
    // Setup
    mockBookRepo := &mocks.MockBookRepository{
        GetByIDFunc: func(ctx context.Context, id string) (book.Entity, error) {
            return book.Entity{
                ID:     id,
                Status: book.StatusMaintenance,  // Bug: should prevent borrowing
            }, nil
        },
    }

    mockMemberRepo := &mocks.MockMemberRepository{
        GetByIDFunc: func(ctx context.Context, id string) (member.Entity, error) {
            return member.Entity{
                ID:             id,
                TotalLateFees:  0,
                Subscription:   "basic",
            }, nil
        },
    }

    uc := loanops.NewCreateLoanUseCase(
        &mocks.MockLoanRepository{},
        mockBookRepo,
        mockMemberRepo,
        loan.NewService(),
    )

    // Execute
    _, err := uc.Execute(context.Background(), loanops.CreateLoanRequest{
        BookID:   "book-123",
        MemberID: "member-456",
    })

    // Assert
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "not available")
}
```

**Run test (should FAIL):**
```bash
go test ./internal/usecase/loanops/ -v -run TestCreateLoanUseCase_Execute_BookInMaintenance
# FAIL (bug reproduced)
```

**Checklist:**
- [ ] Bug reproduced with failing test
- [ ] Test is clear about expected behavior

---

### Phase 2: Locate Bug (5 minutes)

**Read the code:**
```bash
# Find where book status is checked
grep -r "StatusAvailable" internal/usecase/loanops/

# Read the file
cat internal/usecase/loanops/create_loan.go
```

**Found the bug:**
```go
// Line 45 in create_loan.go
if book.Status != book.StatusAvailable {
    return nil, fmt.Errorf("book is not available (status: %s)", book.Status)
}
```

**Analysis:** Code correctly checks status. Let me check if book statuses are defined...

```bash
grep -r "StatusMaintenance" internal/domain/book/
```

**Root cause:** `StatusMaintenance` doesn't exist! Only `StatusAvailable` and `StatusLoaned` are defined.

**Checklist:**
- [ ] Bug located
- [ ] Root cause understood

---

### Phase 3: Fix Bug (10 minutes)

**Fix 1: Add missing status to domain**
```go
// Edit: internal/domain/book/entity.go
const (
    StatusAvailable  Status = "available"
    StatusLoaned     Status = "loaned"
    StatusMaintenance Status = "maintenance"  // ← ADD THIS
    StatusLost       Status = "lost"          // ← ADD THIS
    StatusReserved   Status = "reserved"      // ← ADD THIS
)
```

**Fix 2: Ensure use case checks correctly (already correct)**

**Fix 3: Add database migration for new statuses**
```bash
make migrate-create name=add_book_statuses
```

**Edit migration (up):**
```sql
-- No table changes needed, just documentation
-- New valid statuses: available, loaned, maintenance, lost, reserved
-- Add constraint to enforce valid statuses
ALTER TABLE books
DROP CONSTRAINT IF EXISTS books_status_check;

ALTER TABLE books
ADD CONSTRAINT books_status_check
CHECK (status IN ('available', 'loaned', 'maintenance', 'lost', 'reserved'));
```

**Edit migration (down):**
```sql
ALTER TABLE books
DROP CONSTRAINT IF EXISTS books_status_check;

ALTER TABLE books
ADD CONSTRAINT books_status_check
CHECK (status IN ('available', 'loaned'));
```

**Apply migration:**
```bash
make migrate-up
```

**Run tests:**
```bash
# Test should now PASS
go test ./internal/usecase/loanops/ -v -run TestCreateLoanUseCase_Execute_BookInMaintenance
# PASS

# Run all tests
make test
```

**Checklist:**
- [ ] Bug fixed
- [ ] Test now passes
- [ ] All other tests still pass
- [ ] Database migration added

---

### Phase 4: Commit & PR (10 minutes)

```bash
git add .
git commit -m "fix: prevent borrowing books in maintenance status

Fixed bug where books with 'maintenance' status could be borrowed.

Changes:
- Added missing book statuses (maintenance, lost, reserved) to domain
- Added database constraint to enforce valid statuses
- Added test covering maintenance status

Fixes #123"

git push origin fix/prevent-maintenance-book-borrowing

gh pr create --title "fix: prevent borrowing books in maintenance status" --body "$(cat <<'EOF'
## Bug
Books with 'maintenance' status could be borrowed.

## Root Cause
StatusMaintenance, StatusLost, and StatusReserved were not defined in domain.

## Fix
- Added missing statuses to book.Entity
- Added database constraint to enforce valid statuses
- Added test to prevent regression

## Testing
- [x] Test added that reproduces bug
- [x] Test now passes with fix
- [x] All existing tests pass
- [x] Database migration included

Fixes #123
EOF
)"
```

**Total Time:** ~30 minutes from bug report to PR

---

## Workflow: Optimizing Performance

**Scenario:** "GET /books endpoint is slow (500ms response time)"

### Phase 1: Establish Baseline (10 minutes)

```bash
# 1. Profile current performance
# Start API
make run

# In another terminal, run load test
ab -n 1000 -c 10 http://localhost:8080/api/v1/books

# Results:
# Time per request: 523ms (mean)
# Requests per second: 19.1
```

**Benchmark test:**
```go
// Add: internal/adapters/repository/postgres/book_test.go
func BenchmarkBookRepository_List(b *testing.B) {
    db := setupTestDB(b)
    repo := postgres.NewBookRepository(db)
    ctx := context.Background()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := repo.List(ctx, 50, 0)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

**Run benchmark:**
```bash
go test ./internal/adapters/repository/postgres/ -bench=BenchmarkBookRepository_List -benchmem

# Results:
# 100 iterations
# 45231 ns/op (45ms per operation)
# 12000 B/op
# 150 allocs/op
```

**Checklist:**
- [ ] Baseline established (523ms)
- [ ] Benchmark test written

---

### Phase 2: Profile & Identify Bottleneck (15 minutes)

```bash
# CPU profiling
go test ./internal/adapters/repository/postgres/ \
  -bench=BenchmarkBookRepository_List \
  -cpuprofile=cpu.prof

# Analyze
go tool pprof cpu.prof
# (pprof) top
# (pprof) list BookRepository.List
```

**Findings:**
```
60% time in: r.db.Query() → SQL query
25% time in: row.Scan() → result parsing
10% time in: JSON marshaling
```

**Check SQL query:**
```bash
# Enable PostgreSQL query logging
# Edit: docker-compose.yml
# Add to postgres environment:
#   POSTGRES_LOG_STATEMENT: 'all'

# Restart
make down && make up

# Make request and check logs
docker logs $(docker ps -qf "name=postgres") | grep SELECT
```

**Found:**
```sql
SELECT * FROM books ORDER BY created_at DESC LIMIT 50 OFFSET 0;

-- Problem: No index on created_at
-- Seq Scan on books (cost=0.00..1234.56 rows=50000)
```

**Checklist:**
- [ ] Bottleneck identified (missing index on created_at)
- [ ] Profiling data collected

---

### Phase 3: Implement Optimization (10 minutes)

```bash
# Create migration
make migrate-create name=add_books_created_at_index
```

**Edit migration (up):**
```sql
CREATE INDEX IF NOT EXISTS idx_books_created_at ON books(created_at DESC);

-- Analyze table to update statistics
ANALYZE books;
```

**Edit migration (down):**
```sql
DROP INDEX IF EXISTS idx_books_created_at;
```

**Apply:**
```bash
make migrate-up
```

**Verify index created:**
```sql
docker exec -it $(docker ps -qf "name=postgres") \
  psql -U library -d library -c "\d books"

-- Should show:
-- Indexes:
--   "idx_books_created_at" btree (created_at DESC)
```

**Checklist:**
- [ ] Optimization implemented (index added)
- [ ] Migration tested

---

### Phase 4: Measure Improvement (10 minutes)

```bash
# Re-run benchmark
go test ./internal/adapters/repository/postgres/ \
  -bench=BenchmarkBookRepository_List -benchmem

# Results:
# 2341 ns/op (2.3ms per operation - 20x faster!)
# 8000 B/op (33% less memory)
# 100 allocs/op (33% fewer allocations)

# Re-run load test
ab -n 1000 -c 10 http://localhost:8080/api/v1/books

# Results:
# Time per request: 25ms (mean) - 21x faster!
# Requests per second: 400 (21x improvement)
```

**Comparison:**
| Metric | Before | After | Improvement |
|--------|--------|-------|-------------|
| Response time | 523ms | 25ms | 21x faster |
| RPS | 19.1 | 400 | 21x more |
| Memory | 12000 B/op | 8000 B/op | 33% less |

**Checklist:**
- [ ] Performance improved significantly
- [ ] Metrics documented

---

### Phase 5: Commit Optimization (5 minutes)

```bash
git add .
git commit -m "perf: add index on books.created_at

Optimized GET /books endpoint by adding index.

Performance improvement:
- Response time: 523ms → 25ms (21x faster)
- Throughput: 19 RPS → 400 RPS (21x increase)
- Memory: 12KB → 8KB per operation (33% reduction)

Changes:
- Added idx_books_created_at DESC index
- Benchmark test added

Resolves #performance-issue-45"

git push origin perf/optimize-books-list
```

**Total Time:** ~50 minutes from identification to optimization

---

## Summary: Common Workflows Time Estimates

| Workflow | Estimated Time |
|----------|----------------|
| Add new feature (full stack) | 2-3 hours |
| Fix bug | 30-60 minutes |
| Optimize performance | 1 hour |
| Refactor code | 1-2 hours |
| Add database migration | 15 minutes |
| Add API endpoint only | 30 minutes |
| Code review response | 30 minutes |

**Total documentation:** Shows complete flows, not just snippets.

---

**Last Updated:** 2025-01-19
**Next Review:** When development process changes
