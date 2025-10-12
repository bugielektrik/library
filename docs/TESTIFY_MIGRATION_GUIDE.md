# Testify Migration Guide

This guide helps you migrate existing tests from custom assertion helpers to the industry-standard `testify` library.

## Why Testify?

- ✅ Industry standard (17,431+ packages use it)
- ✅ Better error messages with automatic diffs
- ✅ Both `assert` (continue on fail) and `require` (stop on fail) variants
- ✅ Rich assertion library (100+ assertion functions)
- ✅ Test suite support for setup/teardown
- ✅ Mock generation capabilities

## Quick Reference

### Import Packages

```go
import (
    "testing"

    "github.com/stretchr/testify/assert"  // Continues test on failure
    "github.com/stretchr/testify/require" // Stops test on failure
)
```

### Migration Mapping

| Old (Custom) | New (Testify) | Notes |
|--------------|---------------|-------|
| `testutil.AssertEqual(t, expected, actual)` | `assert.Equal(t, expected, actual)` | Note: argument order is **expected, actual** |
| `testutil.AssertNoError(t, err)` | `require.NoError(t, err)` | Use `require` for critical checks |
| `testutil.AssertError(t, err)` | `assert.Error(t, err)` | |
| `testutil.AssertTrue(t, condition)` | `assert.True(t, condition)` | |
| `testutil.AssertFalse(t, condition)` | `assert.False(t, condition)` | |
| `testutil.AssertNil(t, value)` | `assert.Nil(t, value)` | |
| `testutil.AssertNotNil(t, value)` | `assert.NotNil(t, value)` | |
| `testutil.AssertStringContains(t, s, substr)` | `assert.Contains(t, s, substr)` | Works with strings, slices, maps |
| `testutil.AssertPanic(t, fn)` | `assert.Panics(t, fn)` | |

## Assert vs Require

**Rule of thumb:**
- Use `require` for critical assertions (if failed, test can't continue)
- Use `assert` for non-critical assertions (collect all failures)

```go
func TestBookRepository(t *testing.T) {
    // Use require for setup that must succeed
    book, err := repo.GetByID(ctx, bookID)
    require.NoError(t, err)          // Stop if book not found
    require.NotNil(t, book)           // Stop if book is nil

    // Use assert for validation checks
    assert.Equal(t, "Clean Code", book.Name)     // Continue even if name doesn't match
    assert.Equal(t, "Tech", book.Genre)          // Check genre too
    assert.Len(t, book.Authors, 2)               // Check author count
}
```

## Migration Examples

### Before (Custom Helpers)

```go
package book_test

import (
    "testing"
    "library-service/test/testutil"
)

func TestBookService_ValidateISBN(t *testing.T) {
    tests := []struct {
        name    string
        isbn    string
        wantErr bool
    }{
        {"valid ISBN-13", "978-0-306-40615-7", false},
        {"invalid checksum", "978-0-306-40615-8", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateISBN(tt.isbn)
            if tt.wantErr {
                testutil.AssertError(t, err)
            } else {
                testutil.AssertNoError(t, err)
            }
        })
    }
}
```

### After (Testify)

```go
package book_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestBookService_ValidateISBN(t *testing.T) {
    tests := []struct {
        name    string
        isbn    string
        wantErr bool
    }{
        {"valid ISBN-13", "978-0-306-40615-7", false},
        {"invalid checksum", "978-0-306-40615-8", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateISBN(tt.isbn)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

## Advanced Features

### Custom Error Messages

```go
// Add custom message to any assertion
assert.Equal(t, expected, actual, "Book ID should match after creation")

// Use formatted messages
assert.True(t, book.IsAvailable(), "Book %s should be available", book.ID)
```

### Rich Assertions

```go
// Check specific error types
assert.ErrorIs(t, err, errors.ErrNotFound)

// Check error contains string
assert.ErrorContains(t, err, "book not found")

// Check subsets
assert.Subset(t, []string{"a", "b", "c"}, []string{"a", "b"})

// Check JSON equality
expected := `{"name": "Clean Code", "isbn": "978-0132350884"}`
assert.JSONEq(t, expected, actualJSON)

// Check element in slice
assert.Contains(t, bookIDs, "book-123")

// Check length
assert.Len(t, books, 5)

// Check empty
assert.Empty(t, errors)
assert.NotEmpty(t, results)

// Numeric comparisons
assert.Greater(t, actualCount, 0)
assert.GreaterOrEqual(t, balance, minBalance)
assert.Less(t, usage, limit)
```

### Test Suites

```go
package book_test

import (
    "testing"
    "github.com/stretchr/testify/suite"
)

type BookServiceTestSuite struct {
    suite.Suite
    service *BookService
    repo    *mocks.MockBookRepository
}

// SetupTest runs before each test
func (s *BookServiceTestSuite) SetupTest() {
    s.repo = mocks.NewMockBookRepository(s.T())
    s.service = NewBookService(s.repo)
}

// TearDownTest runs after each test
func (s *BookServiceTestSuite) TearDownTest() {
    s.repo.AssertExpectations(s.T())
}

// Test methods must start with "Test"
func (s *BookServiceTestSuite) TestGetBook_Success() {
    // Use s.Assert() or s.Require()
    book, err := s.service.GetBook(ctx, "book-1")
    s.Require().NoError(err)
    s.Assert().Equal("Clean Code", book.Name)
}

func (s *BookServiceTestSuite) TestGetBook_NotFound() {
    s.repo.On("GetByID", mock.Anything, "invalid-id").Return(nil, errors.ErrNotFound)

    book, err := s.service.GetBook(ctx, "invalid-id")
    s.Assert().Error(err)
    s.Assert().Nil(book)
}

// Run the suite
func TestBookServiceTestSuite(t *testing.T) {
    suite.Run(t, new(BookServiceTestSuite))
}
```

## Common Patterns

### HTTP Handler Testing

```go
func TestBookHandler_GetBook(t *testing.T) {
    // Setup
    req := httptest.NewRequest("GET", "/api/v1/books/book-1", nil)
    w := httptest.NewRecorder()

    // Execute
    handler.GetBook(w, req)

    // Assert response
    assert.Equal(t, http.StatusOK, w.Code)

    var response dto.BookResponse
    err := json.NewDecoder(w.Body).Decode(&response)
    require.NoError(t, err)

    assert.Equal(t, "book-1", response.ID)
    assert.Equal(t, "Clean Code", response.Name)
}
```

### Repository Testing

```go
func TestBookRepository_Create(t *testing.T) {
    // Setup
    ctx := context.Background()
    book := fixtures.ValidBook()

    // Execute
    createdBook, err := repo.Create(ctx, book)

    // Assert
    require.NoError(t, err)
    require.NotNil(t, createdBook)

    assert.NotEmpty(t, createdBook.ID)
    assert.Equal(t, book.Name, createdBook.Name)
    assert.Equal(t, book.ISBN, createdBook.ISBN)
    assert.WithinDuration(t, time.Now(), createdBook.CreatedAt, time.Second)
}
```

### Use Case Testing

```go
func TestCreateBookUseCase_Execute(t *testing.T) {
    // Setup mocks
    bookRepo := bookmocks.NewMockBookRepository(t)
    authorRepo := authormocks.NewMockAuthorRepository(t)

    useCase := NewCreateBookUseCase(bookRepo, authorRepo)

    // Setup expectations
    bookRepo.On("Create", mock.Anything, mock.AnythingOfType("book.Book")).
        Return(&book.Book{ID: "new-id"}, nil)

    req := CreateBookRequest{
        Name:    "Clean Code",
        ISBN:    "978-0132350884",
        Authors: []string{"author-1"},
    }

    // Execute
    resp, err := useCase.Execute(context.Background(), req)

    // Assert
    require.NoError(t, err)
    assert.Equal(t, "new-id", resp.ID)
    assert.Equal(t, "Clean Code", resp.Name)

    // Verify mock expectations
    bookRepo.AssertExpectations(t)
}
```

## Migration Script

```bash
#!/bin/bash
# scripts/migrate-to-testify.sh

# Find all test files using old assertions
files=$(find . -name "*_test.go" -type f -exec grep -l "testutil.Assert\|helpers.Assert" {} \;)

echo "Found $(echo "$files" | wc -l) files to update"
echo ""
echo "Manual steps required:"
echo "1. Update imports:"
echo "   - Remove: \"library-service/test/testutil\""
echo "   - Remove: \"library-service/test/helpers\""
echo "   + Add: \"github.com/stretchr/testify/assert\""
echo "   + Add: \"github.com/stretchr/testify/require\" (if needed)"
echo ""
echo "2. Replace assertions (preserve argument order!):"
echo "   testutil.AssertEqual     -> assert.Equal"
echo "   testutil.AssertNoError   -> require.NoError (or assert.NoError)"
echo "   testutil.AssertError     -> assert.Error"
echo "   etc."
echo ""
echo "Files to update:"
echo "$files"
```

## Tips

1. **Start Small**: Migrate one test file at a time
2. **Run Tests**: Run `go test` after each file to catch issues early
3. **Use require for Setup**: If setup fails, test can't continue - use `require`
4. **Use assert for Validation**: Collect all failures in validation - use `assert`
5. **Add Messages**: Use custom messages for complex assertions
6. **Leverage Suites**: For tests with common setup, use test suites

## Resources

- [Testify Documentation](https://pkg.go.dev/github.com/stretchr/testify)
- [Assert Package](https://pkg.go.dev/github.com/stretchr/testify/assert)
- [Require Package](https://pkg.go.dev/github.com/stretchr/testify/require)
- [Suite Package](https://pkg.go.dev/github.com/stretchr/testify/suite)
- [Mock Package](https://pkg.go.dev/github.com/stretchr/testify/mock)

## Next Steps

1. Review this guide
2. Pick a simple test file to start with
3. Update imports and assertions
4. Run tests to verify
5. Repeat for remaining files
6. Delete old custom assertion files once migration is complete

---

*Generated as part of the library refactoring project*
