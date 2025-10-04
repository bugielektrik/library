# Package Overview & Dependency Map

**Visual guide to the Library Management System package structure and dependencies.**

## Package Structure Diagram

```
library-service/
│
├── cmd/                           # Entry points
│   ├── api/                      # REST API server
│   ├── worker/                   # Background worker
│   └── migrate/                  # Database migrations
│
├── internal/                     # Private application code
│   │
│   ├── domain/                   # 🏛️ DOMAIN LAYER (Core)
│   │   ├── book/                # Book domain
│   │   │   ├── entity.go        # Book entity
│   │   │   ├── service.go       # Business rules
│   │   │   ├── repository.go    # Repository interface
│   │   │   └── cache.go         # Cache interface
│   │   │
│   │   ├── member/              # Member domain
│   │   │   ├── entity.go
│   │   │   ├── service.go       # Subscription logic
│   │   │   └── repository.go
│   │   │
│   │   └── author/              # Author domain
│   │       ├── entity.go
│   │       ├── repository.go
│   │       └── cache.go
│   │
│   ├── usecase/                  # 🎯 USE CASE LAYER
│   │   ├── book/                # Book use cases
│   │   │   ├── create_book.go
│   │   │   ├── update_book.go
│   │   │   ├── delete_book.go
│   │   │   ├── list_books.go
│   │   │   └── dto.go
│   │   │
│   │   ├── member/              # Member use cases
│   │   │   ├── create_member.go
│   │   │   └── dto.go
│   │   │
│   │   └── subscription/        # Subscription use cases
│   │       ├── subscribe_member.go
│   │       └── dto.go
│   │
│   └── adapters/                 # 🔌 ADAPTER LAYER
│       ├── http/                # HTTP handlers
│       │   ├── book/           # Book endpoints
│       │   │   ├── handler.go
│       │   │   ├── dto.go
│       │   │   └── mapper.go
│       │   ├── member/         # Member endpoints
│       │   └── middleware/     # HTTP middleware
│       │
│       ├── repository/          # Database implementations
│       │   ├── book_postgres.go
│       │   ├── member_postgres.go
│       │   └── author_postgres.go
│       │
│       ├── cache/               # Cache implementations
│       │   ├── book_redis.go
│       │   └── author_redis.go
│       │
│       └── storage/             # File/object storage
│           └── s3_storage.go
│
├── pkg/                          # 🔧 SHARED UTILITIES
│   ├── errors/                  # Error handling
│   ├── validator/               # Validation helpers
│   ├── logger/                  # Logging utilities
│   └── config/                  # Configuration
│
├── test/                         # 🧪 TEST INFRASTRUCTURE
│   ├── fixtures/                # Shared test data
│   ├── integration/             # Integration tests
│   └── e2e/                     # End-to-end tests
│
├── examples/                     # 📚 CODE EXAMPLES
│   ├── basic_crud/              # CRUD examples
│   ├── domain_service/          # Domain service examples
│   └── testing/                 # Testing examples
│
├── docs/                         # 📖 DOCUMENTATION
│   ├── architecture.md          # Architecture overview
│   ├── adr/                     # Architecture decisions
│   └── guides/                  # Developer guides
│
├── api/                          # 📡 API SPECIFICATIONS
│   └── openapi/                 # OpenAPI/Swagger specs
│
└── deployments/                  # 🚀 DEPLOYMENT
    └── docker/                  # Docker configs
```

## Dependency Flow Diagram

### Layer Dependencies (Clean Architecture)

```
┌─────────────────────────────────────────────────────┐
│                  INFRASTRUCTURE                     │
│         (Gin, PostgreSQL, Redis, Docker)            │
└─────────────────────────────────────────────────────┘
                          ↑ implements
┌─────────────────────────────────────────────────────┐
│                  ADAPTERS LAYER                     │
│      (HTTP Handlers, Repositories, Cache)           │
│                                                     │
│  internal/adapters/                                 │
│  ├── http/          (Gin handlers)                 │
│  ├── repository/    (PostgreSQL impl)               │
│  └── cache/         (Redis impl)                    │
└─────────────────────────────────────────────────────┘
                          ↑ depends on
┌─────────────────────────────────────────────────────┐
│                 USE CASE LAYER                      │
│          (Application Business Logic)               │
│                                                     │
│  internal/usecase/                                  │
│  ├── book/          (Book operations)               │
│  ├── member/        (Member operations)             │
│  └── subscription/  (Subscription flows)            │
└─────────────────────────────────────────────────────┘
                          ↑ depends on
┌─────────────────────────────────────────────────────┐
│                  DOMAIN LAYER                       │
│           (Core Business Logic)                     │
│                                                     │
│  internal/domain/                                   │
│  ├── book/          (Entities, Services, Interfaces)│
│  ├── member/        (Entities, Services, Interfaces)│
│  └── author/        (Entities, Interfaces)          │
│                                                     │
│  ⚠️  NO EXTERNAL DEPENDENCIES                       │
└─────────────────────────────────────────────────────┘

Key: ↑ = "depends on" (arrows point to dependencies)
```

## Package Relationships

### Book Domain Flow

```
HTTP Request
    ↓
BookHandler (adapters/http/book)
    ↓ calls
CreateBookUseCase (usecase/book)
    ↓ uses
┌─────────────────┬──────────────────┬─────────────────┐
│                 │                  │                 │
BookRepository    BookService        BookCache
(domain/book)     (domain/book)      (domain/book)
│                 │                  │
↓ implements      ↓ (no deps)        ↓ implements
│                                    │
PostgresBookRepo                     RedisBookCache
(adapters/repository)                (adapters/cache)
```

### Member Subscription Flow

```
HTTP Request (POST /subscribe)
    ↓
MemberHandler (adapters/http/member)
    ↓ calls
SubscribeMemberUseCase (usecase/subscription)
    ↓ uses
┌──────────────────┬────────────────────┬──────────────┐
│                  │                    │              │
MemberRepository   MemberService        SubscriptionRepo
(domain/member)    (domain/member)      (domain/member)
│                  │                    │
│                  ↓ calculates         │
│                  - Pricing            │
│                  - Expiration         │
│                  - Grace Period       │
│                                       │
↓ implements                            ↓ implements
PostgresMemberRepo                      PostgresSubRepo
(adapters/repository)                   (adapters/repository)
```

## Cross-Domain Relationships

### Book ↔ Author

```
Book Entity
  ├── Authors: []string  (author IDs)
  │
  └── GetAuthors() use case
        ↓
        AuthorRepository.GetByIDs(authorIDs)
        ↓
        Returns: []Author
```

### Member ↔ Book

```
Member Entity
  ├── Books: []string  (book IDs)
  │
  └── GetBorrowedBooks() use case
        ↓
        BookRepository.GetByIDs(bookIDs)
        ↓
        Returns: []Book
```

## Shared Package Usage

```
All Layers
    ↓ can use
┌──────────────────────────────────┐
│        pkg/ (Utilities)          │
│                                  │
│  ├── errors/     (Error types)   │
│  ├── validator/  (Validation)    │
│  ├── logger/     (Logging)       │
│  └── config/     (Config)        │
│                                  │
│  ✅ No dependencies on internal/ │
└──────────────────────────────────┘
```

## Test Package Dependencies

```
Integration Tests (test/integration/)
    ↓ uses
┌─────────────────────────────────────┐
│     Test Fixtures (test/fixtures/)  │
│                                     │
│  ├── books.go      (Book entities)  │
│  ├── members.go    (Member entities)│
│  └── authors.go    (Author entities)│
└─────────────────────────────────────┘
    ↓ uses
Domain Entities (internal/domain/)
```

## Import Rules

### ✅ Allowed Imports

```go
// Domain layer
package book
import (
    "context"           // ✅ Standard library
    "time"              // ✅ Standard library
    // NO external dependencies
)

// Use case layer
package usecase
import (
    "context"                          // ✅ Standard library
    "library-service/internal/domain/book"  // ✅ Domain layer
    "library-service/pkg/errors"           // ✅ Shared utilities
)

// Adapter layer
package http
import (
    "github.com/gin-gonic/gin"              // ✅ External framework
    "library-service/internal/usecase/book"  // ✅ Use case layer
    "library-service/pkg/errors"            // ✅ Shared utilities
)
```

### ❌ Forbidden Imports

```go
// Domain layer
package book
import (
    "library-service/internal/usecase/book"  // ❌ Cannot depend on use case
    "library-service/internal/adapters/..."   // ❌ Cannot depend on adapters
    "github.com/gin-gonic/gin"               // ❌ No external frameworks
)

// Use case layer
package usecase
import (
    "library-service/internal/adapters/..."  // ❌ Cannot depend on adapters
)
```

## Package Metrics

| Package | Files | Lines | Test Coverage | Complexity |
|---------|-------|-------|---------------|------------|
| `domain/book` | 6 | ~400 | 100% | Low |
| `domain/member` | 5 | ~350 | 100% | Low |
| `usecase/book` | 8 | ~600 | 85% | Medium |
| `adapters/http` | 12 | ~800 | 65% | Medium |
| `adapters/repository` | 6 | ~500 | 70% | Medium |
| **Total** | **~60** | **~5000** | **75%** | **Low-Med** |

## Key Design Patterns

### 1. Repository Pattern
```
Domain defines interface → Adapter implements
book.Repository (interface) → PostgresBookRepository (impl)
```

### 2. Dependency Injection
```
Constructor injection throughout
NewCreateBookUseCase(repo, service, cache)
```

### 3. DTO Pattern
```
HTTP DTO → Use Case DTO → Domain Entity
CreateBookRequest → CreateBookInput → book.Entity
```

### 4. Service Pattern
```
Complex business logic in domain services
book.Service.ValidateISBN()
member.Service.CalculateSubscriptionPrice()
```

## Navigation Guide

### To Add a New Feature

1. **Start**: `internal/domain/{domain}/` - Define entity & business rules
2. **Then**: `internal/domain/{domain}/repository.go` - Define interface
3. **Next**: `internal/usecase/{domain}/` - Create use case
4. **Then**: `internal/adapters/repository/` - Implement repository
5. **Finally**: `internal/adapters/http/` - Create HTTP handler

### To Find Business Logic

1. **Domain Services**: `internal/domain/{domain}/service.go`
2. **Use Cases**: `internal/usecase/{domain}/`
3. **Validation**: `internal/domain/{domain}/service.go` or `pkg/validator/`

### To Find Infrastructure

1. **Database**: `internal/adapters/repository/`
2. **Cache**: `internal/adapters/cache/`
3. **HTTP**: `internal/adapters/http/`
4. **Config**: `pkg/config/`

## Quick Reference

### File Naming Conventions

- **Entities**: `entity.go`
- **Services**: `service.go`
- **Repositories**: `repository.go` (interface), `{domain}_postgres.go` (impl)
- **Use Cases**: `{action}_{entity}.go` (e.g., `create_book.go`)
- **Handlers**: `handler.go`
- **Tests**: `*_test.go`, `*_benchmark_test.go`

### Package Naming

- **Domain packages**: Singular (`book`, not `books`)
- **Use case packages**: Singular (`book`, not `books`)
- **Test packages**: Append `_test` for black-box testing

## References

- [Architecture Overview](./architecture.md)
- [ADR-001: Clean Architecture](./adr/001-clean-architecture.md)
- [Development Guide](./guides/DEVELOPMENT.md)
- [Domain Layer README](../internal/domain/README.md)
- [Use Case Layer README](../internal/usecase/README.md)
- [Adapter Layer README](../internal/adapters/README.md)
