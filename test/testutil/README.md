# Test Utilities

This package provides reusable test helper functions for writing cleaner, more readable tests.

## Purpose

Test utilities reduce boilerplate in test code and make tests more expressive. Instead of writing verbose if-else blocks, use these helpers.

## Available Utilities

### Assertions (`assertions.go`)

#### Error Assertions
- `AssertNoError(t, err)` - Fails if err is not nil
- `AssertError(t, err)` - Fails if err is nil

#### Value Assertions
- `AssertEqual(t, got, want)` - Fails if values are not equal (uses reflect.DeepEqual)
- `AssertNotEqual(t, got, want)` - Fails if values are equal
- `AssertNil(t, value)` - Fails if value is not nil
- `AssertNotNil(t, value)` - Fails if value is nil

#### Boolean Assertions
- `AssertTrue(t, condition)` - Fails if condition is false
- `AssertFalse(t, condition)` - Fails if condition is true

#### String Assertions
- `AssertStringContains(t, s, substr)` - Fails if s does not contain substr
- `AssertStringNotContains(t, s, substr)` - Fails if s contains substr

#### Panic Assertions
- `AssertPanic(t, fn)` - Fails if fn does not panic
- `AssertNoPanic(t, fn)` - Fails if fn panics

### Context Helpers (`context.go`)

- `NewContext()` - Creates a context with 5-second timeout
- `NewContextWithTimeout(duration)` - Creates a context with custom timeout
- `NewContextWithValue(key, value)` - Creates a context with a key-value pair

## Usage Examples

### Basic Assertions

```go
import "library-service/test/testutil"

func TestBookService(t *testing.T) {
    result, err := bookService.GetBook(ctx, "book-id")

    // Instead of:
    // if err != nil {
    //     t.Fatalf("unexpected error: %v", err)
    // }
    testutil.AssertNoError(t, err)

    // Instead of:
    // if result.ID != "book-id" {
    //     t.Errorf("got ID %s, want book-id", result.ID)
    // }
    testutil.AssertEqual(t, result.ID, "book-id")
}
```

### Table-Driven Tests

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
            err := bookService.Validate(tt.book)

            if tt.wantErr {
                testutil.AssertError(t, err)
            } else {
                testutil.AssertNoError(t, err)
            }
        })
    }
}
```

### Context Usage

```go
func TestWithContext(t *testing.T) {
    // Create a test context with timeout
    ctx := testutil.NewContext()

    result, err := useCase.Execute(ctx, request)
    testutil.AssertNoError(t, err)
}

func TestWithCustomTimeout(t *testing.T) {
    // Create a context with longer timeout
    ctx := testutil.NewContextWithTimeout(30 * time.Second)

    result, err := longRunningOperation(ctx)
    testutil.AssertNoError(t, err)
}

func TestWithContextValue(t *testing.T) {
    // Create a context with a specific value
    ctx := testutil.NewContextWithValue("member_id", "123")

    result, err := authService.Authorize(ctx)
    testutil.AssertNoError(t, err)
}
```

### String Assertions

```go
func TestErrorMessage(t *testing.T) {
    err := someOperation()
    testutil.AssertError(t, err)
    testutil.AssertStringContains(t, err.Error(), "not found")
}
```

### Panic Testing

```go
func TestPanicBehavior(t *testing.T) {
    // Test that a function panics
    testutil.AssertPanic(t, func() {
        riskyOperation(nil) // Should panic
    })

    // Test that a function does not panic
    testutil.AssertNoPanic(t, func() {
        safeOperation("valid-input")
    })
}
```

## Benefits

- ✅ **Less Boilerplate**: One-line assertions instead of 3-5 lines of if-else
- ✅ **More Readable**: `AssertNoError(t, err)` is clearer than verbose error checking
- ✅ **Consistent**: Same assertion style across all tests
- ✅ **Better Error Messages**: Assertions provide clear failure messages
- ✅ **t.Helper()**: All assertions mark themselves as helpers for better stack traces

## Design Principles

1. **Simple and Focused**: Each helper does one thing well
2. **No Dependencies**: Pure Go stdlib, no external testing frameworks
3. **t.Helper() Always**: All assertions call t.Helper() for clean stack traces
4. **Fail Fast**: Use t.Fatalf() for critical failures, t.Errorf() for comparisons
5. **Descriptive Names**: Function names clearly indicate what they assert

## Adding New Helpers

When adding new helpers:

1. Add to appropriate file (assertions.go, context.go, etc.)
2. Always call `t.Helper()` at the start
3. Provide clear error messages
4. Document in this README
5. Write tests for the helper itself

## Comparison with Other Frameworks

This package provides similar functionality to popular testing frameworks like testify/assert, but:

- ✅ Zero external dependencies
- ✅ Lightweight and fast
- ✅ Project-specific (can add domain-specific helpers)
- ✅ Simple to understand and maintain

Use these helpers for all new tests to maintain consistency across the codebase.
