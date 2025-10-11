# Common Tasks

> **Step-by-step guides for frequent development operations**

## Task Index

- [Adding a New Domain Entity](#adding-a-new-domain-entity)
- [Adding an API Endpoint](#adding-an-api-endpoint)
- [Creating a Database Migration](#creating-a-database-migration)
- [Writing Tests](#writing-tests)
- [Adding a New Use Case](#adding-a-new-use-case)
- [Debugging an Issue](#debugging-an-issue)
- [Running Specific Tests](#running-specific-tests)
- [Updating Swagger Documentation](#updating-swagger-documentation)

---

## Adding a New Domain Entity

**Scenario**: You need to add a "Review" feature for book reviews.

### Step 1: Create Domain Layer (20-30 minutes)

```bash
# 1. Create directory
mkdir -p internal/domain/review

# 2. Create files
touch internal/domain/review/entity.go
touch internal/domain/review/service.go
touch internal/domain/review/repository.go
touch internal/domain/review/entity_test.go
touch internal/domain/review/service_test.go
```

### Step 2: Define Entity

**File**: `internal/domain/review/entity.go`

```go
package review

import "time"

type Entity struct {
    ID        string
    BookID    string
    MemberID  string
    Rating    int       // 1-5 stars
    Comment   string
    CreatedAt time.Time
    UpdatedAt time.Time
}

// NewEntity creates a new review
func NewEntity(bookID, memberID string, rating int, comment string) Entity {
    now := time.Now()
    return Entity{
        BookID:    bookID,
        MemberID:  memberID,
        Rating:    rating,
        Comment:   comment,
        CreatedAt: now,
        UpdatedAt: now,
    }
}
```

### Step 3: Create Domain Service

**File**: `internal/domain/review/service.go`

```go
package review

import "library-service/pkg/errors"

// Service handles review business rules
type Service struct{}

// NewService creates a review service
func NewService() *Service {
    return &Service{}
}

// ValidateRating ensures rating is between 1-5
func (s *Service) ValidateRating(rating int) error {
    if rating < 1 || rating > 5 {
        return errors.ErrValidation.WithDetails("rating", "must be between 1 and 5")
    }
    return nil
}

// ValidateReview ensures review data is valid
func (s *Service) ValidateReview(review Entity) error {
    if review.BookID == "" {
        return errors.ErrValidation.WithDetails("book_id", "cannot be empty")
    }
    if review.MemberID == "" {
        return errors.ErrValidation.WithDetails("member_id", "cannot be empty")
    }
    if err := s.ValidateRating(review.Rating); err != nil {
        return err
    }
    return nil
}
```

### Step 4: Define Repository Interface

**File**: `internal/domain/review/repository.go`

```go
package review

import "context"

// Repository defines review data access operations
type Repository interface {
    Create(ctx context.Context, review Entity) (string, error)
    GetByID(ctx context.Context, id string) (Entity, error)
    ListByBookID(ctx context.Context, bookID string) ([]Entity, error)
    ListByMemberID(ctx context.Context, memberID string) ([]Entity, error)
    Update(ctx context.Context, id string, review Entity) error
    Delete(ctx context.Context, id string) error
}
```

### Step 5: Write Domain Tests

**File**: `internal/domain/review/service_test.go`

```go
package review

import (
    "testing"
)

func TestService_ValidateRating(t *testing.T) {
    service := NewService()

    tests := []struct {
        name    string
        rating  int
        wantErr bool
    }{
        {"valid rating 1", 1, false},
        {"valid rating 5", 5, false},
        {"invalid rating 0", 0, true},
        {"invalid rating 6", 6, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := service.ValidateRating(tt.rating)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateRating() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

**Verify**: `go test -v ./internal/domain/review/`

---

## Adding an API Endpoint

**Scenario**: Add POST /api/v1/reviews endpoint to create reviews.

### Step 1: Create Use Case

```bash
mkdir -p internal/usecase/reviewops
touch internal/usecase/reviewops/create_review.go
```

**File**: `internal/usecase/reviewops/create_review.go`

```go
package reviewops

import (
    "context"
    "github.com/google/uuid"
    "library-service/internal/domain/review"
    "library-service/pkg/errors"
)

type CreateReviewRequest struct {
    BookID   string
    MemberID string
    Rating   int
    Comment  string
}

type CreateReviewResponse struct {
    ReviewID string
    Review   review.Entity
}

type CreateReviewUseCase struct {
    reviewRepo    review.Repository
    reviewService *review.Service
}

func NewCreateReviewUseCase(
    reviewRepo review.Repository,
    reviewService *review.Service,
) *CreateReviewUseCase {
    return &CreateReviewUseCase{
        reviewRepo:    reviewRepo,
        reviewService: reviewService,
    }
}

func (uc *CreateReviewUseCase) Execute(
    ctx context.Context,
    req CreateReviewRequest,
) (CreateReviewResponse, error) {
    // Create entity
    reviewEntity := review.NewEntity(req.BookID, req.MemberID, req.Rating, req.Comment)

    // Validate
    if err := uc.reviewService.ValidateReview(reviewEntity); err != nil {
        return CreateReviewResponse{}, err
    }

    // Generate ID
    reviewEntity.ID = uuid.New().String()

    // Persist
    id, err := uc.reviewRepo.Create(ctx, reviewEntity)
    if err != nil {
        return CreateReviewResponse{}, errors.ErrInternal.Wrap(err)
    }

    return CreateReviewResponse{
        ReviewID: id,
        Review:   reviewEntity,
    }, nil
}
```

### Step 2: Create DTO

**File**: `internal/adapters/http/dto/review.go`

```go
package dto

import "library-service/internal/domain/review"

type CreateReviewRequest struct {
    BookID  string `json:"book_id" validate:"required"`
    Rating  int    `json:"rating" validate:"required,min=1,max=5"`
    Comment string `json:"comment"`
}

type ReviewResponse struct {
    ID        string `json:"id"`
    BookID    string `json:"book_id"`
    MemberID  string `json:"member_id"`
    Rating    int    `json:"rating"`
    Comment   string `json:"comment"`
    CreatedAt string `json:"created_at"`
}

func FromReviewEntity(entity review.Entity) ReviewResponse {
    return ReviewResponse{
        ID:        entity.ID,
        BookID:    entity.BookID,
        MemberID:  entity.MemberID,
        Rating:    entity.Rating,
        Comment:   entity.Comment,
        CreatedAt: entity.CreatedAt.Format("2006-01-02T15:04:05Z"),
    }
}
```

### Step 3: Create HTTP Handler

**File**: `internal/adapters/http/handlers/review.go`

```go
package v1

import (
    "encoding/json"
    "net/http"

    "github.com/go-chi/chi/v5"
    "go.uber.org/zap"

    "library-service/internal/adapters/http/dto"
    "library-service/internal/adapters/http/middleware"
    "library-service/internal/infrastructure/log"
    "library-service/internal/usecase/reviewops"
    "library-service/pkg/errors"
)

type ReviewHandler struct {
    createReviewUC *reviewops.CreateReviewUseCase
    validator      *middleware.Validator
}

func NewReviewHandler(createReviewUC *reviewops.CreateReviewUseCase) *ReviewHandler {
    return &ReviewHandler{
        createReviewUC: createReviewUC,
        validator:      middleware.NewValidator(),
    }
}

func (h *ReviewHandler) Routes() chi.Router {
    r := chi.NewRouter()
    r.Post("/", h.createReview)
    return r
}

// @Summary Create review
// @Description Create a new book review
// @Tags reviews
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateReviewRequest true "Review details"
// @Success 201 {object} dto.ReviewResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /reviews [post]
func (h *ReviewHandler) createReview(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    logger := log.FromContext(ctx).Named("review_handler.create")

    // Get member ID from JWT
    memberID, ok := ctx.Value("member_id").(string)
    if !ok || memberID == "" {
        h.respondError(w, r, errors.ErrUnauthorized)
        return
    }

    // Decode request
    var req dto.CreateReviewRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.respondError(w, r, errors.ErrInvalidInput.Wrap(err))
        return
    }

    // Validate
    if !h.validator.ValidateStruct(w, req) {
        return
    }

    // Execute use case
    result, err := h.createReviewUC.Execute(ctx, reviewops.CreateReviewRequest{
        BookID:   req.BookID,
        MemberID: memberID,
        Rating:   req.Rating,
        Comment:  req.Comment,
    })
    if err != nil {
        h.respondError(w, r, err)
        return
    }

    logger.Info("review created", zap.String("review_id", result.ReviewID))
    h.respondJSON(w, http.StatusCreated, dto.FromReviewEntity(result.Review))
}

func (h *ReviewHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}

func (h *ReviewHandler) respondError(w http.ResponseWriter, r *http.Request, err error) {
    logger := log.FromContext(r.Context())
    status := errors.GetHTTPStatus(err)

    if status >= 500 {
        logger.Error("internal error", zap.Error(err))
    } else {
        logger.Warn("client error", zap.Error(err), zap.Int("status", status))
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(dto.FromError(err))
}
```

### Step 4: Implement Repository

**File**: `internal/adapters/repository/postgres/review.go`

```go
package postgres

import (
    "context"
    "github.com/jmoiron/sqlx"
    "library-service/internal/domain/review"
)

type ReviewRepository struct {
    db *sqlx.DB
}

func NewReviewRepository(db *sqlx.DB) *ReviewRepository {
    return &ReviewRepository{db: db}
}

func (r *ReviewRepository) Create(ctx context.Context, rev review.Entity) (string, error) {
    query := `
        INSERT INTO reviews (id, book_id, member_id, rating, comment, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id
    `
    var id string
    err := r.db.QueryRowContext(
        ctx, query,
        rev.ID, rev.BookID, rev.MemberID, rev.Rating, rev.Comment,
        rev.CreatedAt, rev.UpdatedAt,
    ).Scan(&id)

    return id, err
}

// ... implement other methods
```

### Step 5: Wire in Container

**File**: `internal/usecase/container.go`

```go
// Add to Repositories struct
type Repositories struct {
    // ... existing
    Review review.Repository  // ADD THIS
}

// Add to Container struct
type Container struct {
    // ... existing
    CreateReview *reviewops.CreateReviewUseCase  // ADD THIS
}

// In NewContainer function
func NewContainer(repos *Repositories, ...) *Container {
    // Create domain services
    reviewService := review.NewService()  // ADD THIS

    return &Container{
        // ... existing
        CreateReview: reviewops.NewCreateReviewUseCase(repos.Review, reviewService),  // ADD THIS
    }
}
```

### Step 6: Update App Bootstrap

**File**: `internal/infrastructure/app/app.go`

```go
// In setupRepositories function
repos := &usecase.Repositories{
    // ... existing
    Review: postgres.NewReviewRepository(db),  // ADD THIS
}
```

### Step 7: Add Routes

**File**: `internal/adapters/http/router.go`

```go
// Create handler
reviewHandler := v1.NewReviewHandler(cfg.Usecases.CreateReview)

// Add routes (protected by auth middleware)
r.Group(func(r chi.Router) {
    r.Use(authMiddleware.Authenticate)
    r.Mount("/reviews", reviewHandler.Routes())  // ADD THIS
})
```

### Step 8: Regenerate Swagger

```bash
make gen-docs
```

### Step 9: Test

```bash
# Start server
make run

# Test endpoint
curl -X POST http://localhost:8080/api/v1/reviews \
  -H "Authorization: Bearer <your_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "book_id": "123",
    "rating": 5,
    "comment": "Great book!"
  }'
```

---

## Creating a Database Migration

### Step 1: Create Migration Files

```bash
make migrate-create name=create_reviews_table
```

This creates:
- `migrations/postgres/NNNNNN_create_reviews_table.up.sql`
- `migrations/postgres/NNNNNN_create_reviews_table.down.sql`

### Step 2: Write UP Migration

**File**: `migrations/postgres/NNNNNN_create_reviews_table.up.sql`

```sql
CREATE TABLE IF NOT EXISTS reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    book_id UUID NOT NULL,
    member_id UUID NOT NULL,
    rating INT NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_book FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE,
    CONSTRAINT fk_member FOREIGN KEY (member_id) REFERENCES members(id) ON DELETE CASCADE
);

CREATE INDEX idx_reviews_book_id ON reviews(book_id);
CREATE INDEX idx_reviews_member_id ON reviews(member_id);
CREATE INDEX idx_reviews_rating ON reviews(rating);
```

### Step 3: Write DOWN Migration

**File**: `migrations/postgres/NNNNNN_create_reviews_table.down.sql`

```sql
DROP INDEX IF EXISTS idx_reviews_rating;
DROP INDEX IF EXISTS idx_reviews_member_id;
DROP INDEX IF EXISTS idx_reviews_book_id;
DROP TABLE IF EXISTS reviews;
```

### Step 4: Apply Migration

```bash
make migrate-up
```

### Step 5: Verify

```bash
psql "postgres://library:library123@localhost:5432/library" -c "\dt reviews"
```

---

## Writing Tests

### Unit Test (Domain Service)

**File**: `internal/domain/review/service_test.go`

```go
func TestService_ValidateReview(t *testing.T) {
    service := NewService()

    tests := []struct {
        name    string
        review  Entity
        wantErr bool
    }{
        {
            name: "valid review",
            review: Entity{
                BookID:   "book-123",
                MemberID: "member-456",
                Rating:   5,
                Comment:  "Great book!",
            },
            wantErr: false,
        },
        {
            name: "missing book ID",
            review: Entity{
                MemberID: "member-456",
                Rating:   5,
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := service.ValidateReview(tt.review)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateReview() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

**Run**: `go test -v ./internal/domain/review/`

### Use Case Test (with Mocks)

**File**: `internal/usecase/reviewops/create_review_test.go`

```go
func TestCreateReviewUseCase_Execute(t *testing.T) {
    mockRepo := mocks.NewMockReviewRepository(t)
    service := review.NewService()
    uc := NewCreateReviewUseCase(mockRepo, service)

    req := CreateReviewRequest{
        BookID:   "book-123",
        MemberID: "member-456",
        Rating:   5,
        Comment:  "Great!",
    }

    mockRepo.On("Create", mock.Anything, mock.AnythingOfType("review.Entity")).
        Return("review-789", nil)

    result, err := uc.Execute(context.Background(), req)

    require.NoError(t, err)
    assert.Equal(t, "review-789", result.ReviewID)
    mockRepo.AssertExpectations(t)
}
```

---

## Debugging an Issue

### Step 1: Enable Debug Logging

```bash
APP_MODE=dev LOG_LEVEL=debug make run
```

### Step 2: Check Logs

Look for:
- Error messages with stack traces
- Zap log entries with context (request_id, user_id, etc.)

### Step 3: Add Temporary Logging

```go
logger.Debug("debugging payment flow",
    zap.String("payment_id", paymentID),
    zap.Any("request", req),
)
```

### Step 4: Use Debugger

```bash
# Install delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Start debugger
dlv debug ./cmd/api/main.go

# Set breakpoint
(dlv) break internal/usecase/reviewops/create_review.go:45
(dlv) continue
```

---

## Running Specific Tests

```bash
# Run all tests in package
go test -v ./internal/domain/review/

# Run specific test function
go test -v -run TestService_ValidateRating ./internal/domain/review/

# Run with coverage
go test -v -cover ./internal/domain/review/

# Run with race detection
go test -v -race ./internal/domain/review/

# Run integration tests only
make test-integration
```

---

## Updating Swagger Documentation

### After Adding/Changing Endpoints

```bash
# Regenerate documentation
make gen-docs

# Restart server
make run

# View at http://localhost:8080/swagger/index.html
```

### Important Swagger Annotations

```go
// @Summary Short description
// @Description Longer description
// @Tags review
// @Accept json
// @Produce json
// @Security BearerAuth  // For protected endpoints
// @Param id path string true "Review ID"
// @Param request body dto.CreateReviewRequest true "Review details"
// @Success 200 {object} dto.ReviewResponse
// @Failure 400 {object} dto.ErrorResponse
// @Router /reviews/{id} [get]
```

---

## Summary Checklist

When adding a new feature, complete in this order:

- [ ] **Domain Layer**: Entity → Service → Repository interface → Tests
- [ ] **Use Case Layer**: Create use case → Tests
- [ ] **Adapter Layer**: Repository implementation → HTTP handler → DTOs
- [ ] **Wiring**: Update container.go → Update app.go → Update router.go
- [ ] **Migration**: Create migration → Apply migration
- [ ] **Documentation**: Add Swagger annotations → Regenerate docs
- [ ] **Testing**: Write unit tests → Write integration tests → Manual testing
- [ ] **CI**: Run `make ci` to ensure everything passes
