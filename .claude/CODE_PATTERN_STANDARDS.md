# Code Pattern Standards for Domains

## Overview

This document defines the **code-level patterns** that ALL domains must follow for consistency, maintainability, and predictability.

---

## 1. Entity Patterns

### Structure
```go
// Entity represents [description]
type EntityName struct {
    // ID is the unique identifier (UUID v4 format)
    ID string `db:"id" bson:"_id" json:"id"`

    // Required fields (value types)
    RequiredField string `db:"required_field" bson:"required_field" json:"required_field"`

    // Optional fields (pointer types)
    OptionalField *string `db:"optional_field" bson:"optional_field" json:"optional_field,omitempty"`

    // Timestamps
    CreatedAt time.Time  `db:"created_at" bson:"created_at" json:"created_at"`
    UpdatedAt time.Time  `db:"updated_at" bson:"updated_at" json:"updated_at"`
}
```

### Constructor Pattern
```go
// New creates a new EntityName instance.
// Returns the entity ready for validation and persistence.
func New(req Request) EntityName {
    now := time.Now()
    return EntityName{
        ID:            uuid.New().String(),
        RequiredField: req.RequiredField,
        OptionalField: req.OptionalField,
        CreatedAt:     now,
        UpdatedAt:     now,
    }
}
```

**Rules:**
- Constructor is always named `New`
- Takes domain `Request` DTO as parameter
- Generates UUID for ID
- Sets timestamps
- Returns entity by value (not pointer)

### Entity Methods
```go
// IsValid performs basic field validation.
// Returns true if entity has required fields populated.
func (e EntityName) IsValid() bool {
    return e.ID != "" && e.RequiredField != ""
}

// Update updates mutable fields from a request.
// Only updates non-nil fields from the request.
func (e *EntityName) Update(req UpdateRequest) {
    if req.RequiredField != "" {
        e.RequiredField = req.RequiredField
    }

    if req.OptionalField != nil {
        e.OptionalField = req.OptionalField
    }

    e.UpdatedAt = time.Now()
}
```

**Rules:**
- `IsValid()` - basic field check (receiver by value)
- `Update()` - modify entity (receiver by pointer)
- Business logic methods use consistent naming

---

## 2. Service Patterns

### Structure
```go
// Service provides business logic for EntityName domain.
// Domain services are stateless and contain pure business rules.
type Service struct {
    // Domain services are typically stateless
}

// NewService creates a new EntityName domain service.
func NewService() *Service {
    return &Service{}
}
```

**Rules:**
- Service struct is empty (stateless)
- Has explanatory comment about being stateless
- Constructor is `NewService()`
- Returns pointer to Service

### Validation Method Pattern
```go
// Validate performs comprehensive domain validation on EntityName.
// This enforces all business rules before persistence.
func (s *Service) Validate(entity EntityName) error {
    // Check required fields
    if entity.RequiredField == "" {
        return errors.NewError(errors.CodeValidation).
            WithDetail("field", "required_field").
            WithDetail("reason", "field is required").
            Build()
    }

    // Business rule validations
    if err := s.validateBusinessRule(entity); err != nil {
        return err
    }

    return nil
}
```

**Rules:**
- Method name: `Validate` (not `ValidateEntityName`)
- Takes entity as parameter (by value)
- Returns error
- Uses error builder pattern with `.Build()`
- Validates business rules, not just fields

### Business Logic Methods
```go
// MethodName performs [description].
// [Additional context about the method]
func (s *Service) MethodName(params...) (result, error) {
    // Implementation
}
```

**Rules:**
- Clear, descriptive names
- Start with verb (Validate, Calculate, Check, etc.)
- Document purpose and behavior
- Return error as last value

---

## 3. Repository Patterns

### Interface Definition
```go
// Repository defines persistence operations for EntityName.
// Implementations must handle errors consistently and use context for cancellation.
type Repository interface {
    // Add creates a new entity in the repository.
    // Returns the generated ID on success.
    Add(ctx context.Context, entity EntityName) (string, error)

    // Get retrieves an entity by ID.
    // Returns ErrNotFound if entity doesn't exist.
    Get(ctx context.Context, id string) (EntityName, error)

    // List retrieves all entities with optional filtering.
    List(ctx context.Context) ([]EntityName, error)

    // Update modifies an existing entity.
    // Returns ErrNotFound if entity doesn't exist.
    Update(ctx context.Context, entity EntityName) error

    // Delete removes an entity by ID.
    // Returns ErrNotFound if entity doesn't exist.
    Delete(ctx context.Context, id string) error
}
```

**Rules:**
- Interface name: `Repository`
- All methods take `context.Context` as first parameter
- CRUD methods: Add, Get, List, Update, Delete
- Add returns `(string, error)` (ID)
- Get/List return `(Entity/[]Entity, error)`
- Update/Delete return `error`
- Document error cases (ErrNotFound, etc.)

### Additional Repository Interfaces
```go
// SubEntityRepository defines persistence for sub-entities.
type SubEntityRepository interface {
    Create(ctx context.Context, entity SubEntity) (string, error)
    GetByID(ctx context.Context, id string) (SubEntity, error)
    // ...
}
```

**Rules:**
- Named `SubEntityRepository`
- Same context pattern
- Consistent CRUD naming

---

## 4. DTO Patterns

### Request DTOs
```go
// Request represents the input for creating EntityName.
type Request struct {
    RequiredField string  `json:"required_field" validate:"required"`
    OptionalField *string `json:"optional_field,omitempty"`
}

// Validate validates the request.
func (r Request) Validate() error {
    if r.RequiredField == "" {
        return errors.NewError(errors.CodeValidation).
            WithDetail("field", "required_field").
            WithDetail("reason", "field is required").
            Build()
    }
    return nil
}
```

### Update Request DTOs
```go
// UpdateRequest represents the input for updating EntityName.
type UpdateRequest struct {
    RequiredField string  `json:"required_field,omitempty"`
    OptionalField *string `json:"optional_field,omitempty"`
}
```

### Response DTOs
```go
// Response represents EntityName in API responses.
type Response struct {
    ID            string     `json:"id"`
    RequiredField string     `json:"required_field"`
    OptionalField *string    `json:"optional_field,omitempty"`
    CreatedAt     time.Time  `json:"created_at"`
    UpdatedAt     time.Time  `json:"updated_at"`
}

// ToResponse converts an entity to a response DTO.
func ToResponse(entity EntityName) Response {
    return Response{
        ID:            entity.ID,
        RequiredField: entity.RequiredField,
        OptionalField: entity.OptionalField,
        CreatedAt:     entity.CreatedAt,
        UpdatedAt:     entity.UpdatedAt,
    }
}

// ToResponses converts multiple entities to response DTOs.
func ToResponses(entities []EntityName) []Response {
    responses := make([]Response, len(entities))
    for i, entity := range entities {
        responses[i] = ToResponse(entity)
    }
    return responses
}
```

**Rules:**
- Request: `Request`, `UpdateRequest`, `CreateRequest`
- Response: `Response`, `DetailedResponse`
- Conversion functions: `ToResponse`, `ToResponses`
- All DTOs in single `dto.go` file

---

## 5. Error Handling Patterns

### Domain Error Construction
```go
// Use error builder pattern
return errors.NewError(errors.CodeValidation).
    WithDetail("field", "field_name").
    WithDetail("reason", "descriptive reason").
    Build()

// Or use domain errors
return errors.ErrNotFound
return errors.ErrAlreadyExists
```

### Error Wrapping in Repository
```go
if err != nil {
    return errors.Database("operation description", err)
}
```

### Error Wrapping in Use Case
```go
if err != nil {
    return errors.External("external service", err)
}
```

**Rules:**
- Always use error builder with `.Build()`
- Use `WithDetail` for context (not `WithDetails`)
- Wrap errors with context
- Use domain errors when appropriate

---

## 6. Documentation Patterns

### Package Documentation (doc.go)
```go
/*
Package entityname provides the EntityName domain model and business logic.

This package defines:
  - EntityName entity with validation
  - Service for business rules
  - Repository interface for persistence
  - DTOs for data transfer

Business Rules:
  - Rule 1 description
  - Rule 2 description

Usage:
    entity := entityname.New(req)
    svc := entityname.NewService()
    if err := svc.Validate(entity); err != nil {
        // handle error
    }
*/
package entityname
```

### Type Documentation
```go
// EntityName represents [full description].
//
// Key fields:
//   - ID: Unique identifier (UUID)
//   - Field: Description
//
// Business rules:
//   - Rule 1
//   - Rule 2
type EntityName struct { ... }
```

### Method Documentation
```go
// MethodName performs [action] on [subject].
// [Additional details about behavior]
//
// Returns error if [conditions].
func (s *Service) MethodName(params) error {
```

**Rules:**
- All exported types documented
- All exported methods documented
- Use godoc format
- Include examples when helpful

---

## 7. Naming Conventions

### Files
- `entity.go` - main entity
- `entity_subentity.go` - sub-entities
- `service.go` - domain service
- `repository.go` - repository interface
- `dto.go` - all DTOs
- `constants.go` - constants
- `interfaces.go` - external interfaces

### Types
- Entity: `EntityName` (PascalCase, singular)
- Service: `Service`
- Repository: `Repository` or `SubEntityRepository`
- DTOs: `Request`, `Response`, `UpdateRequest`
- Constants: `CONSTANT_NAME` or `ConstantName`

### Functions
- Constructors: `New`, `NewService`
- Validation: `Validate`, `ValidateField`
- Conversion: `ToResponse`, `ToEntity`
- Business logic: `Calculate`, `Check`, `Determine`

### Variables
- camelCase for local variables
- PascalCase for exported
- Descriptive names (no single letters except loops)

---

## 8. Testing Patterns

### Entity Tests
```go
func TestNew(t *testing.T) {
    req := Request{RequiredField: "value"}
    entity := New(req)

    assert.NotEmpty(t, entity.ID)
    assert.Equal(t, "value", entity.RequiredField)
}
```

### Service Tests
```go
func TestService_Validate(t *testing.T) {
    tests := []struct {
        name    string
        entity  EntityName
        wantErr bool
    }{
        {
            name: "valid entity",
            entity: EntityName{
                ID: "123",
                RequiredField: "value",
            },
            wantErr: false,
        },
        // More test cases
    }

    svc := NewService()
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := svc.Validate(tt.entity)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

**Rules:**
- Table-driven tests
- Test all validation paths
- Test edge cases
- Mock external dependencies

---

## 9. Import Organization

```go
package entityname

import (
    // Standard library
    "context"
    "time"

    // External dependencies
    "github.com/google/uuid"

    // Internal packages
    "library-service/pkg/errors"
)
```

**Rules:**
- Group imports (stdlib, external, internal)
- Sort alphabetically within groups
- Use goimports for automatic formatting

---

## 10. Code Quality Standards

### Cyclomatic Complexity
- Maximum 10 per function
- Extract complex logic to helper functions

### Function Length
- Maximum 50 lines preferred
- Maximum 100 lines absolute
- Extract long functions

### Variable Scope
- Minimize scope
- Declare close to usage
- No global mutable state

### Comments
- Document "why" not "what"
- Update with code changes
- Remove dead code, don't comment it out

---

## Checklist for New Domains

- [ ] Entity with `New(req Request)` constructor
- [ ] Service with `Validate(entity)` method
- [ ] Repository with CRUD interface
- [ ] DTOs with conversion functions
- [ ] Error builder pattern with `.Build()`
- [ ] Comprehensive documentation
- [ ] Table-driven tests
- [ ] Follows naming conventions
- [ ] Passes linter (golangci-lint)
- [ ] 80%+ test coverage

---

## Version

- **Version**: 1.0
- **Date**: October 11, 2025
- **Status**: Active
