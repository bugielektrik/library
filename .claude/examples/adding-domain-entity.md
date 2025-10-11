# Adding a Domain Entity - Complete Walkthrough

This guide walks you through adding a complete "Review" domain entity to the Library Management System, following Clean Architecture principles.

**Time Estimate:** 2-3 hours for a complete feature

**What We'll Build:** A book review system where members can rate and review books.

---

## Step 1: Domain Layer (45 minutes)

### 1.1 Create the Entity

**File:** `internal/domain/review/review.go`

```go
package review

import (
	"time"
)

// Review represents a book review by a member
type Review struct {
	ID        string
	BookID    string
	MemberID  string
	Rating    int       // 1-5 stars
	Comment   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Repository defines the interface for review storage operations
type Repository interface {
	Create(ctx context.Context, review Review) (string, error)
	Get(ctx context.Context, id string) (Review, error)
	GetByBookID(ctx context.Context, bookID string) ([]Review, error)
	GetByMemberID(ctx context.Context, memberID string) ([]Review, error)
	Update(ctx context.Context, review Review) error
	Delete(ctx context.Context, id string) error
}
```

**Key Points:**
- Entity contains only domain data (no DB tags, no JSON tags)
- Repository interface defines what operations we need
- Interface lives in domain layer (implementation in adapters)

---

### 1.2 Create the Domain Service

**File:** `internal/domain/review/service.go`

```go
package review

import (
	"errors"
	"strings"
)

// Service provides business logic for reviews
type Service struct{}

// NewService creates a new review service
func NewService() *Service {
	return &Service{}
}

// ValidateRating ensures rating is within valid range
func (s *Service) ValidateRating(rating int) error {
	if rating < 1 || rating > 5 {
		return errors.New("rating must be between 1 and 5 stars")
	}
	return nil
}

// ValidateComment ensures comment meets requirements
func (s *Service) ValidateComment(comment string) error {
	comment = strings.TrimSpace(comment)

	if len(comment) == 0 {
		return errors.New("comment cannot be empty")
	}

	if len(comment) < 10 {
		return errors.New("comment must be at least 10 characters")
	}

	if len(comment) > 5000 {
		return errors.New("comment must not exceed 5000 characters")
	}

	return nil
}

// ValidateReview validates the entire review entity
func (s *Service) ValidateReview(review Review) error {
	if review.BookID == "" {
		return errors.New("book ID is required")
	}

	if review.MemberID == "" {
		return errors.New("member ID is required")
	}

	if err := s.ValidateRating(review.Rating); err != nil {
		return err
	}

	if err := s.ValidateComment(review.Comment); err != nil {
		return err
	}

	return nil
}

// CanMemberReviewBook checks if a member can review a book
// Business rule: Members can only review books they've borrowed
func (s *Service) CanMemberReviewBook(memberID, bookID string, borrowedBooks []string) bool {
	for _, borrowed := range borrowedBooks {
		if borrowed == bookID {
			return true
		}
	}
	return false
}
```

**Key Points:**
- Pure business logic (no database, no HTTP)
- Easy to test (no mocking needed)
- Business rules encoded in code

---

### 1.3 Write Domain Tests

**File:** `internal/domain/review/service_test.go`

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
		{"valid rating 3", 3, false},
		{"valid rating 5", 5, false},
		{"invalid rating 0", 0, true},
		{"invalid rating 6", 6, true},
		{"invalid rating negative", -1, true},
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

func TestService_ValidateComment(t *testing.T) {
	service := NewService()

	tests := []struct {
		name    string
		comment string
		wantErr bool
	}{
		{
			name:    "valid comment",
			comment: "This is a great book with amazing content!",
			wantErr: false,
		},
		{
			name:    "empty comment",
			comment: "",
			wantErr: true,
		},
		{
			name:    "too short comment",
			comment: "Good",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			comment: "   ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateComment(tt.comment)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateComment() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_CanMemberReviewBook(t *testing.T) {
	service := NewService()

	tests := []struct {
		name          string
		memberID      string
		bookID        string
		borrowedBooks []string
		want          bool
	}{
		{
			name:          "member borrowed the book",
			memberID:      "member-1",
			bookID:        "book-1",
			borrowedBooks: []string{"book-1", "book-2"},
			want:          true,
		},
		{
			name:          "member did not borrow the book",
			memberID:      "member-1",
			bookID:        "book-3",
			borrowedBooks: []string{"book-1", "book-2"},
			want:          false,
		},
		{
			name:          "no borrowed books",
			memberID:      "member-1",
			bookID:        "book-1",
			borrowedBooks: []string{},
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.CanMemberReviewBook(tt.memberID, tt.bookID, tt.borrowedBooks)
			if got != tt.want {
				t.Errorf("CanMemberReviewBook() = %v, want %v", got, tt.want)
			}
		})
	}
}
```

**Run tests:**
```bash
go test ./internal/domain/review/... -v
```

**Expected:** 100% coverage on domain service

---

### 1.4 Add Package Documentation

**File:** `internal/domain/review/doc.go`

```go
// Package review provides the domain model and business logic for book reviews.
//
// The review domain allows library members to rate and review books they have
// borrowed. Each review consists of a 1-5 star rating and a text comment.
//
// Business Rules:
//   - Members can only review books they have previously borrowed
//   - Ratings must be between 1 and 5 stars (inclusive)
//   - Comments must be between 10 and 5000 characters
//   - Each member can only review a book once
//
// The domain service (Service) encodes these business rules and validates
// review entities before they are persisted.
package review
```

---

## Step 2: Use Case Layer (45 minutes)

### 2.1 Create Use Case - Create Review

**File:** `internal/usecase/reviewops/create_review.go`

```go
package reviewops

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"library-service/internal/domain/book"
	"library-service/internal/domain/member"
	"library-service/internal/domain/review"
	"library-service/pkg/errors"
	"library-service/pkg/logutil"
)

// CreateReviewRequest represents the input for creating a review
type CreateReviewRequest struct {
	BookID   string
	MemberID string
	Rating   int
	Comment  string
}

// CreateReviewResponse represents the output of creating a review
type CreateReviewResponse struct {
	ReviewID  string
	BookID    string
	Rating    int
	CreatedAt time.Time
}

// CreateReviewUseCase handles creating a new book review
type CreateReviewUseCase struct {
	reviewRepo  review.Repository
	bookRepo    book.Repository
	memberRepo  member.Repository
	reviewService *review.Service
}

// NewCreateReviewUseCase creates a new instance
func NewCreateReviewUseCase(
	reviewRepo review.Repository,
	bookRepo book.Repository,
	memberRepo member.Repository,
	reviewService *review.Service,
) *CreateReviewUseCase {
	return &CreateReviewUseCase{
		reviewRepo:    reviewRepo,
		bookRepo:      bookRepo,
		memberRepo:    memberRepo,
		reviewService: reviewService,
	}
}

// Execute creates a new review
func (uc *CreateReviewUseCase) Execute(ctx context.Context, req CreateReviewRequest) (*CreateReviewResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "create_review",
		zap.String("book_id", req.BookID),
		zap.String("member_id", req.MemberID),
		zap.Int("rating", req.Rating),
	)

	// Verify book exists
	_, err := uc.bookRepo.Get(ctx, req.BookID)
	if err != nil {
		logger.Warn("book not found", zap.Error(err))
		return nil, errors.ErrNotFound.WithDetails("entity", "book")
	}

	// Verify member exists and get borrowed books
	memberEntity, err := uc.memberRepo.Get(ctx, req.MemberID)
	if err != nil {
		logger.Warn("member not found", zap.Error(err))
		return nil, errors.ErrNotFound.WithDetails("entity", "member")
	}

	// Business rule: Can only review borrowed books
	if !uc.reviewService.CanMemberReviewBook(req.MemberID, req.BookID, memberEntity.Books) {
		logger.Warn("member has not borrowed this book")
		return nil, errors.ErrInvalidInput.WithDetails("reason", "you can only review books you have borrowed")
	}

	// Check if review already exists
	existingReviews, err := uc.reviewRepo.GetByMemberID(ctx, req.MemberID)
	if err == nil {
		for _, r := range existingReviews {
			if r.BookID == req.BookID {
				logger.Warn("review already exists")
				return nil, errors.ErrAlreadyExists.WithDetails("entity", "review")
			}
		}
	}

	// Create review entity
	now := time.Now()
	newReview := review.Review{
		BookID:    req.BookID,
		MemberID:  req.MemberID,
		Rating:    req.Rating,
		Comment:   req.Comment,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Validate using domain service
	if err := uc.reviewService.ValidateReview(newReview); err != nil {
		logger.Warn("review validation failed", zap.Error(err))
		return nil, errors.ErrInvalidInput.WithDetails("error", err.Error())
	}

	// Save to repository
	reviewID, err := uc.reviewRepo.Create(ctx, newReview)
	if err != nil {
		logger.Error("failed to create review", zap.Error(err))
		return nil, fmt.Errorf("failed to create review: %w", err)
	}

	logger.Info("review created successfully", zap.String("review_id", reviewID))

	return &CreateReviewResponse{
		ReviewID:  reviewID,
		BookID:    req.BookID,
		Rating:    req.Rating,
		CreatedAt: now,
	}, nil
}
```

**Key Points:**
- Orchestrates multiple repositories (book, member, review)
- Enforces business rules using domain service
- Comprehensive logging at key points
- Returns domain-level responses (not DTOs)

---

### 2.2 Add Package Documentation

**File:** `internal/usecase/reviewops/doc.go`

```go
// Package reviewops provides use cases for review management operations.
//
// This package orchestrates review-related business logic by coordinating
// domain services and repositories. Use cases include:
//   - Creating reviews for books
//   - Updating existing reviews
//   - Deleting reviews
//   - Querying reviews by book or member
//
// Use cases enforce business rules and handle cross-entity operations.
package reviewops
```

---

## Step 3: Repository Implementation (30 minutes)

### 3.1 PostgreSQL Repository

**File:** `internal/adapters/repository/postgres/review.go`

```go
package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"library-service/internal/domain/review"
	"library-service/internal/infrastructure/store"
)

// ReviewRepository handles review operations with PostgreSQL
type ReviewRepository struct {
	BaseRepository[review.Review]
}

// NewReviewRepository creates a new review repository
func NewReviewRepository(db *sqlx.DB) *ReviewRepository {
	return &ReviewRepository{
		BaseRepository: NewBaseRepository[review.Review](db, "reviews"),
	}
}

// Create inserts a new review
func (r *ReviewRepository) Create(ctx context.Context, data review.Review) (string, error) {
	query := `
		INSERT INTO reviews (book_id, member_id, rating, comment, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	args := []interface{}{
		data.BookID,
		data.MemberID,
		data.Rating,
		data.Comment,
		data.CreatedAt,
		data.UpdatedAt,
	}

	var id string
	err := r.GetDB().QueryRowContext(ctx, query, args...).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("creating review: %w", HandleSQLError(err))
	}

	return id, nil
}

// GetByBookID retrieves all reviews for a book
func (r *ReviewRepository) GetByBookID(ctx context.Context, bookID string) ([]review.Review, error) {
	query := `
		SELECT id, book_id, member_id, rating, comment, created_at, updated_at
		FROM reviews
		WHERE book_id = $1
		ORDER BY created_at DESC
	`

	var reviews []review.Review
	err := r.GetDB().SelectContext(ctx, &reviews, query, bookID)
	if err != nil {
		return nil, fmt.Errorf("getting reviews by book ID: %w", err)
	}

	return reviews, nil
}

// GetByMemberID retrieves all reviews by a member
func (r *ReviewRepository) GetByMemberID(ctx context.Context, memberID string) ([]review.Review, error) {
	query := `
		SELECT id, book_id, member_id, rating, comment, created_at, updated_at
		FROM reviews
		WHERE member_id = $1
		ORDER BY created_at DESC
	`

	var reviews []review.Review
	err := r.GetDB().SelectContext(ctx, &reviews, query, memberID)
	if err != nil {
		return nil, fmt.Errorf("getting reviews by member ID: %w", err)
	}

	return reviews, nil
}

// Update modifies an existing review
func (r *ReviewRepository) Update(ctx context.Context, data review.Review) error {
	query := `
		UPDATE reviews
		SET rating = $1, comment = $2, updated_at = $3
		WHERE id = $4
		RETURNING id
	`

	args := []interface{}{
		data.Rating,
		data.Comment,
		data.UpdatedAt,
		data.ID,
	}

	var id string
	err := r.GetDB().QueryRowContext(ctx, query, args...).Scan(&id)
	return HandleSQLError(err)
}
```

---

## Step 4: Database Migration (15 minutes)

### 4.1 Create Migration

```bash
make migrate-create name=create_reviews_table
```

### 4.2 Write Migration Up

**File:** `migrations/postgres/NNNNNN_create_reviews_table.up.sql`

```sql
CREATE TABLE IF NOT EXISTS reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    book_id UUID NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    member_id UUID NOT NULL REFERENCES members(id) ON DELETE CASCADE,
    rating INTEGER NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment TEXT NOT NULL CHECK (LENGTH(comment) >= 10 AND LENGTH(comment) <= 5000),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Ensure one review per member per book
    UNIQUE(book_id, member_id)
);

-- Indexes for common queries
CREATE INDEX idx_reviews_book_id ON reviews(book_id);
CREATE INDEX idx_reviews_member_id ON reviews(member_id);
CREATE INDEX idx_reviews_rating ON reviews(rating);
CREATE INDEX idx_reviews_created_at ON reviews(created_at DESC);
```

### 4.3 Write Migration Down

**File:** `migrations/postgres/NNNNNN_create_reviews_table.down.sql`

```sql
DROP INDEX IF EXISTS idx_reviews_created_at;
DROP INDEX IF EXISTS idx_reviews_rating;
DROP INDEX IF EXISTS idx_reviews_member_id;
DROP INDEX IF EXISTS idx_reviews_book_id;
DROP TABLE IF EXISTS reviews;
```

### 4.4 Run Migration

```bash
make migrate-up
```

---

## Step 5: HTTP Layer (45 minutes)

### 5.1 Create DTOs

**File:** `internal/adapters/http/dto/review.go`

```go
package dto

// CreateReviewRequest represents the HTTP request to create a review
type CreateReviewRequest struct {
	BookID  string `json:"book_id" validate:"required,uuid"`
	Rating  int    `json:"rating" validate:"required,min=1,max=5"`
	Comment string `json:"comment" validate:"required,min=10,max=5000"`
}

// UpdateReviewRequest represents the HTTP request to update a review
type UpdateReviewRequest struct {
	Rating  int    `json:"rating" validate:"required,min=1,max=5"`
	Comment string `json:"comment" validate:"required,min=10,max=5000"`
}

// ReviewResponse represents a review in HTTP responses
type ReviewResponse struct {
	ID        string `json:"id"`
	BookID    string `json:"book_id"`
	MemberID  string `json:"member_id"`
	Rating    int    `json:"rating"`
	Comment   string `json:"comment"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// ListReviewsResponse represents the response for listing reviews
type ListReviewsResponse struct {
	Reviews []ReviewResponse `json:"reviews"`
	Total   int              `json:"total"`
}
```

### 5.2 Create HTTP Handler

**File:** `internal/adapters/http/handlers/review.go`

```go
package v1

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"library-service/internal/adapters/http/dto"
	"library-service/internal/adapters/http/middleware"
	"library-service/internal/usecase"
	"library-service/internal/usecase/reviewops"
	"library-service/pkg/httputil"
	"library-service/pkg/logutil"
)

// ReviewHandler handles HTTP requests for reviews
type ReviewHandler struct {
	BaseHandler
	useCases  *usecase.Container
	validator *middleware.Validator
}

// NewReviewHandler creates a new review handler
func NewReviewHandler(
	useCases *usecase.Container,
	validator *middleware.Validator,
) *ReviewHandler {
	return &ReviewHandler{
		useCases:  useCases,
		validator: validator,
	}
}

// Routes returns the router for review endpoints
func (h *ReviewHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Post("/", h.createReview)
	r.Get("/books/{bookId}", h.listBookReviews)
	r.Get("/my", h.listMyReviews)

	return r
}

// @Summary Create a review
// @Tags reviews
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateReviewRequest true "Review data"
// @Success 201 {object} dto.ReviewResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /reviews [post]
func (h *ReviewHandler) createReview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logutil.HandlerLogger(ctx, "review_handler", "create")

	// Get member ID from context
	memberID, ok := h.GetMemberID(w, r)
	if !ok {
		return
	}

	// Decode request
	var req dto.CreateReviewRequest
	if err := httputil.DecodeJSON(r, &req); err != nil {
		h.RespondError(w, r, err)
		return
	}

	// Validate request
	if !h.validator.ValidateStruct(w, req) {
		return
	}

	// Execute use case
	result, err := h.useCases.CreateReview.Execute(ctx, reviewops.CreateReviewRequest{
		BookID:   req.BookID,
		MemberID: memberID,
		Rating:   req.Rating,
		Comment:  req.Comment,
	})
	if err != nil {
		h.RespondError(w, r, err)
		return
	}

	response := dto.ReviewResponse{
		ID:        result.ReviewID,
		BookID:    result.BookID,
		MemberID:  memberID,
		Rating:    result.Rating,
		CreatedAt: result.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	logger.Info("review created", zap.String("review_id", result.ReviewID))
	h.RespondJSON(w, http.StatusCreated, response)
}
```

---

## Step 6: Wire Dependencies (15 minutes)

### 6.1 Add to Container

**File:** `internal/usecase/container.go`

```go
// Add to Repositories struct:
type Repositories struct {
	// ... existing repos ...
	ReviewRepo review.Repository  // ADD THIS
}

// Add to Caches struct (if needed):
type Caches struct {
	// ... existing caches ...
	ReviewCache review.Cache  // OPTIONAL
}

// Add to Container struct:
type Container struct {
	// ... existing use cases ...

	// Review use cases
	CreateReview *reviewops.CreateReviewUseCase  // ADD THIS
}

// Update NewContainer function:
func NewContainer(repos Repositories, caches Caches, authServices AuthServices) *Container {
	// ... existing services ...

	// Create review service
	reviewService := review.NewService()  // ADD THIS

	return &Container{
		// ... existing use cases ...

		// Wire review use cases
		CreateReview: reviewops.NewCreateReviewUseCase(
			repos.ReviewRepo,
			repos.BookRepo,
			repos.MemberRepo,
			reviewService,
		),  // ADD THIS
	}
}
```

### 6.2 Wire in Application Bootstrap

**File:** `internal/infrastructure/app/app.go`

```go
// In the Run() function, add:

// Create review repository
reviewRepo := postgres.NewReviewRepository(db)  // ADD THIS

// Wire repositories
repos := usecase.Repositories{
	// ... existing repos ...
	ReviewRepo: reviewRepo,  // ADD THIS
}

// Register review routes
router.Mount("/api/v1/reviews", v1.NewReviewHandler(useCases, validator).Routes())  // ADD THIS
```

---

## Step 7: Testing (30 minutes)

### 7.1 Write Use Case Tests

**File:** `internal/usecase/reviewops/create_review_test.go`

```go
package reviewops

import (
	"context"
	"testing"
	"time"

	"library-service/internal/domain/book"
	"library-service/internal/domain/member"
	"library-service/internal/domain/review"
)

// Mock repositories for testing
type mockReviewRepo struct {
	createFunc         func(ctx context.Context, data review.Review) (string, error)
	getByMemberIDFunc  func(ctx context.Context, memberID string) ([]review.Review, error)
}

func (m *mockReviewRepo) Create(ctx context.Context, data review.Review) (string, error) {
	return m.createFunc(ctx, data)
}

func (m *mockReviewRepo) GetByMemberID(ctx context.Context, memberID string) ([]review.Review, error) {
	return m.getByMemberIDFunc(ctx, memberID)
}

// ... implement other interface methods ...

func TestCreateReviewUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		request CreateReviewRequest
		wantErr bool
	}{
		{
			name: "successful review creation",
			request: CreateReviewRequest{
				BookID:   "book-123",
				MemberID: "member-123",
				Rating:   5,
				Comment:  "This is an excellent book with great content!",
			},
			wantErr: false,
		},
		// Add more test cases...
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks and run test
			// ...
		})
	}
}
```

### 7.2 Manual API Testing

```bash
# Start the server
make run

# Create a review
curl -X POST http://localhost:8080/api/v1/reviews \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "book_id": "some-book-uuid",
    "rating": 5,
    "comment": "Amazing book! Highly recommend it to everyone."
  }'

# List reviews for a book
curl -X GET http://localhost:8080/api/v1/reviews/books/{book-id} \
  -H "Authorization: Bearer YOUR_TOKEN"
```

---

## Summary

You've successfully added a complete domain entity! Here's what you created:

**Domain Layer:**
- ✅ Entity definition (review.go)
- ✅ Domain service (service.go)
- ✅ Repository interface
- ✅ Tests with 100% coverage
- ✅ Package documentation

**Use Case Layer:**
- ✅ CreateReview use case
- ✅ Business logic orchestration
- ✅ Package documentation

**Repository Layer:**
- ✅ PostgreSQL implementation
- ✅ Database migration

**HTTP Layer:**
- ✅ DTOs for requests/responses
- ✅ HTTP handlers with Swagger docs

**Wiring:**
- ✅ Dependency injection in container
- ✅ Routes registered in app

**Total Time:** ~2.5 hours for a complete, production-ready feature

---

## Next Steps

1. **Add more use cases:** UpdateReview, DeleteReview, GetReview
2. **Add more queries:** GetAverageRating, GetReviewStats
3. **Add caching:** Cache book reviews for performance
4. **Add notifications:** Email member when their review is liked
5. **Add moderation:** Admin endpoints to moderate reviews

**Pattern established!** Follow this same workflow for any new domain entity.
