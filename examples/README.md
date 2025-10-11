# Canonical Code Examples

This directory contains canonical examples of code patterns used in this project. **Claude Code should reference these examples** when creating new code to ensure consistency and token efficiency.

## Purpose

These examples serve as:
1. **Reference patterns** for AI coding assistants
2. **Documentation** of project conventions
3. **Training examples** for new developers
4. **Token-efficient context** (load 1 example file instead of searching 10+ actual files)

## Available Examples

### 1. Handler Example (`handler_example.go`)

**Token cost:** ~600 tokens

Demonstrates:
- HTTP handler structure and patterns
- Private methods (lowercase)
- Request/response DTOs
- Validation with `h.validator.ValidateStruct()`
- Logging with `logutil.HandlerLogger()`
- Error handling with `h.RespondError()`
- CRUD operations (create, get, update, delete, list)

**Use when:** Creating or modifying HTTP endpoints

### 2. Use Case Example (`usecase_example.go`)

**Token cost:** ~700 tokens

Demonstrates:
- Use case structure and patterns
- Single responsibility principle
- Repository interface dependencies
- Proper error wrapping
- Logging with `logutil.UseCaseLogger()`
- Helper methods for complex logic
- Request/Response patterns

**Use when:** Creating business logic or use cases

### 3. Repository Example (`repository_example.go`)

**Token cost:** ~750 tokens

Demonstrates:
- Repository interface implementation
- Database operations (CRUD)
- Transaction handling with `UpdateFn` pattern
- Error mapping (DB errors → domain errors)
- Logging with `logutil.RepositoryLogger()`
- SQL best practices

**Use when:** Creating data access layers

### 4. Test Example (`test_example_test.go`)

**Token cost:** ~650 tokens

Demonstrates:
- Table-driven tests
- Mock creation and usage
- Test structure and naming
- Integration vs unit tests
- Assertion patterns
- Benchmark tests

**Use when:** Writing tests for any component

## Usage for Claude Code

### When Creating New Code

1. **Identify the component type** (handler, use case, repository, test)
2. **Read the relevant example** (Claude Code: load only the needed example)
3. **Follow the patterns** shown in the example
4. **Adapt to specific requirements**

### Example Prompt Patterns

**Good prompt (token-efficient):**
```
Create a new handler for managing orders following the pattern in examples/handler_example.go.
Include: create, get, list operations.
```
*Token cost:* ~1,200 tokens (600 for example + 600 for implementation)

**Bad prompt (token-inefficient):**
```
Create a new handler for orders. Look at existing handlers to understand the pattern.
```
*Token cost:* ~4,500 tokens (loading 5-8 different handler files to infer pattern)

### Token Savings

| Approach | Files Loaded | Avg Tokens | Example |
|----------|--------------|------------|---------|
| **With Examples** | 1-2 | 600-1,200 | Load handler_example.go |
| **Without Examples** | 8-12 | 4,000-6,000 | Search through all handlers |
| **Savings** | -85% | -70% | **3,000-5,000 tokens saved** |

## Maintenance

### When to Update Examples

- ✅ When introducing new patterns project-wide
- ✅ When refactoring common code structures
- ✅ When adding new best practices
- ❌ For one-off or experimental patterns
- ❌ For domain-specific logic (belongs in actual code)

### Keeping Examples Current

1. **Review quarterly** - ensure examples match current codebase
2. **Update with refactoring** - when patterns change, update examples immediately
3. **Keep token-efficient** - each example should be 500-800 tokens max
4. **One pattern per file** - don't mix handler + repository in one example

## Anti-Patterns (DO NOT DO)

❌ **Don't copy examples verbatim** - adapt to your specific use case
❌ **Don't skip examples** - they save 70% of tokens vs. searching codebase
❌ **Don't create example variants** - keep one canonical example per pattern
❌ **Don't add business logic** - examples show structure, not domain logic
❌ **Don't let examples diverge** - update when patterns change

## File Size Targets

- Handlers: ~150 lines = ~300 tokens
- Use Cases: ~200 lines = ~400 tokens
- Repositories: ~250 lines = ~500 tokens
- Tests: ~150 lines = ~300 tokens

**Total:** ~1,500 tokens for all examples vs. ~10,000-15,000 tokens searching codebase

## Related Documentation

- `.claude/` - Comprehensive project documentation
- `CLAUDE.md` - Project overview and guidelines
- `.claude-context/` - Session memory and patterns

---

**Last Updated:** October 11, 2025
**Purpose:** Token-efficient AI coding assistance
