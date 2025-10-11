# Adding an API Endpoint - Quick Guide

This guide shows how to add a new REST API endpoint to an existing domain entity.

**Scenario:** Add a `GET /books/{id}/availability` endpoint to check book availability

**Time Estimate:** 30-45 minutes

---

## Step 1: Create Use Case (15 minutes)

**File:** `internal/usecase/bookops/check_availability.go`

```go
package bookops

import (
	"context"

	"go.uber.org/zap"

	"library-service/internal/domain/book"
	"library-service/internal/domain/reservation"
	"library-service/pkg/errors"
	"library-service/pkg/logutil"
)

// CheckAvailabilityRequest represents the input for checking book availability
type CheckAvailabilityRequest struct {
	BookID string
}

// CheckAvailabilityResponse represents the book availability status
type CheckAvailabilityResponse struct {
	BookID         string
	IsAvailable    bool
	TotalCopies    int
	AvailableCopies int
	ReservedCopies  int
}

// CheckAvailabilityUseCase handles checking book availability
type CheckAvailabilityUseCase struct {
	bookRepo        book.Repository
	reservationRepo reservation.Repository
}

// NewCheckAvailabilityUseCase creates a new instance
func NewCheckAvailabilityUseCase(
	bookRepo book.Repository,
	reservationRepo reservation.Repository,
) *CheckAvailabilityUseCase {
	return &CheckAvailabilityUseCase{
		bookRepo:        bookRepo,
		reservationRepo: reservationRepo,
	}
}

// Execute checks if a book is available for borrowing
func (uc *CheckAvailabilityUseCase) Execute(ctx context.Context, req CheckAvailabilityRequest) (*CheckAvailabilityResponse, error) {
	logger := logutil.UseCaseLogger(ctx, "check_book_availability",
		zap.String("book_id", req.BookID),
	)

	// Verify book exists
	bookEntity, err := uc.bookRepo.Get(ctx, req.BookID)
	if err != nil {
		logger.Warn("book not found", zap.Error(err))
		return nil, errors.ErrNotFound.WithDetails("entity", "book")
	}

	// Get active reservations for this book
	reservations, err := uc.reservationRepo.GetByBookID(ctx, req.BookID)
	if err != nil {
		logger.Error("failed to get reservations", zap.Error(err))
		return nil, err
	}

	// Count active reservations (pending or fulfilled)
	activeReservations := 0
	for _, r := range reservations {
		if r.Status == "pending" || r.Status == "fulfilled" {
			activeReservations++
		}
	}

	// In a real system, books would have a "total_copies" field
	// For this example, we'll assume 1 copy per book
	totalCopies := 1
	availableCopies := totalCopies - activeReservations

	if availableCopies < 0 {
		availableCopies = 0
	}

	isAvailable := availableCopies > 0

	logger.Info("book availability checked",
		zap.Bool("is_available", isAvailable),
		zap.Int("available_copies", availableCopies),
	)

	return &CheckAvailabilityResponse{
		BookID:          bookEntity.ID,
		IsAvailable:     isAvailable,
		TotalCopies:     totalCopies,
		AvailableCopies: availableCopies,
		ReservedCopies:  activeReservations,
	}, nil
}
```

**Key Points:**
- Reuses existing repositories (book, reservation)
- Implements clear business logic
- Returns structured response

---

## Step 2: Create DTO (5 minutes)

**File:** `internal/adapters/http/dto/book.go` (add to existing file)

```go
// Add to existing dto/book.go file:

// BookAvailabilityResponse represents book availability information
type BookAvailabilityResponse struct {
	BookID          string `json:"book_id"`
	IsAvailable     bool   `json:"is_available"`
	TotalCopies     int    `json:"total_copies"`
	AvailableCopies int    `json:"available_copies"`
	ReservedCopies  int    `json:"reserved_copies"`
}

// ToBookAvailabilityResponse converts use case response to DTO
func ToBookAvailabilityResponse(r bookops.CheckAvailabilityResponse) BookAvailabilityResponse {
	return BookAvailabilityResponse{
		BookID:          r.BookID,
		IsAvailable:     r.IsAvailable,
		TotalCopies:     r.TotalCopies,
		AvailableCopies: r.AvailableCopies,
		ReservedCopies:  r.ReservedCopies,
	}
}
```

---

## Step 3: Add HTTP Handler (10 minutes)

**File:** `internal/adapters/http/handlers/book_query.go` (add to existing file)

```go
// Add to existing book_query.go file:

// @Summary Check book availability
// @Description Check if a book is available for borrowing
// @Tags books
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Book ID"
// @Success 200 {object} dto.BookAvailabilityResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /books/{id}/availability [get]
func (h *BookHandler) checkAvailability(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := logutil.HandlerLogger(ctx, "book_handler", "check_availability")

	id, ok := h.GetURLParam(w, r, "id")
	if !ok {
		return
	}

	// Execute use case
	result, err := h.useCases.CheckBookAvailability.Execute(ctx, bookops.CheckAvailabilityRequest{
		BookID: id,
	})
	if err != nil {
		h.RespondError(w, r, err)
		return
	}

	// Convert to DTO
	response := dto.ToBookAvailabilityResponse(*result)

	logger.Info("book availability checked",
		zap.String("book_id", id),
		zap.Bool("is_available", response.IsAvailable),
	)
	h.RespondJSON(w, http.StatusOK, response)
}
```

---

## Step 4: Add Route (5 minutes)

**File:** `internal/adapters/http/handlers/book_handler.go` (update Routes method)

```go
// Update the Routes() method to include the new route:

func (h *BookHandler) Routes() chi.Router {
	r := chi.NewRouter()

	r.Get("/", h.list)
	r.Post("/", h.create)

	r.Route("/{id}", func(r chi.Router) {
		r.Get("/", h.get)
		r.Put("/", h.update)
		r.Delete("/", h.delete)
		r.Get("/authors", h.listAuthors)
		r.Get("/availability", h.checkAvailability)  // ADD THIS LINE
	})

	return r
}
```

---

## Step 5: Wire in Container (5 minutes)

**File:** `internal/usecase/container.go`

```go
// Add to Container struct:
type Container struct {
	// ... existing fields ...

	// Add new use case
	CheckBookAvailability *bookops.CheckAvailabilityUseCase  // ADD THIS
}

// Update NewContainer function:
func NewContainer(repos Repositories, caches Caches, authServices AuthServices) *Container {
	// ... existing code ...

	return &Container{
		// ... existing fields ...

		// Wire new use case
		CheckBookAvailability: bookops.NewCheckAvailabilityUseCase(
			repos.BookRepo,
			repos.ReservationRepo,
		),  // ADD THIS
	}
}
```

---

## Step 6: Test the Endpoint (5 minutes)

### 6.1 Build & Run

```bash
# Build
go build ./...

# Run tests
go test ./internal/usecase/bookops/... -v

# Start server
make run
```

### 6.2 Test with curl

```bash
# Get a valid book ID first
TOKEN="your-jwt-token"
BOOK_ID="some-book-uuid"

# Check availability
curl -X GET "http://localhost:8080/api/v1/books/${BOOK_ID}/availability" \
  -H "Authorization: Bearer ${TOKEN}" \
  -H "Content-Type: application/json"

# Expected response:
{
  "book_id": "uuid-here",
  "is_available": true,
  "total_copies": 1,
  "available_copies": 1,
  "reserved_copies": 0
}
```

### 6.3 Check Swagger Docs

```bash
# Regenerate Swagger documentation
make gen-docs

# View at http://localhost:8080/swagger/index.html
```

---

## Step 7: Add Tests (Optional, 10 minutes)

**File:** `internal/usecase/bookops/check_availability_test.go`

```go
package bookops

import (
	"context"
	"testing"

	"library-service/internal/domain/book"
	"library-service/internal/domain/reservation"
)

func TestCheckAvailabilityUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		bookID          string
		activeReservations int
		wantAvailable   bool
		wantErr         bool
	}{
		{
			name:               "book available",
			bookID:             "book-1",
			activeReservations: 0,
			wantAvailable:      true,
			wantErr:            false,
		},
		{
			name:               "book not available",
			bookID:             "book-1",
			activeReservations: 1,
			wantAvailable:      false,
			wantErr:            false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			// ...

			// Execute
			result, err := uc.Execute(ctx, CheckAvailabilityRequest{
				BookID: tt.bookID,
			})

			// Assert
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result != nil && result.IsAvailable != tt.wantAvailable {
				t.Errorf("IsAvailable = %v, want %v", result.IsAvailable, tt.wantAvailable)
			}
		})
	}
}
```

---

## Summary Checklist

- ✅ **Step 1:** Created use case with business logic
- ✅ **Step 2:** Created DTO for HTTP response
- ✅ **Step 3:** Added HTTP handler method
- ✅ **Step 4:** Registered route in router
- ✅ **Step 5:** Wired use case in container
- ✅ **Step 6:** Tested endpoint with curl
- ✅ **Step 7:** Added unit tests (optional)

**Total Time:** 30-45 minutes

---

## Common Patterns

### Pattern 1: Query Endpoint (GET)
```
Use Case → Handler → Route → Test
No database changes needed
```

### Pattern 2: Command Endpoint (POST/PUT/DELETE)
```
Use Case → Handler → DTO Validation → Route → Test
May need migration if adding fields
```

### Pattern 3: Endpoint with Path Parameters
```go
// Route:
r.Get("/books/{id}/something", h.getSomething)

// Handler:
id, ok := h.GetURLParam(w, r, "id")
```

### Pattern 4: Endpoint with Query Parameters
```go
// Route:
r.Get("/books", h.list)

// Handler:
query := r.URL.Query()
limit := query.Get("limit")  // ?limit=10
genre := query.Get("genre")  // ?genre=fiction
```

### Pattern 5: Authenticated Endpoint
```go
// Handler:
memberID, ok := h.GetMemberID(w, r)
if !ok {
	return // 401 Unauthorized
}
```

---

## Quick Reference

### Swagger Annotations
```go
// @Summary       Short description
// @Description   Longer description
// @Tags          category-name
// @Accept        json
// @Produce       json
// @Security      BearerAuth  // For protected endpoints
// @Param         name location type required "description"
//                ↑    ↑        ↑    ↑        ↑
//                |    path/    |    true/    Description
//                |    query/   Type false
//                |    body
//                Parameter name
// @Success       200 {object} dto.ResponseType
// @Failure       400 {object} dto.ErrorResponse
// @Router        /path [method]
```

### Common Response Codes
```go
http.StatusOK                  // 200 - Success (GET, PUT)
http.StatusCreated             // 201 - Created (POST)
http.StatusNoContent           // 204 - Success, no body (DELETE)
http.StatusBadRequest          // 400 - Validation error
http.StatusUnauthorized        // 401 - Not authenticated
http.StatusForbidden           // 403 - Not authorized
http.StatusNotFound            // 404 - Resource not found
http.StatusConflict            // 409 - Already exists
http.StatusInternalServerError // 500 - Server error
```

---

## Next Steps

After adding the endpoint:

1. **Update Documentation:** Add to API documentation if needed
2. **Update Postman/Insomnia:** Add to API collection
3. **Frontend Integration:** Share endpoint details with frontend team
4. **Monitor:** Add metrics/logging if high-traffic endpoint
5. **Performance:** Consider caching if frequently accessed

**Pattern Established!** Use this same workflow for any new endpoint.
