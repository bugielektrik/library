# Unified Domain Pattern

## Overview

All domains in `/internal/domain` must follow this consistent structure for maintainability and clarity.

## Standard Domain Structure

```
domain/
├── doc.go              # Package documentation (REQUIRED)
├── entity.go           # Main domain entity and constructor (REQUIRED)
├── entity_*.go         # Additional entities (e.g., entity_receipt.go)
├── repository.go       # Repository interface (REQUIRED)
├── service.go          # Domain service with business logic (REQUIRED)
├── dto.go              # Domain transfer objects (OPTIONAL)
├── cache.go            # Cache interface (OPTIONAL)
├── constants.go        # Domain constants (OPTIONAL)
└── interfaces.go       # External interfaces (e.g., Gateway) (OPTIONAL)
```

## File Purposes

### Required Files

#### `doc.go`
- Package-level documentation
- Domain overview and purpose
- Key concepts and entities

#### `entity.go`
- Main domain entity struct
- Entity constructor (`New()` function)
- Entity methods (getters, business logic)
- Validation methods

#### `repository.go`
- Repository interface definition
- CRUD operations
- Domain-specific query methods

#### `service.go`
- Domain business logic
- Complex validations
- Business rules enforcement
- Stateless operations on entities

### Optional Files

#### `entity_*.go`
- Additional domain entities (when domain has multiple entities)
- Naming: `entity_<name>.go` (e.g., `entity_receipt.go`, `entity_callback_retry.go`)
- Each file contains one related entity

#### `dto.go`
- Domain-specific DTOs
- Value objects
- All DTOs for the domain (including sub-entities)
- **NO separate DTO files** for sub-entities

#### `cache.go`
- Cache interface definition
- Only for domains that benefit from caching
- Typically for frequently read, rarely updated data

#### `constants.go`
- Domain-specific constants
- Enums, status codes, limits
- Magic numbers with business meaning

#### `interfaces.go`
- External service interfaces (e.g., Gateway, EmailService)
- Third-party integration interfaces
- NOT repository or cache (they have their own files)

## Naming Conventions

### Files
- **Main entity**: `entity.go`
- **Sub-entities**: `entity_<name>.go` (lowercase, underscore-separated)
- **Service**: `service.go` (singular)
- **Repository**: `repository.go` (singular)

### Types
- **Entity**: `<DomainName>` (e.g., `Book`, `Payment`, `SavedCard`)
- **Service**: `Service` (e.g., `book.Service`)
- **Repository**: `Repository` (e.g., `book.Repository`)
- **DTOs**: `<Purpose>DTO` or descriptive name

### Functions
- **Constructor**: `New()` or `New<Entity>()`
- **Service constructor**: `NewService()`
- **Validation**: `Validate<Aspect>()`

## Domain Examples

### Simple Domain (Author, Book)
```
author/
├── doc.go          # Package documentation
├── entity.go       # Author entity
├── repository.go   # Author repository interface
├── service.go      # Author business logic
├── cache.go        # Author cache interface
└── dto.go          # Author DTOs
```

### Complex Domain (Payment)
```
payment/
├── doc.go                      # Package documentation
├── entity.go                   # Payment entity (main)
├── entity_saved_card.go        # SavedCard entity
├── entity_receipt.go           # Receipt entity
├── entity_callback_retry.go   # CallbackRetry entity
├── repository.go               # All repository interfaces
├── service.go                  # Payment business logic
├── dto.go                      # All payment DTOs
├── constants.go                # Payment statuses, limits
└── interfaces.go               # Gateway interface
```

### Minimal Domain (Reservation)
```
reservation/
├── doc.go          # Package documentation
├── entity.go       # Reservation entity
├── repository.go   # Reservation repository
├── service.go      # Reservation business logic
└── dto.go          # Reservation DTOs
```

## Best Practices

### Entity Files (`entity.go`)
1. Define the struct first
2. Constructor function next
3. Validation methods
4. Business logic methods
5. Helper methods

```go
// Entity definition
type Book struct { ... }

// Constructor
func New(params...) Book { ... }

// Validation
func (b Book) Validate() error { ... }

// Business logic
func (b Book) IsAvailable() bool { ... }
```

### Service Files (`service.go`)
1. Service struct (usually empty)
2. Constructor
3. Validation functions (complex rules)
4. Business logic functions (stateless)

```go
// Service struct
type Service struct {}

// Constructor
func NewService() *Service { return &Service{} }

// Business logic
func (s *Service) ValidateISBN(isbn string) error { ... }
```

### Repository Files (`repository.go`)
1. Main repository interface
2. Sub-entity repository interfaces (if any)
3. All in same file for cohesion

```go
// Main repository
type Repository interface {
    Add(ctx context.Context, payment Payment) (string, error)
    // ...
}

// Sub-entity repository
type SavedCardRepository interface {
    Create(ctx context.Context, card SavedCard) (string, error)
    // ...
}
```

## Migration Checklist

When refactoring a domain:

- [ ] Ensure `doc.go` exists and is comprehensive
- [ ] Main entity in `entity.go`
- [ ] Additional entities in `entity_*.go` files
- [ ] All DTOs in single `dto.go` file
- [ ] Repository interfaces in `repository.go`
- [ ] Business logic in `service.go`
- [ ] Constants in `constants.go` (if needed)
- [ ] External interfaces in `interfaces.go` (if needed)
- [ ] Cache interface in `cache.go` (if beneficial)
- [ ] Update all imports
- [ ] Run tests
- [ ] Verify build

## Anti-Patterns to Avoid

❌ **Separate DTO files for sub-entities**
```
payment/
├── dto.go
├── saved_card_dto.go  # ❌ Wrong
└── receipt_dto.go     # ❌ Wrong
```

✅ **All DTOs in one file**
```
payment/
└── dto.go  # Contains all DTOs
```

❌ **Entity files without entity_ prefix**
```
payment/
├── payment.go         # Main entity - OK
├── saved_card.go      # ❌ Wrong - not clear it's an entity
└── receipt.go         # ❌ Wrong - not clear it's an entity
```

✅ **Clear entity naming**
```
payment/
├── entity.go               # Main entity
├── entity_saved_card.go    # ✅ Correct
└── entity_receipt.go       # ✅ Correct
```

❌ **Missing service.go**
Every domain with business logic needs service.go

❌ **Multiple service files**
One service.go per domain (can have helper files if needed)

## Benefits of This Pattern

1. **Consistency**: All domains follow same structure
2. **Predictability**: Developers know where to find code
3. **Scalability**: Pattern works for simple and complex domains
4. **Maintainability**: Clear separation of concerns
5. **Discoverability**: File names indicate purpose
6. **Testing**: Easy to locate and test each layer

## Version

- **Version**: 1.0
- **Date**: October 11, 2025
- **Status**: Active
