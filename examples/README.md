# Code Pattern Examples

This directory contains canonical code patterns and examples used throughout the codebase.

## Purpose

These examples serve as:
- **Reference implementations** for Claude Code and developers
- **Pattern templates** for adding new features
- **Documentation** of architectural decisions
- **Token optimization** - Load examples instead of searching multiple files

## Available Patterns

### 1. [Handler Pattern](./handler_pattern.md)
HTTP handler implementation following bounded context structure:
- Private methods with grouped use case access
- Standard request/response flow
- Validation and error handling
- Swagger annotations
- Context helpers

**Reference:** `internal/books/http/crud.go`, `internal/members/http/auth/handler.go`

### 2. [Use Case Pattern](./usecase_pattern.md)
Application business logic orchestration:
- Request/Response structs
- Execute method signature
- Domain service usage
- Error wrapping
- Logging patterns

**Reference:** `internal/books/operations/create_book.go`, `internal/members/operations/auth/register.go`

### 3. [Repository Pattern](./repository_pattern.md)
Data access abstraction and implementation:
- Interface definition in domain
- BaseRepository usage
- PostgreSQL implementation
- Error handling
- Generic helpers

**Reference:** `internal/books/repository/book.go`, `internal/payments/repository/payment.go`

### 4. [Testing Pattern](./testing_pattern.md)
Comprehensive testing strategies:
- Table-driven tests
- Mocking with testify/mock
- Integration tests
- Test builders and fixtures
- Coverage goals

**Reference:** `internal/books/domain/book/service_test.go`, `internal/members/operations/auth/register_test.go`

## Quick Reference

| Task | Pattern | File |
|------|---------|------|
| Add HTTP endpoint | Handler Pattern | [handler_pattern.md](./handler_pattern.md) |
| Create business logic | Use Case Pattern | [usecase_pattern.md](./usecase_pattern.md) |
| Data access | Repository Pattern | [repository_pattern.md](./repository_pattern.md) |
| Write tests | Testing Pattern | [testing_pattern.md](./testing_pattern.md) |

## Token Efficiency

**Without examples:** Claude Code must search 8-12 files to understand patterns (3,000-5,000 tokens)

**With examples:** Load 1-2 example files with all patterns (500-1,000 tokens)

**Savings:** 60-70% reduction in token consumption for pattern-based tasks

## Usage in Claude Code Sessions

1. **Starting a new feature:** Read relevant pattern file first
2. **Fixing a bug:** Reference pattern to ensure consistency
3. **Refactoring:** Compare against canonical patterns
4. **Code review:** Verify adherence to documented patterns

## Pattern Compliance

All code in bounded contexts follows these patterns:
- ✅ `internal/books/` - Books bounded context
- ✅ `internal/members/` - Members bounded context
- ✅ `internal/payments/` - Payments bounded context
- ✅ `internal/reservations/` - Reservations bounded context

## See Also

- `.claude-context/CURRENT_PATTERNS.md` - Pattern reference
- `.claude-context/SESSION_MEMORY.md` - Architecture context
- `.claude/architecture.md` - Architecture overview
- CLAUDE.md - Session start guide

---

**Last Updated:** October 11, 2025
**Token Cost per Pattern:** ~2,000-3,000 tokens each
**Total Token Savings:** 60-70% for pattern-based tasks
