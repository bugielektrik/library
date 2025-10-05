# Code Examples

> **Quick copy-paste examples for common tasks in this codebase**

## Quick Navigation

- [Adding a New Domain](#adding-a-new-domain)
- [Adding a New Use Case](#adding-a-new-use-case)
- [Adding a New API Endpoint](#adding-a-new-api-endpoint)
- [Writing Tests](#writing-tests)
- [Common Patterns](#common-patterns)

## Adding a New Domain

### 1. Domain Entity

```go
// internal/domain/loan/entity.go
package loan

import (
    "time"
    "github.com/google/uuid"
)

type Entity struct {
    ID         string
    BookID     string
    MemberID   string
    LoanDate   time.Time
    DueDate    time.Time
    ReturnDate *time.Time
    Status     Status
    CreatedAt  time.Time
    UpdatedAt  time.Time
}

type Status string

const (
    StatusActive   Status = "active"
    StatusReturned Status = "returned"
    StatusOverdue  Status = "overdue"
)

func NewEntity(bookID, memberID string, loanDuration time.Duration) Entity {
    now := time.Now()
    return Entity{
        ID:        uuid.New().String(),
        BookID:    bookID,
        MemberID:  memberID,
        LoanDate:  now,
        DueDate:   now.Add(loanDuration),
        Status:    StatusActive,
        CreatedAt: now,
        UpdatedAt: now,
    }
}

func (e *Entity) IsOverdue() bool {
    return e.Status == StatusActive && time.Now().After(e.DueDate)
}

func (e *Entity) Return() error {
    if e.Status == StatusReturned {
        return ErrAlreadyReturned
    }
    now := time.Now()
    e.ReturnDate = &now
    e.Status = StatusReturned
    e.UpdatedAt = now
    return nil
}
```

### 2. Domain Service

```go
// internal/domain/loan/service.go
package loan

import (
    "time"
)

var (
    ErrAlreadyReturned = errors.New("loan already returned")
    ErrInvalidDuration = errors.New("invalid loan duration")
)

type Service struct {
    maxLoanDuration time.Duration
}

func NewService() *Service {
    return &Service{
        maxLoanDuration: 14 * 24 * time.Hour, // 14 days
    }
}

func (s *Service) ValidateLoanDuration(duration time.Duration) error {
    if duration <= 0 {
        return ErrInvalidDuration
    }
    if duration > s.maxLoanDuration {
        return ErrInvalidDuration
    }
    return nil
}

func (s *Service) CalculateLateFee(loan Entity) float64 {
    if loan.ReturnDate == nil {
        return 0
    }
    if loan.ReturnDate.Before(loan.DueDate) {
        return 0
    }
    daysLate := int(loan.ReturnDate.Sub(loan.DueDate).Hours() / 24)
    return float64(daysLate) * 0.50 // $0.50 per day
}
```

### 3. Repository Interface

```go
// internal/domain/loan/repository.go
package loan

import "context"

type Repository interface {
    Create(ctx context.Context, loan Entity) error
    GetByID(ctx context.Context, id string) (Entity, error)
    GetByMemberID(ctx context.Context, memberID string) ([]Entity, error)
    Update(ctx context.Context, loan Entity) error
    Delete(ctx context.Context, id string) error
}
```

### 4. Domain Tests

```go
// internal/domain/loan/service_test.go
package loan

import (
    "testing"
    "time"
)

func TestService_ValidateLoanDuration(t *testing.T) {
    svc := NewService()

    tests := []struct {
        name     string
        duration time.Duration
        wantErr  bool
    }{
        {"valid 7 days", 7 * 24 * time.Hour, false},
        {"valid 14 days", 14 * 24 * time.Hour, false},
        {"invalid zero", 0, true},
        {"invalid negative", -1 * time.Hour, true},
        {"invalid too long", 15 * 24 * time.Hour, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := svc.ValidateLoanDuration(tt.duration)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateLoanDuration() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}

func TestEntity_IsOverdue(t *testing.T) {
    pastDue := time.Now().Add(-24 * time.Hour)
    futureDue := time.Now().Add(24 * time.Hour)

    tests := []struct {
        name   string
        entity Entity
        want   bool
    }{
        {
            name: "overdue loan",
            entity: Entity{
                Status:  StatusActive,
                DueDate: pastDue,
            },
            want: true,
        },
        {
            name: "not overdue",
            entity: Entity{
                Status:  StatusActive,
                DueDate: futureDue,
            },
            want: false,
        },
        {
            name: "returned loan",
            entity: Entity{
                Status:  StatusReturned,
                DueDate: pastDue,
            },
            want: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := tt.entity.IsOverdue(); got != tt.want {
                t.Errorf("IsOverdue() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Adding a New Use Case

```go
// internal/usecase/loanops/create_loan.go
package loanops

import (
    "context"
    "fmt"
    "time"

    "library-service/internal/domain/loan"
    "library-service/internal/infrastructure/log"
    "library-service/pkg/errors"
)

type CreateLoanRequest struct {
    BookID       string
    MemberID     string
    LoanDuration time.Duration
}

type CreateLoanUseCase struct {
    loanRepo    loan.Repository
    loanService *loan.Service
}

func NewCreateLoanUseCase(repo loan.Repository, svc *loan.Service) *CreateLoanUseCase {
    return &CreateLoanUseCase{
        loanRepo:    repo,
        loanService: svc,
    }
}

func (uc *CreateLoanUseCase) Execute(ctx context.Context, req CreateLoanRequest) (*loan.Entity, error) {
    // Validate loan duration using domain service
    if err := uc.loanService.ValidateLoanDuration(req.LoanDuration); err != nil {
        log.Warn("Invalid loan duration", "duration", req.LoanDuration, "error", err)
        return nil, errors.ErrValidation
    }

    // Create loan entity
    newLoan := loan.NewEntity(req.BookID, req.MemberID, req.LoanDuration)

    // Persist to repository
    if err := uc.loanRepo.Create(ctx, newLoan); err != nil {
        log.Error("Failed to create loan", "error", err)
        return nil, fmt.Errorf("creating loan: %w", err)
    }

    log.Info("Loan created successfully", "loan_id", newLoan.ID, "member_id", req.MemberID)
    return &newLoan, nil
}
```

## Adding a New API Endpoint

### 1. DTO (Data Transfer Object)

```go
// internal/adapters/http/dto/loan.go
package dto

import "time"

type CreateLoanRequest struct {
    BookID       string `json:"book_id" validate:"required,uuid"`
    MemberID     string `json:"member_id" validate:"required,uuid"`
    LoanDuration int    `json:"loan_duration_days" validate:"required,min=1,max=14"`
}

type LoanResponse struct {
    ID         string     `json:"id"`
    BookID     string     `json:"book_id"`
    MemberID   string     `json:"member_id"`
    LoanDate   time.Time  `json:"loan_date"`
    DueDate    time.Time  `json:"due_date"`
    ReturnDate *time.Time `json:"return_date,omitempty"`
    Status     string     `json:"status"`
}
```

### 2. HTTP Handler

```go
// internal/adapters/http/handlers/loan.go
package handlers

import (
    "encoding/json"
    "net/http"
    "time"

    "library-service/internal/adapters/http/dto"
    "library-service/internal/usecase/loanops"
    "library-service/pkg/errors"
)

type LoanHandler struct {
    createLoanUC *loanops.CreateLoanUseCase
}

func NewLoanHandler(createLoanUC *loanops.CreateLoanUseCase) *LoanHandler {
    return &LoanHandler{
        createLoanUC: createLoanUC,
    }
}

// CreateLoan creates a new book loan
// @Summary Create a new loan
// @Description Create a new book loan for a member
// @Tags loans
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateLoanRequest true "Loan details"
// @Success 201 {object} dto.LoanResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /loans [post]
func (h *LoanHandler) CreateLoan(w http.ResponseWriter, r *http.Request) {
    var req dto.CreateLoanRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, errors.ErrValidation, http.StatusBadRequest)
        return
    }

    // Validate request
    if err := validate.Struct(req); err != nil {
        respondError(w, err, http.StatusBadRequest)
        return
    }

    // Convert days to duration
    loanDuration := time.Duration(req.LoanDuration) * 24 * time.Hour

    // Execute use case
    loan, err := h.createLoanUC.Execute(r.Context(), loanops.CreateLoanRequest{
        BookID:       req.BookID,
        MemberID:     req.MemberID,
        LoanDuration: loanDuration,
    })
    if err != nil {
        respondError(w, err, http.StatusInternalServerError)
        return
    }

    // Map to response DTO
    response := dto.LoanResponse{
        ID:         loan.ID,
        BookID:     loan.BookID,
        MemberID:   loan.MemberID,
        LoanDate:   loan.LoanDate,
        DueDate:    loan.DueDate,
        ReturnDate: loan.ReturnDate,
        Status:     string(loan.Status),
    }

    respondJSON(w, response, http.StatusCreated)
}
```

### 3. Wire in Container

```go
// internal/usecase/container.go
// Add to Repositories struct:
type Repositories struct {
    // ... existing repos
    Loan loan.Repository  // ADD THIS
}

// Add to Container struct:
type Container struct {
    // ... existing use cases
    CreateLoan *loanops.CreateLoanUseCase  // ADD THIS
}

// Add to NewContainer function:
func NewContainer(repos *Repositories, caches *Caches, authSvcs *AuthServices) *Container {
    // ... existing services
    loanService := loan.NewService()  // ADD THIS

    return &Container{
        // ... existing use cases
        CreateLoan: loanops.NewCreateLoanUseCase(repos.Loan, loanService),  // ADD THIS
    }
}
```

### 4. Add Routes

```go
// internal/adapters/http/router.go
// In setupRoutes function:
func setupRoutes(r *chi.Mux, handlers *Handlers) {
    // ... existing routes

    // Loan routes
    r.Route("/loans", func(r chi.Router) {
        r.Use(authMiddleware)  // Protected routes
        r.Post("/", handlers.Loan.CreateLoan)
        r.Get("/{id}", handlers.Loan.GetLoan)
        r.Post("/{id}/return", handlers.Loan.ReturnLoan)
    })
}
```

## Writing Tests

### Mock Repository (for use case tests)

```go
// internal/domain/loan/mocks/repository.go
package mocks

import (
    "context"
    "library-service/internal/domain/loan"
)

type MockLoanRepository struct {
    CreateFunc func(ctx context.Context, loan loan.Entity) error
    GetByIDFunc func(ctx context.Context, id string) (loan.Entity, error)
}

func (m *MockLoanRepository) Create(ctx context.Context, loan loan.Entity) error {
    if m.CreateFunc != nil {
        return m.CreateFunc(ctx, loan)
    }
    return nil
}

func (m *MockLoanRepository) GetByID(ctx context.Context, id string) (loan.Entity, error) {
    if m.GetByIDFunc != nil {
        return m.GetByIDFunc(ctx, id)
    }
    return loan.Entity{}, nil
}
```

### Use Case Test with Mocks

```go
// internal/usecase/loanops/create_loan_test.go
package loanops

import (
    "context"
    "testing"
    "time"

    "library-service/internal/domain/loan"
    "library-service/internal/domain/loan/mocks"
)

func TestCreateLoanUseCase_Execute(t *testing.T) {
    ctx := context.Background()

    tests := []struct {
        name    string
        req     CreateLoanRequest
        setup   func(*mocks.MockLoanRepository)
        wantErr bool
    }{
        {
            name: "successful loan creation",
            req: CreateLoanRequest{
                BookID:       "book-123",
                MemberID:     "member-456",
                LoanDuration: 7 * 24 * time.Hour,
            },
            setup: func(repo *mocks.MockLoanRepository) {
                repo.CreateFunc = func(ctx context.Context, l loan.Entity) error {
                    return nil
                }
            },
            wantErr: false,
        },
        {
            name: "invalid duration",
            req: CreateLoanRequest{
                BookID:       "book-123",
                MemberID:     "member-456",
                LoanDuration: 0,
            },
            setup:   func(repo *mocks.MockLoanRepository) {},
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo := &mocks.MockLoanRepository{}
            tt.setup(mockRepo)

            svc := loan.NewService()
            uc := NewCreateLoanUseCase(mockRepo, svc)

            _, err := uc.Execute(ctx, tt.req)
            if (err != nil) != tt.wantErr {
                t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## Common Patterns

### Error Handling Pattern

```go
// Always wrap errors with context
if err := repo.Create(ctx, entity); err != nil {
    log.Error("Failed to create entity", "error", err)
    return fmt.Errorf("creating entity: %w", err)  // Use %w for wrapping
}

// Use domain errors for known cases
if entity.ID == "" {
    return errors.ErrValidation
}

// Check wrapped errors
if errors.Is(err, errors.ErrNotFound) {
    // Handle not found
}
```

### Logging Pattern

```go
import "library-service/internal/infrastructure/log"

// Info logging with structured fields
log.Info("Operation successful", "entity_id", id, "user_id", userID)

// Warning for expected errors
log.Warn("Invalid input", "field", "email", "value", email)

// Error for unexpected errors
log.Error("Database operation failed", "error", err, "query", query)

// Debug for development
log.Debug("Processing item", "item", item)
```

### Context Usage Pattern

```go
// Always pass context as first parameter
func (uc *UseCase) Execute(ctx context.Context, req Request) (*Response, error) {
    // Check for cancellation
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
    }

    // Pass context to all downstream calls
    result, err := uc.repo.GetByID(ctx, req.ID)
    if err != nil {
        return nil, err
    }

    return &Response{Data: result}, nil
}
```

### Validation Pattern

```go
// Use validator tags in DTOs
type CreateRequest struct {
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
    Name     string `json:"name" validate:"required,max=100"`
}

// Validate in handler
if err := validate.Struct(req); err != nil {
    respondError(w, err, http.StatusBadRequest)
    return
}
```
