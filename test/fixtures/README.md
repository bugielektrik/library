# Test Fixtures

This package provides reusable test fixtures for all domain entities in the library management system.

## Purpose

Test fixtures help create consistent, realistic test data across unit tests, integration tests, and benchmarks. Instead of duplicating entity creation logic in every test, import these fixtures.

## Usage

```go
import "library-service/test/fixtures"

func TestBookService(t *testing.T) {
    // Use a valid book fixture
    book := fixtures.ValidBook()

    // Or use a specialized fixture
    bookWithMultipleAuthors := fixtures.ValidBookWithMultipleAuthors()

    // Test your logic
    result := bookService.Process(book)
    // ...
}
```

## Available Fixtures

### Book Fixtures (`book.go`)
- `ValidBook()` - Standard valid book
- `ValidBookWithMultipleAuthors()` - Book with 3 authors
- `MinimalBook()` - Book with only required fields
- `BookWithInvalidISBN()` - For testing validation
- `CreateBookRequest()` - Use case request
- `UpdateBookRequest()` - Use case request
- `BookResponse()` - Domain response
- `BookResponses()` - List of responses

### Author Fixtures (`author.go`)
- `ValidAuthor()` - Standard author with pseudonym
- `AuthorWithoutPseudonym()` - Author without pseudonym
- `MinimalAuthor()` - Author with only required fields
- `AuthorResponse()` - Domain response
- `AuthorResponses()` - List of responses

### Member Fixtures (`member.go`)
- `ValidMember()` - Standard member with active subscription
- `AdminMember()` - Member with admin role
- `MemberWithoutSubscription()` - Member without subscription
- `ExpiredSubscriptionMember()` - Member with expired subscription
- `MemberResponse()` - Domain response
- `MemberResponses()` - List of responses

### Reservation Fixtures (`reservation.go`)
- `ValidReservation()` - Active reservation
- `PendingReservation()` - Pending reservation
- `CancelledReservation()` - Cancelled reservation
- `ExpiredReservation()` - Expired reservation
- `ReservationResponse()` - Domain response
- `ReservationResponses()` - List of responses

## Design Principles

1. **Realistic Data**: Fixtures use realistic values (e.g., real ISBNs, proper UUIDs)
2. **Consistent IDs**: Same entity IDs are reused across fixtures for relationships
3. **Multiple Variants**: Each entity has multiple variants for different test scenarios
4. **Immutable**: Fixtures return new instances, never modify shared state
5. **Self-Contained**: Each fixture is complete and valid by default

## ID Conventions

Consistent IDs are used to maintain relationships:

- Books: `550e8400-e29b-41d4-a716-44665544000X`
- Authors: `550e8400-e29b-41d4-a716-44665544000X`
- Members: `[a-d]4101570-0a35-4dd3-b8f7-745d5601326X`
- Reservations: `r4101570-0a35-4dd3-b8f7-745d5601326X`

## Adding New Fixtures

When adding new fixtures:

1. Create a new file in `test/fixtures/` named after the domain entity
2. Follow the naming convention: `Valid{Entity}()`, `{Entity}With{Variant}()`
3. Use `strutil.SafeStringPtr()` for optional string fields
4. Include both entity and response fixtures
5. Document the fixture in this README

## Examples

### Testing a Use Case

```go
func TestCreateBookUseCase(t *testing.T) {
    // Arrange
    mockRepo := &mocks.MockBookRepository{
        AddFunc: func(ctx context.Context, b book.Book) (string, error) {
            return "new-id", nil
        },
    }

    uc := bookops.NewCreateBookUseCase(mockRepo)
    req := fixtures.CreateBookRequest()

    // Act
    result, err := uc.Execute(context.Background(), req)

    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if result.ID == "" {
        t.Error("expected non-empty ID")
    }
}
```

### Testing Edge Cases

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
            if (err != nil) != tt.wantErr {
                t.Errorf("got error %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## Benefits

- ✅ **Consistency**: Same test data across all tests
- ✅ **Maintainability**: Update fixtures once, all tests benefit
- ✅ **Readability**: `fixtures.ValidBook()` is more readable than inline construction
- ✅ **Speed**: Write tests faster without duplicating setup code
- ✅ **Relationships**: Consistent IDs make it easy to test entity relationships
