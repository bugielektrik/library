# Contributing Guidelines

Thank you for contributing to the Library Management System! This guide will help you make effective contributions.

## Code of Conduct

- Be respectful and inclusive
- Focus on constructive feedback
- Help others learn and grow
- Maintain professional communication

## Getting Started

### 1. Fork and Clone

```bash
# Fork the repository on GitHub
git clone https://github.com/yourusername/library.git
cd library
git remote add upstream https://github.com/original/library.git
```

### 2. Set Up Development Environment

```bash
make init             # Download dependencies
make up               # Start services (PostgreSQL, Redis)
make migrate-up       # Run migrations
make test             # Verify setup
```

### 3. Create a Branch

```bash
git checkout -b feature/your-feature-name
# or: git checkout -b fix/bug-description
```

**Branch Naming**:
- `feature/*` - New features
- `fix/*` - Bug fixes
- `refactor/*` - Code improvements
- `docs/*` - Documentation updates
- `test/*` - Test additions

## Development Workflow

### 1. Write Tests First (TDD)

```bash
# Create test file
touch internal/domain/book/service_test.go

# Write failing test
# Implement feature
# Verify test passes
make test
```

### 2. Follow Clean Architecture

```
Domain → Use Case → Adapters → Infrastructure
```

**Rules**:
- Domain layer has no external dependencies
- Use cases orchestrate domain logic
- Adapters handle external communication
- Dependency injection for testability

### 3. Code Quality Checks

```bash
make fmt              # Format code
make vet              # Run go vet
make lint             # Run golangci-lint
make test             # Run all tests
make ci               # Full CI pipeline
```

**Before committing**:
```bash
make ci               # Must pass locally
```

## Coding Standards

### File Organization

```go
// 1. Package declaration
package book

// 2. Imports (grouped: stdlib, external, internal)
import (
    "context"
    "time"

    "github.com/google/uuid"

    "library-service/pkg/errors"
)

// 3. Constants and variables
const MaxBookTitleLength = 255

// 4. Types
type Entity struct { ... }

// 5. Constructor
func NewEntity() *Entity { ... }

// 6. Methods
func (e *Entity) Method() { ... }
```

### Naming Conventions

**Packages**: Lowercase, singular (`book`, not `books`)
**Files**: Snake_case for multi-word (`create_book.go`)
**Types**: PascalCase (`BookEntity`)
**Functions**: PascalCase for exported, camelCase for private
**Interfaces**: `-er` suffix (`Repository`, `Validator`)

### Error Handling

```go
// DO: Wrap errors with context
if err := repo.Create(ctx, book); err != nil {
    return fmt.Errorf("failed to create book: %w", err)
}

// DON'T: Swallow errors
if err := repo.Create(ctx, book); err != nil {
    log.Println(err) // Bad: error lost
}
```

### Testing Standards

**Coverage Requirements**:
- Domain services: 100%
- Use cases: 80%+
- Adapters: 60%+
- Overall: 60%+

**Test Structure**:
```go
func TestService_Method(t *testing.T) {
    // Arrange
    service := NewService()
    input := "test-input"

    // Act
    result, err := service.Method(input)

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "expected", result)
}
```

**Table-Driven Tests**:
```go
func TestValidateISBN(t *testing.T) {
    tests := []struct {
        name    string
        isbn    string
        wantErr bool
    }{
        {"valid ISBN-10", "0-306-40615-2", false},
        {"invalid", "bad-isbn", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateISBN(tt.isbn)
            if (err != nil) != tt.wantErr {
                t.Errorf("got error %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## Commit Message Format

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types**:
- `feat`: New feature
- `fix`: Bug fix
- `refactor`: Code improvement
- `test`: Test additions
- `docs`: Documentation
- `chore`: Maintenance tasks
- `perf`: Performance improvement

**Examples**:

```
feat(book): add ISBN validation to domain service

Implement ISBN-10 and ISBN-13 validation with checksum verification
to ensure data integrity at the domain level.

Closes #123
```

```
fix(member): correct subscription expiration calculation

Previous calculation didn't account for leap years. Now using
time.AddDate() for accurate month arithmetic.

Fixes #456
```

**Scope**: Domain name (`book`, `member`, `author`) or layer (`domain`, `usecase`, `http`)

## Pull Request Process

### 1. Prepare Your PR

```bash
# Sync with upstream
git fetch upstream
git rebase upstream/main

# Ensure quality
make ci

# Push to your fork
git push origin feature/your-feature
```

### 2. Create Pull Request

**Title**: Follow commit message format
```
feat(book): add ISBN validation
```

**Description Template**:
```markdown
## Summary
Brief description of changes (1-3 sentences)

## Changes
- Added ISBN validation service
- Implemented checksum verification
- Added comprehensive tests

## Test Plan
- [ ] Unit tests pass (100% coverage)
- [ ] Integration tests pass
- [ ] Manual testing performed
- [ ] Linter passes

## Screenshots (if applicable)

## Related Issues
Closes #123
```

### 3. Code Review

**What Reviewers Check**:
- Clean architecture principles followed
- Tests cover new code (with edge cases)
- No breaking changes (or documented)
- Code follows project conventions
- Documentation updated

**Responding to Feedback**:
```bash
# Make requested changes
git add .
git commit -m "refactor: address review feedback"
git push origin feature/your-feature
```

### 4. Merge Requirements

- ✅ All CI checks pass (lint, test, build, security)
- ✅ Code review approved (1+ approver)
- ✅ No merge conflicts
- ✅ Documentation updated
- ✅ Tests added/updated

## Feature Development Checklist

### Domain Layer
- [ ] Create/update entity in `internal/domain/{domain}/entity.go`
- [ ] Add business rules in `internal/domain/{domain}/service.go`
- [ ] Define repository interface in `internal/domain/{domain}/repository.go`
- [ ] Write unit tests (100% coverage)

### Use Case Layer
- [ ] Create use case in `internal/usecase/{domain}/{action}.go`
- [ ] Define DTOs in `internal/usecase/{domain}/dto.go`
- [ ] Implement business flow
- [ ] Write unit tests with mocks (80%+ coverage)

### Adapter Layer
- [ ] Create HTTP handler in `internal/adapters/http/{domain}/handler.go`
- [ ] Define request/response DTOs
- [ ] Implement repository if needed
- [ ] Add integration tests

### Infrastructure
- [ ] Update routes in `cmd/api/main.go`
- [ ] Add migrations if database changes
- [ ] Update API documentation (Swagger)
- [ ] Update architecture docs if needed

## Documentation

### When to Update Docs

- **New Feature**: Update README.md, add usage examples
- **Architecture Change**: Update docs/architecture.md, create ADR
- **API Changes**: Regenerate Swagger (`make gen-docs`)
- **Breaking Changes**: Update CHANGELOG.md, migration guide

### Documentation Files

- `README.md` - Project overview
- `docs/architecture.md` - Architecture details
- `docs/guides/QUICKSTART.md` - Quick start guide
- `docs/guides/DEVELOPMENT.md` - Development guide (this file)
- `docs/adr/` - Architecture Decision Records

## Database Migrations

### Creating Migrations

```bash
make migrate-create name=add_books_isbn_index

# Creates:
# migrations/000X_add_books_isbn_index.up.sql
# migrations/000X_add_books_isbn_index.down.sql
```

### Migration Guidelines

- **Backward Compatible**: Avoid breaking existing data
- **Reversible**: Always implement `.down.sql`
- **Tested**: Test both up and down migrations
- **Idempotent**: Use `IF NOT EXISTS` where possible

```sql
-- Good: Idempotent
CREATE TABLE IF NOT EXISTS books (
    id UUID PRIMARY KEY
);

-- Good: Backward compatible
ALTER TABLE books ADD COLUMN IF NOT EXISTS isbn VARCHAR(13);

-- Bad: Breaking change without migration plan
ALTER TABLE books DROP COLUMN title;
```

## Security Guidelines

- **No Secrets in Code**: Use environment variables
- **Input Validation**: Validate all user input
- **SQL Injection**: Use parameterized queries (sqlx handles this)
- **Error Messages**: Don't expose internal details
- **Dependencies**: Keep updated (`make mod-update`)

```bash
# Security scanning
make security         # Run gosec
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...
```

## Performance Guidelines

- **Database**: Add indexes for frequently queried columns
- **Caching**: Use Redis for read-heavy data
- **Pagination**: Implement for list endpoints
- **Benchmarks**: Add for critical paths

```bash
make benchmark        # Run performance benchmarks
```

## Getting Help

- **Questions**: Open a GitHub Discussion
- **Bugs**: Open an issue with reproduction steps
- **Features**: Open an issue with use case description
- **Urgent**: Contact maintainers (see README.md)

## Recognition

Contributors are recognized in:
- `CONTRIBUTORS.md` file
- Release notes
- GitHub contributions graph

Thank you for making this project better! 🎉

## Quick Reference

```bash
# Setup
make init && make up && make migrate-up

# Development
make dev              # Start development environment
make test             # Run tests
make lint             # Run linters

# Before commit
make ci               # Full quality check

# Before PR
git fetch upstream && git rebase upstream/main
make ci
```

**Happy coding!** 🚀
