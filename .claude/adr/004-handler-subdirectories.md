# ADR 004: Handler Organization by Domain

**Status:** Accepted

**Date:** 2025-10-09

**Context:**

As the application grew, all HTTP handlers were in a flat directory structure:

```
internal/infrastructure/pkg/handlers/
├── auth.go
├── book.go
├── author.go
├── member.go
├── payment.go
├── receipt.go
├── reservation.go
├── saved_card.go
└── ... (22 files total)
```

**Problems:**
1. **Hard to navigate:** 22 files in one directory
2. **Unclear boundaries:** Which files relate to which domain?
3. **Package conflicts:** All handlers in "handlers" package
4. **Scaling issues:** Adding more endpoints makes it worse
5. **Merge conflicts:** Multiple people editing same directory

## Decision

Organize handlers into **domain-specific subdirectories**:

```
internal/infrastructure/pkg/handlers/
├── auth/
│   ├── handler.go       # Handler struct + routes
│   ├── login.go         # Login endpoint
│   ├── register.go      # Registration endpoint
│   └── doc.go           # Package documentation
├── book/
│   ├── handler.go       # Handler struct + routes
│   ├── crud.go          # CRUD operations
│   ├── query.go         # Query operations
│   └── doc.go
├── payment/
│   ├── handler.go       # Handler struct + routes
│   ├── initiate.go      # Initiate payment
│   ├── callback.go      # Gateway callbacks
│   ├── manage.go        # Cancel, refund
│   ├── query.go         # List, verify
│   └── doc.go
└── ... (8 domains total)
```

## Implementation

### File Organization Pattern

Each domain subdirectory contains:

1. **handler.go** - Handler struct, constructor, route registration
2. **Operation files** - Grouped by operation type (CRUD, queries, etc.)
3. **doc.go** - Package documentation

### Example: Book Handler

```go
// internal/infrastructure/pkg/handler/book/handler.go
package book

type BookHandler struct {
    useCases struct {
        CreateBook *bookops.CreateBookUseCase
        GetBook    *bookops.GetBookUseCase
        // ...
    }
    validator *httputil.Validator
}

func NewBookHandler(useCases /*...*/) *BookHandler {
    return &BookHandler{...}
}

func (h *BookHandler) RegisterRoutes(r chi.Router) {
    r.Route("/books", func(r chi.Router) {
        r.Post("/", h.create)           // From crud.go
        r.Get("/", h.list)              // From query.go
        r.Get("/{id}", h.get)           // From crud.go
        r.Put("/{id}", h.update)        // From crud.go
        r.Delete("/{id}", h.delete)     // From crud.go
        r.Get("/{id}/authors", h.listAuthors)  // From query.go
    })
}
```

```go
// internal/infrastructure/pkg/handler/book/crud.go
package book

func (h *BookHandler) create(w http.ResponseWriter, r *http.Request) {
    // Implementation
}

func (h *BookHandler) update(w http.ResponseWriter, r *http.Request) {
    // Implementation
}
// ...
```

```go
// internal/infrastructure/pkg/handler/book/query.go
package book

func (h *BookHandler) list(w http.ResponseWriter, r *http.Request) {
    // Implementation
}

func (h *BookHandler) listAuthors(w http.ResponseWriter, r *http.Request) {
    // Implementation
}
```

### File Splitting Strategy

**When to split:**
- More than 200 lines in one file
- Mix of different operation types (CRUD vs queries vs admin)
- Difficult to find specific endpoint

**How to split:**
- **crud.go** - Create, Read (single), Update, Delete
- **query.go** - List operations, searches, filters
- **manage.go** - Admin operations (cancel, refund, approve)
- **callback.go** - Webhook/callback endpoints
- **page.go** - Page rendering (if applicable)

## Consequences

### Positive

✅ **Better navigation:**
```bash
# Find book endpoints
ls handler/book/

# Find payment endpoints
ls handler/payment/
```

✅ **Clear ownership:** Each domain has its own package

✅ **Reduced merge conflicts:** Changes isolated to domain subdirectories

✅ **Scalable:** Easy to add more domains without cluttering

✅ **Domain-specific packages:**
```go
import "library-service/internal/infrastructure/pkg/handler/book"
import "library-service/internal/infrastructure/pkg/handler/payment"
```

✅ **Package documentation:** Each handler has doc.go explaining its purpose

### Negative

❌ **More directories:** 8 subdirectories vs 1 flat directory

❌ **Deeper paths:**
- Before: `handlers/book.go`
- After: `handlers/book/crud.go`

❌ **Migration effort:** Had to update imports across codebase

## Migration Process

### What We Did

1. **Created subdirectories:** One per domain (auth, book, payment, etc.)

2. **Moved files:**
```bash
mv handler/book.go handler/book/handler.go
mv handler/auth_*.go handler/auth/
# ...
```

3. **Updated package declarations:**
```go
// Before
package handlers

// After
package book
```

4. **Updated imports in router.go:**
```go
// Before
import "library-service/internal/infrastructure/pkg/handler"

// After
import (
    "library-service/internal/infrastructure/pkg/handler/auth"
    "library-service/internal/infrastructure/pkg/handler/book"
    "library-service/internal/infrastructure/pkg/handler/payment"
    // ...
)
```

5. **Split large files:** Payment handler (400+ lines) split into 5 files

6. **Added doc.go:** Package documentation for each subdirectory

7. **Verified:** `go build ./...` and `go test ./...`

## Domain Subdirectories

Current organization:

| Subdirectory | Endpoints | Files |
|--------------|-----------|-------|
| `auth/` | Register, Login, Refresh | 4 files |
| `author/` | List authors | 2 files |
| `book/` | CRUD, list authors | 3 files |
| `member/` | List, profile | 2 files |
| `payment/` | Initiate, verify, refund, cancel | 6 files |
| `receipt/` | Generate, get, list | 2 files |
| `reservation/` | Create, cancel, list | 2 files |
| `savedcard/` | Save, list, delete, set default | 2 files |

## Alternative Considered

### Grouping by Operation Type (Rejected)

```
handlers/
├── crud/
│   ├── book.go
│   ├── member.go
├── queries/
│   ├── book.go
│   ├── payment.go
```

**Why rejected:**
- Splits related endpoints across directories
- Doesn't follow domain-driven design
- Harder to find all operations for one domain

## Related Decisions

- **ADR 001:** Use Case "ops" Suffix - Similar domain-based organization
- **ADR 002:** Clean Architecture - Handlers are in adapter layer

## References

- **Implementation:** `internal/infrastructure/pkg/handlers/*/`
- **Commit:** fa693e6 - Initial handler reorganization
- **Documentation:** Each handler has `doc.go` explaining its purpose

## Notes for AI Assistants

### Adding New Handler

1. Create subdirectory: `internal/infrastructure/pkg/handlers/{domain}/`
2. Add files:
   - `handler.go` - Struct + route registration
   - `crud.go` or `operations.go` - Endpoint implementations
   - `doc.go` - Package documentation

3. Register in router:
```go
// internal/infrastructure/server/router.go
domainHandler := handlers.NewDomainHandler(useCases.Domain...)
domainHandler.RegisterRoutes(r)
```

### File Naming Conventions

- `handler.go` - Always the handler struct and route registration
- `crud.go` - Standard CRUD operations
- `query.go` - List and search operations
- `manage.go` - Admin/special operations
- `callback.go` - Webhook endpoints
- `doc.go` - Package documentation (required)

## Revision History

- **2025-10-09:** Handler reorganization implemented
- **2025-10-09:** Initial ADR documenting the pattern
