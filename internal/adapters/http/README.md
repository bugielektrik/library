# HTTP Adapter Layer

HTTP interface implementation including handlers, middleware, and routing.

## Structure

```
http/
├── handlers/    # HTTP request handlers organized by domain
│   ├── auth/    # Authentication endpoints
│   ├── book/    # Book management endpoints
│   ├── member/  # Member management endpoints
│   ├── payment/ # Payment processing endpoints
│   └── base.go  # Base handler with common methods
├── middleware/  # HTTP middleware (auth, logging, CORS)
├── dto/         # Data Transfer Objects for requests/responses
├── router.go    # Route configuration
└── validator.go # Request validation
```

## Handlers

Each handler:
- Receives HTTP requests
- Validates input using DTOs
- Calls appropriate use case
- Returns JSON responses

### Handler Pattern

```go
func (h *BookHandler) Create(w http.ResponseWriter, r *http.Request) {
    // 1. Extract request data
    var req dto.CreateBookRequest
    httputil.DecodeJSON(r, &req)

    // 2. Validate
    if !h.validator.ValidateStruct(w, req) {
        return
    }

    // 3. Execute use case
    result, err := h.useCases.CreateBook.Execute(ctx, ...)

    // 4. Respond
    h.RespondJSON(w, http.StatusCreated, result)
}
```

## Middleware

- **AuthMiddleware**: JWT token validation
- **LoggingMiddleware**: Request/response logging
- **RecoveryMiddleware**: Panic recovery
- **CORSMiddleware**: Cross-origin resource sharing

## DTOs (Data Transfer Objects)

- **Purpose**: Decouple HTTP layer from domain
- **Validation**: Using struct tags (`validate:"required,min=1"`)
- **Transformation**: Convert between DTOs and domain entities

## Routes

Defined in `router.go`:
- `/api/v1/auth/*` - Authentication
- `/api/v1/books/*` - Books (protected)
- `/api/v1/members/*` - Members (protected)
- `/api/v1/payments/*` - Payments (protected)

## Swagger Documentation

All handlers include Swagger annotations:
```go
// @Summary Create a new book
// @Tags books
// @Security BearerAuth
// @Param book body dto.CreateBookRequest true "Book data"
// @Success 201 {object} dto.BookResponse
// @Router /api/v1/books [post]
```

## Error Handling

Centralized error response using `RespondError()`:
- Maps domain errors to HTTP status codes
- Provides consistent error format
- Logs errors with context

## Testing

Mock HTTP requests using `httptest`:
```go
req := httptest.NewRequest("POST", "/api/v1/books", body)
w := httptest.NewRecorder()
handler.Create(w, req)
```