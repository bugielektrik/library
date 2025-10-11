# Test Mocks

This directory contains centralized mock implementations for testing.

## Purpose

Mocks allow you to test code in isolation by replacing real dependencies with controllable test doubles. This project uses two types of mocks:

1. **Auto-generated mocks** (via mockery) - For repository and cache interfaces
2. **Manual mocks** (in this directory) - For integration tests and special cases

## Mock Locations

### Repository & Cache Mocks (Auto-generated)

**Location**: `internal/adapters/repository/mocks/`

**Generated via**: [mockery](https://github.com/vektra/mockery)

**Available Mocks**:
- `MockRepository` - Book repository interface
- `MockCache` - Book cache interface

**Usage Example**:
```go
import (
    "testing"
    "github.com/stretchr/testify/mock"
    "library-service/internal/adapters/repository/mocks"
    "library-service/internal/domain/book"
)

func TestBookUseCase(t *testing.T) {
    // Create mock repository
    mockRepo := mocks.NewMockRepository(t)

    // Setup expectations
    mockRepo.EXPECT().
        Get(mock.Anything, "book-id").
        Return(book.Book{ID: "book-id"}, nil).
        Once()

    // Use mock in your test
    // ...

    // Verify expectations were met (automatic cleanup via t.Cleanup)
}
```

### Integration Test Mocks (Manual)

**Location**: `test/mocks/` (this directory)

**Purpose**: Mocks for integration tests that need more control or don't have interfaces in the domain layer.

**Available Mocks**:
- `PaymentGateway` - Mock payment gateway for integration tests

**Usage Example**:
```go
//go:build integration
// +build integration

import (
    "testing"
    "library-service/test/mocks"
)

func TestPaymentIntegration(t *testing.T) {
    // Create mock gateway
    gateway := mocks.NewPaymentGateway()

    // Customize behavior
    gateway.SetCheckPaymentResponse(map[string]interface{}{
        "Status": "success",
    })

    // Use in integration test
    // ...
}
```

## Generating New Mocks

### Prerequisites

Install mockery:
```bash
go install github.com/vektra/mockery/v2@latest
```

### Configuration

Mock generation is configured in `.mockery.yaml` at the project root.

**Current configuration**:
- Output directory: `internal/adapters/repository/mocks/`
- Naming pattern: `Mock{InterfaceName}`
- Expecter pattern enabled for better API

**Configured interfaces**:
- `internal/domain/book` - Repository, Cache
- `internal/domain/author` - Repository, Cache
- `internal/domain/member` - Repository
- `internal/domain/reservation` - Repository
- `internal/domain/payment` - Repository

### Generate Mocks

```bash
# Generate all mocks
make gen-mocks

# Or directly with mockery
mockery --all
```

## Adding New Mocks

### For Domain Repository/Cache Interfaces

1. **Add interface to `.mockery.yaml`**:
```yaml
packages:
  library-service/internal/domain/newdomain:
    interfaces:
      Repository:
      Cache:  # Only if cache interface exists
```

2. **Generate mocks**:
```bash
make gen-mocks
```

3. **Use in tests**:
```go
import "library-service/internal/adapters/repository/mocks"

mockRepo := mocks.NewMockRepository(t)
```

### For Manual Mocks (Integration Tests)

1. **Create mock file** in `test/mocks/`:
```go
//go:build integration
// +build integration

package mocks

type MyServiceMock struct {
    // fields
}

func NewMyServiceMock() *MyServiceMock {
    return &MyServiceMock{}
}

// Implement interface methods...
```

2. **Document in this README** under "Available Mocks"

3. **Use in integration tests**:
```go
//go:build integration

import "library-service/test/mocks"

mock := mocks.NewMyServiceMock()
```

## Mock Patterns

### Using Mockery Mocks

**Pattern 1: Expecter API (Recommended)**
```go
mockRepo := mocks.NewMockRepository(t)

// Setup expectation
mockRepo.EXPECT().
    Add(mock.Anything, mock.MatchedBy(func(b book.Book) bool {
        return b.Name != nil && *b.Name == "Expected Name"
    })).
    Return("book-id", nil).
    Once()
```

**Pattern 2: Classic API**
```go
mockRepo := mocks.NewMockRepository(t)

// Setup expectation
mockRepo.On("Add", mock.Anything, mock.Anything).
    Return("book-id", nil).
    Once()
```

**Pattern 3: Multiple Calls**
```go
// Expect exactly 3 calls
mockRepo.EXPECT().
    Get(mock.Anything, "book-id").
    Return(book.Book{ID: "book-id"}, nil).
    Times(3)

// Expect at least 1 call
mockRepo.EXPECT().
    Get(mock.Anything, "book-id").
    Return(book.Book{ID: "book-id"}, nil)
```

**Pattern 4: Argument Matching**
```go
// Match any context, specific ID
mockRepo.EXPECT().
    Get(mock.Anything, "specific-id").
    Return(book.Book{}, nil)

// Custom matcher
mockRepo.EXPECT().
    Add(mock.Anything, mock.MatchedBy(func(b book.Book) bool {
        return b.ISBN != nil && len(*b.ISBN) == 13
    })).
    Return("id", nil)
```

**Pattern 5: Error Returns**
```go
// Return specific error
mockRepo.EXPECT().
    Get(mock.Anything, "invalid-id").
    Return(book.Book{}, errors.ErrNotFound)

// Use testify's assert.AnError for generic errors
mockRepo.EXPECT().
    Get(mock.Anything, "error-id").
    Return(book.Book{}, assert.AnError)
```

### Using Manual Mocks

**Pattern 1: Constructor with Defaults**
```go
func NewMyMock() *MyMock {
    return &MyMock{
        field: "default-value",
    }
}
```

**Pattern 2: Setter Methods for Test Control**
```go
mock := NewMyMock()
mock.SetBehavior("custom-behavior")
mock.SetResponse(expectedResponse)
```

**Pattern 3: Build Tags for Integration Only**
```go
//go:build integration
// +build integration

package mocks
```

## Mock Best Practices

### ✅ Do

- **Use mockery for repository/cache interfaces** - Automatic generation and type safety
- **Use manual mocks for complex integration scenarios** - More control when needed
- **Always call `AssertExpectations(t)`** - Verify mocks were called as expected (automatic with mockery v2)
- **Use meaningful test names** - Clear expectation setup
- **Test failure paths** - Mock error returns
- **Keep mocks focused** - One interface per mock

### ❌ Don't

- **Don't mock domain entities** - Use real entities or fixtures
- **Don't mock everything** - Only mock external dependencies
- **Don't use mocks in domain tests** - Domain should be pure logic
- **Don't manually write repository mocks** - Use mockery
- **Don't share mock instances** - Create fresh mocks per test

## Testing with Mocks vs Fixtures vs Real Data

| Approach | Use Case | Example |
|----------|----------|---------|
| **Mocks** | External dependencies (DB, cache, APIs) | Repository, Cache, Payment Gateway |
| **Fixtures** | Test data for entities | `fixtures.ValidBook()`, `fixtures.AdminMember()` |
| **Real Data** | Integration tests, E2E tests | Database, actual HTTP requests |

**Example: Use Case Test (Mocks + Fixtures)**
```go
func TestCreateBook(t *testing.T) {
    // Mock external dependencies
    mockRepo := mocks.NewMockRepository(t)
    mockCache := mocks.NewMockCache(t)

    // Use fixtures for test data
    bookData := fixtures.CreateBookRequest()

    // Setup mock expectations
    mockRepo.EXPECT().Add(mock.Anything, mock.Anything).Return("id", nil)
    mockCache.EXPECT().Set(mock.Anything, "id", mock.Anything).Return(nil)

    // Execute use case
    uc := bookops.NewCreateBookUseCase(mockRepo, mockCache, bookService)
    result, err := uc.Execute(ctx, bookData)

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "id", result.ID)
}
```

## Troubleshooting

### "mockery: command not found"

Install mockery:
```bash
go install github.com/vektra/mockery/v2@latest
```

### "Mock expectations not met"

Ensure you're calling methods exactly as expected:
```go
// This expectation
mockRepo.EXPECT().Get(mock.Anything, "book-1")

// Will NOT match this call
mockRepo.Get(ctx, "book-2")  // Different ID!
```

### "Mock returns zero values"

You must setup expectations BEFORE using the mock:
```go
// ✅ Correct order
mockRepo.EXPECT().Get(...).Return(book.Book{ID: "1"}, nil)
result := uc.Execute(ctx, request)

// ❌ Wrong order - returns zero values
result := uc.Execute(ctx, request)
mockRepo.EXPECT().Get(...).Return(book.Book{ID: "1"}, nil)
```

### Regenerating Mocks Fails

Check `.mockery.yaml` configuration:
```bash
# Validate config
cat .mockery.yaml

# Check interface exists
ls internal/domain/book/repository.go

# Regenerate with verbose output
mockery --config .mockery.yaml --verbose
```

## Related Documentation

- [Testing Guide](../../.claude/testing.md) - Comprehensive testing strategies
- [Test Fixtures](../fixtures/README.md) - Reusable test data
- [Test Utilities](../testutil/README.md) - Test helper functions
- [mockery Documentation](https://vektra.github.io/mockery/) - Official mockery docs

## Migration Notes

**Previous Structure** (Before Phase 7.3.3):
- Integration test mocks scattered in `test/integration/mocks.go`
- No centralized location for manual mocks

**Current Structure** (After Phase 7.3.3):
- Auto-generated mocks: `internal/adapters/repository/mocks/`
- Manual mocks: `test/mocks/`
- Clear separation and documentation
