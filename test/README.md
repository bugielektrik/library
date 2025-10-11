# Test Infrastructure

This directory contains the test infrastructure for the Library Management System.

## Directory Structure

```
test/
├── fixtures/          # Reusable test data (entities, requests, responses)
├── mocks/            # Manual mocks for integration tests
├── testutil/         # Test helper functions (assertions, context)
└── integration/      # Integration tests
```

## Quick Start

### Unit Tests (Domain/Use Cases)

```go
import (
    "library-service/test/fixtures"
    "library-service/test/testutil"
    "library-service/internal/adapters/repository/mocks"
)

func TestBookService(t *testing.T) {
    // Use fixtures for test data
    book := fixtures.ValidBook()

    // Use mocks for dependencies
    mockRepo := mocks.NewMockRepository(t)
    mockRepo.EXPECT().Add(mock.Anything, book).Return("id", nil)

    // Use testutil for assertions
    result, err := service.CreateBook(ctx, book)
    testutil.AssertNoError(t, err)
    testutil.AssertEqual(t, "id", result.ID)
}
```

### Integration Tests

```go
//go:build integration
// +build integration

import (
    "library-service/test/mocks"
    "library-service/test/fixtures"
)

func TestPaymentFlow(t *testing.T) {
    // Use mocks for external services
    gateway := mocks.NewPaymentGateway()

    // Use fixtures for test data
    payment := fixtures.ValidPayment()

    // Test with real database
    // ...
}
```

## Packages

### fixtures/ - Test Data

**Purpose**: Provides reusable test data for entities, requests, and responses.

**Benefits**:
- Consistent test data across all tests
- Easy to create valid/invalid test scenarios
- Reduces test boilerplate

**Example**:
```go
book := fixtures.ValidBook()
author := fixtures.ValidAuthor()
member := fixtures.MemberWithBooks()
```

**Documentation**: [fixtures/README.md](./fixtures/README.md)

### mocks/ - Integration Test Mocks

**Purpose**: Manual mock implementations for integration tests.

**Use Cases**:
- External services (payment gateways, APIs)
- Services without domain interfaces
- Complex integration scenarios

**Example**:
```go
gateway := mocks.NewPaymentGateway()
gateway.SetCheckPaymentResponse(map[string]interface{}{
    "Status": "success",
})
```

**Documentation**: [mocks/README.md](./mocks/README.md)

### testutil/ - Test Helpers

**Purpose**: Helper functions for cleaner test assertions and setup.

**Benefits**:
- One-line assertions instead of verbose if-else
- Consistent error messages
- Better test readability

**Example**:
```go
testutil.AssertNoError(t, err)
testutil.AssertEqual(t, expected, actual)
testutil.AssertStringContains(t, str, "substring")
```

**Documentation**: [testutil/README.md](./testutil/README.md)

### integration/ - Integration Tests

**Purpose**: End-to-end tests with real database and services.

**Build Tag**: `//go:build integration`

**Running**:
```bash
make test-integration
```

## Test Types

### Unit Tests

**Location**: Next to source files (`*_test.go`)

**Characteristics**:
- Fast (< 1 second per test)
- No external dependencies
- Use mocks for repositories/caches
- Focus on business logic

**Example Structure**:
```go
func TestCreateBook(t *testing.T) {
    // Setup: Create mocks
    mockRepo := mocks.NewMockRepository(t)

    // Arrange: Setup test data and expectations
    book := fixtures.CreateBookRequest()
    mockRepo.EXPECT().Add(mock.Anything, mock.Anything).Return("id", nil)

    // Act: Execute the function
    result, err := uc.Execute(ctx, book)

    // Assert: Verify results
    testutil.AssertNoError(t, err)
    testutil.AssertEqual(t, "id", result.ID)
}
```

### Integration Tests

**Location**: `test/integration/`

**Characteristics**:
- Slower (real database operations)
- Real PostgreSQL database
- Mock external services only
- Test full workflows

**Example Structure**:
```go
//go:build integration

func TestPaymentFlow(t *testing.T) {
    // Setup: Real database
    db, cleanup := setupTestDB(t)
    defer cleanup()

    // Use real repositories
    paymentRepo := postgres.NewPaymentRepository(db)

    // Mock external services
    gateway := mocks.NewPaymentGateway()

    // Test workflow
    // ...
}
```

## Mock Locations

| Mock Type | Location | Generator | Use Case |
|-----------|----------|-----------|----------|
| Repository Mocks | `internal/adapters/repository/mocks/` | mockery | Unit tests |
| Cache Mocks | `internal/adapters/repository/mocks/` | mockery | Unit tests |
| Service Mocks | `test/mocks/` | manual | Integration tests |
| Integration Mocks | `test/integration/mocks.go` | manual | Integration tests only |

## Running Tests

### All Tests
```bash
make test
```

### Unit Tests Only
```bash
make test-unit
go test ./internal/...
```

### Integration Tests Only
```bash
make test-integration
go test -tags integration ./test/integration/...
```

### Specific Package
```bash
go test -v ./internal/domain/book/...
go test -v ./internal/usecase/bookops/...
```

### With Coverage
```bash
make test-coverage
```

### With Race Detection
```bash
go test -race ./...
```

## Test Patterns

### Pattern 1: Table-Driven Tests

**Best for**: Testing multiple scenarios with similar structure

```go
func TestBookValidation(t *testing.T) {
    tests := []struct {
        name    string
        book    book.Book
        wantErr bool
    }{
        {
            name:    "valid book",
            book:    fixtures.ValidBook(),
            wantErr: false,
        },
        {
            name:    "invalid ISBN",
            book:    fixtures.BookWithInvalidISBN(),
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := service.Validate(tt.book)

            if tt.wantErr {
                testutil.AssertError(t, err)
            } else {
                testutil.AssertNoError(t, err)
            }
        })
    }
}
```

### Pattern 2: Setup/Teardown with t.Cleanup

**Best for**: Tests that need resource cleanup

```go
func TestWithDatabase(t *testing.T) {
    db, err := setupDatabase()
    testutil.AssertNoError(t, err)

    t.Cleanup(func() {
        db.Close()
    })

    // Test code using db
}
```

### Pattern 3: Subtests

**Best for**: Grouping related test cases

```go
func TestBookOperations(t *testing.T) {
    t.Run("Create", func(t *testing.T) {
        // Test creation
    })

    t.Run("Update", func(t *testing.T) {
        // Test update
    })

    t.Run("Delete", func(t *testing.T) {
        // Test deletion
    })
}
```

### Pattern 4: Mocking with Expectations

**Best for**: Verifying dependencies are called correctly

```go
func TestCreateBookUseCase(t *testing.T) {
    mockRepo := mocks.NewMockRepository(t)
    mockCache := mocks.NewMockCache(t)

    // Setup expectations in order
    mockRepo.EXPECT().
        Get(mock.Anything, "isbn").
        Return(book.Book{}, errors.ErrNotFound).
        Once()

    mockRepo.EXPECT().
        Add(mock.Anything, mock.Anything).
        Return("book-id", nil).
        Once()

    mockCache.EXPECT().
        Set(mock.Anything, "book-id", mock.Anything).
        Return(nil).
        Once()

    // Execute use case
    result, err := uc.Execute(ctx, request)

    // Assertions auto-verified by mock cleanup
    testutil.AssertNoError(t, err)
}
```

## Test Coverage Goals

| Layer | Target Coverage | Current |
|-------|----------------|---------|
| Domain | 100% | ~80% |
| Use Cases | 80%+ | ~60% |
| Repositories | 80%+ | ~40% |
| Handlers | 60%+ | ~20% |
| Overall | 70%+ | ~50% |

**Generate Coverage Report**:
```bash
make test-coverage
open coverage.html
```

## Best Practices

### ✅ Do

- **Use fixtures** for test data instead of inline struct literals
- **Use testutil** for assertions instead of verbose if-else blocks
- **Use table-driven tests** for multiple similar scenarios
- **Mock at boundaries** (repositories, external services)
- **Test failure paths** not just success cases
- **Use subtests** for grouping related tests
- **Clean up resources** with `t.Cleanup()`
- **Test business logic** thoroughly in domain layer

### ❌ Don't

- **Don't mock domain entities** - use real entities or fixtures
- **Don't test implementation details** - test behavior
- **Don't share test state** - each test should be independent
- **Don't skip cleanup** - use `t.Cleanup()` or `defer`
- **Don't ignore errors in test setup** - fail early
- **Don't write brittle tests** - avoid hardcoding exact error messages
- **Don't mock everything** - only mock external dependencies

## Common Issues

### Issue: "Mock expectations not met"

**Cause**: Mock was called differently than expected

**Solution**:
```go
// Use mock.Anything for arguments you don't care about
mockRepo.EXPECT().
    Get(mock.Anything, "book-id").  // ✅ Flexible context
    Return(book, nil)

// Or use specific matchers
mockRepo.EXPECT().
    Get(mock.Anything, mock.MatchedBy(func(id string) bool {
        return strings.HasPrefix(id, "book-")
    })).
    Return(book, nil)
```

### Issue: "Database already exists" in integration tests

**Cause**: Previous test didn't clean up database

**Solution**:
```go
func TestIntegration(t *testing.T) {
    db, cleanup := setupTestDB(t)
    defer cleanup()  // Always defer cleanup

    // Test code
}
```

### Issue: Tests pass individually but fail together

**Cause**: Shared state or database conflicts

**Solution**:
- Use `t.Parallel()` to detect race conditions
- Use unique test data (UUIDs, timestamps)
- Clean up database between tests

### Issue: Flaky tests

**Causes**: Time-dependent logic, race conditions, external dependencies

**Solutions**:
- Mock time-dependent code
- Use `go test -race` to detect races
- Mock external dependencies
- Use deterministic test data

## Generating Test Coverage

### HTML Coverage Report
```bash
make test-coverage
# Opens coverage.html in browser
```

### Terminal Coverage
```bash
go test -cover ./...
```

### Package-Specific Coverage
```bash
go test -coverprofile=coverage.out ./internal/domain/book/...
go tool cover -func=coverage.out
```

### Coverage Breakdown
```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep -E 'total|domain|usecase'
```

## CI/CD Integration

Tests run automatically in GitHub Actions on:
- Pull requests
- Pushes to main/develop branches

**Workflow**: `.github/workflows/ci.yml`

**Required Checks**:
- ✅ All tests pass
- ✅ Coverage ≥ 60%
- ✅ No linter errors
- ✅ Integration tests pass (PR only)

## Related Documentation

- [Testing Guide](../.claude/testing.md) - Comprehensive testing strategies
- [Development Workflows](../.claude/development-workflows.md) - Complete workflows
- [Code Standards](../.claude/standards.md) - Go best practices
- [Architecture](../.claude/architecture.md) - System design

## Migration Notes

**Before Phase 7.3** (Scattered test infrastructure):
- No centralized fixtures
- Inline test data everywhere
- 11 scattered mock structs
- Verbose assertion code

**After Phase 7.3** (Centralized infrastructure):
- Fixtures: `test/fixtures/` with 4 entity types
- Testutil: `test/testutil/` with 12 assertions + context helpers
- Mocks: `test/mocks/` for integration + auto-generated in `internal/adapters/repository/mocks/`
- Clear documentation and examples

**Migration Guide**:
```go
// Before
var req dto.CreateBookRequest
req.Name = "Test Book"
req.ISBN = "978-0134190440"
if err != nil {
    t.Fatalf("unexpected error: %v", err)
}
if result.ID != expected.ID {
    t.Errorf("got %s, want %s", result.ID, expected.ID)
}

// After
req := fixtures.CreateBookRequest()
testutil.AssertNoError(t, err)
testutil.AssertEqual(t, result.ID, expected.ID)
```
