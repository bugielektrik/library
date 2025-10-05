# Effective Claude Code Prompts

> **Prompt patterns that work well with this codebase**

## 🎯 General Principles

### ✅ Good Prompts Are:
- **Specific**: "Add a Loan domain with overdue fee calculation"
- **Context-aware**: "Following Clean Architecture patterns..."
- **Actionable**: "Create X, then Y, then test Z"
- **Layer-conscious**: "Add to domain layer first, then use case, then handler"

### ❌ Avoid:
- **Vague**: "Make it better"
- **Too broad**: "Rewrite the entire system"
- **Layer-violating**: "Add database code to domain layer"
- **Testing-skipping**: "Just implement, skip tests"

## 📚 Prompt Templates

### Adding a New Domain

```
Add a "Loan" domain with the following requirements:
- Track which member borrowed which book
- Track loan date, due date (14 days), and return date
- Calculate late fees ($0.50/day after due date)
- Validate loan duration (1-14 days)

Follow Clean Architecture:
1. Create domain layer with entity, service, repository interface
2. Add use cases for creating loans and returning books
3. Add PostgreSQL repository implementation
4. Add HTTP handlers with Swagger annotations
5. Add complete tests (100% domain coverage)
6. Create database migration

Use the "ops" suffix for use case packages (loanops).
```

### Refactoring Existing Code

```
Refactor the Book domain to follow our standards:
1. Extract ISBN validation from use case to domain service
2. Add proper error wrapping
3. Ensure 100% test coverage
4. Add validation for genre (must be one of: Fiction, Non-Fiction, Technology, Science)
5. Update all affected use cases and tests

Maintain backward compatibility with existing API endpoints.
```

### Adding API Endpoints

```
Add these REST endpoints for loans:
- POST /api/v1/loans - Create a loan (authenticated)
- GET /api/v1/loans/:id - Get loan details (authenticated)
- POST /api/v1/loans/:id/return - Return a book (authenticated)
- GET /api/v1/members/:id/loans - Get member's loan history (authenticated)

Requirements:
- Follow existing handler patterns
- Add complete Swagger annotations with @Security BearerAuth
- Add DTOs with validation tags
- Return appropriate status codes (201 for create, 200 for get, 204 for return)
- Add integration tests
```

### Fixing Bugs

```
Fix the authentication bug where tokens expire immediately:
1. Check JWT token generation in internal/infrastructure/auth/jwt.go
2. Verify expiry duration is set correctly (should be 24h)
3. Add tests to verify token expiry
4. Update Swagger examples if needed

Debug systematically and explain what you find.
```

### Writing Tests

```
Add comprehensive tests for the Book domain service:
1. Test ISBN validation (ISBN-10 and ISBN-13, valid/invalid checksums)
2. Test business rules (can't delete book with active loans)
3. Use table-driven tests
4. Aim for 100% coverage
5. Add edge cases and error scenarios

Show coverage report when done.
```

### Database Operations

```
Create a database migration to add book ratings:
1. Add ratings table (id, book_id, member_id, rating 1-5, comment, created_at)
2. Add foreign keys with ON DELETE CASCADE
3. Add indexes for book_id and member_id
4. Add check constraint (rating between 1 and 5)
5. Include both up and down migrations

Name it: create_ratings_table
```

### Performance Optimization

```
Optimize the book listing endpoint:
1. Identify N+1 query issues
2. Add eager loading for authors
3. Implement pagination (limit 20 per page)
4. Add caching for frequently accessed books
5. Add benchmarks to verify improvement

Show before/after performance metrics.
```

## 🎨 Workflow-Specific Prompts

### Quick Start (New Feature)

```
I want to add [FEATURE]. Please:
1. Read .claude/examples/README.md for patterns
2. Follow the architecture shown in .claude/architecture.md
3. Create domain layer first (entity, service, repository interface, tests)
4. Then use case layer with "ops" suffix
5. Then adapters (repository impl, handlers, DTOs)
6. Wire in container.go
7. Add database migration if needed
8. Update Swagger docs

Ask me questions if requirements are unclear.
```

### Code Review

```
Review my changes in [FILE]:
- Check Clean Architecture compliance
- Verify error handling (errors wrapped with context)
- Check test coverage
- Verify Swagger annotations are complete
- Check for security issues
- Suggest improvements

Be specific about line numbers and show examples.
```

### Troubleshooting

```
I'm getting this error: [ERROR MESSAGE]

Help me debug:
1. Check .claude/troubleshooting.md for known issues
2. Verify my environment (go version, docker status)
3. Check logs if needed
4. Provide step-by-step fix
5. Explain root cause

Be thorough and provide commands I can run.
```

### Documentation

```
Update documentation for the new Loan feature:
1. Add to CLAUDE.md if it's a major feature
2. Update .claude/examples/ with loan examples
3. Add to .claude/recipes.md with common loan operations
4. Update API documentation
5. Ensure Swagger UI reflects changes

Keep documentation concise and practical.
```

## 💡 Context-Specific Prompts

### When Working on Domain Layer

```
I'm working on the [ENTITY] domain. Remember:
- ZERO external dependencies allowed
- Business logic only (no HTTP, database, frameworks)
- Define interfaces here, implement in adapters
- 100% test coverage required
- Use domain errors (errors.ErrNotFound, etc.)

Help me implement [FEATURE] following these constraints.
```

### When Working on Use Cases

```
I'm adding a use case for [OPERATION]. Remember:
- Package name should be [entity]ops (e.g., bookops, not book)
- Depends only on domain interfaces
- Returns domain entities, not DTOs
- Error wrapping: fmt.Errorf("context: %w", err)
- Context as first parameter

Help me implement this correctly.
```

### When Working on HTTP Handlers

```
I'm adding an HTTP handler for [ENDPOINT]. Remember:
- Handler should be thin, delegate to use case
- Complete Swagger annotations required
- @Security BearerAuth for protected endpoints
- DTOs with validation tags
- Appropriate status codes (201 create, 200 get, 204 delete, etc.)

Help me implement this with all annotations.
```

### When Writing Tests

```
I need tests for [COMPONENT]. Remember:
- Table-driven tests preferred
- Descriptive test names: TestFunction_Scenario_ExpectedResult
- Mock external dependencies
- No time.Now() or random values (deterministic tests)
- Test happy path and error cases

Help me write comprehensive tests.
```

## 🚀 Advanced Prompts

### Architecture Verification

```
Verify this code follows Clean Architecture:

[PASTE CODE]

Check:
1. Dependency direction (Domain ← Use Case ← Adapters ← Infrastructure)
2. Domain purity (no external dependencies)
3. Interface placement (defined in domain, implemented in adapters)
4. Package naming (use cases have "ops" suffix)
5. Error handling (wrapped with context)

Point out any violations with specific line numbers.
```

### Performance Analysis

```
Analyze performance of [FEATURE]:
1. Check for N+1 queries
2. Verify indexes exist
3. Check connection pooling
4. Look for memory leaks
5. Suggest optimizations

Show specific code examples and improvements.
```

### Security Audit

```
Security audit for [FEATURE]:
1. Check for SQL injection vulnerabilities
2. Verify authentication on protected routes
3. Check for XSS vulnerabilities
4. Verify input validation
5. Check for hardcoded secrets

Provide specific findings with severity level.
```

## 📋 Multi-Step Workflows

### Complete Feature Implementation

```
Implement book ratings feature end-to-end:

Step 1: Planning
- Review requirements
- Design database schema
- Plan API endpoints

Step 2: Domain Layer
- Create Rating entity
- Add Rating service with validation (1-5 stars)
- Define Repository interface
- Write tests (100% coverage)

Step 3: Database
- Create migration for ratings table
- Add indexes and constraints

Step 4: Use Case Layer
- Create AddRating use case
- Create GetBookRatings use case
- Test with mocks

Step 5: Adapter Layer
- Implement PostgreSQL repository
- Create HTTP handlers
- Add DTOs with validation

Step 6: Integration
- Wire in container.go
- Add routes in router.go
- Update Swagger docs

Step 7: Testing
- Integration tests
- API tests with curl examples

Step 8: Documentation
- Update relevant .claude/ files

Ask for approval before starting each step.
```

### Migration & Refactoring

```
Migrate Book domain from old pattern to new:

Step 1: Analysis
- Read current implementation
- Identify issues with Clean Architecture
- Plan changes

Step 2: Domain Layer
- Extract business logic to domain service
- Add missing tests
- Verify 100% coverage

Step 3: Use Cases
- Rename package to bookops
- Update imports everywhere
- Fix tests

Step 4: Adapters
- Verify repository follows interface
- Update handlers if needed

Step 5: Verification
- Run all tests
- Check linter
- Verify API still works

Step 6: Cleanup
- Remove old code
- Update documentation

Proceed step by step, showing diffs for review.
```

## 🎯 Task-Specific Prompts

### "I Need This Done Fast"

```
Quick implementation of [FEATURE]:
1. Use .claude/examples/ as template
2. Copy-paste patterns from .claude/recipes.md
3. Follow checklist in .claude/checklist.md
4. Skip detailed explanations, just implement

I'll review after you're done.
```

### "I Want to Understand"

```
Teach me how [FEATURE] works:
1. Show me the flow from HTTP request to database and back
2. Explain each layer's responsibility
3. Show me the tests
4. Point me to relevant documentation
5. Provide diagrams if helpful

I want to learn, so explain thoroughly.
```

### "Something's Wrong"

```
Debug this issue: [ISSUE]

1. Check .claude/troubleshooting.md first
2. Run diagnostics
3. Check logs (show me commands)
4. Trace through the code
5. Identify root cause
6. Provide fix with explanation

Be systematic and thorough.
```

## 📖 Example Conversations

### Good Conversation Flow

```
User: Add a Loan domain following Clean Architecture

Claude: I'll add a Loan domain following our architecture. Let me start with:

1. Domain Layer:
- Entity with ID, BookID, MemberID, loan dates
- Service for validation and fee calculation
- Repository interface
- Tests (aiming for 100% coverage)

[Shows code]

Does this match your requirements? Should I proceed to use case layer?

User: Yes, looks good. Continue.

Claude: Great! Now adding use case layer:
- Package: loanops (following our "ops" suffix convention)
- CreateLoan use case
- ReturnBook use case
[Shows code]

Ready to add adapter layer?
```

### Iterative Refinement

```
User: Add rating validation to Book domain

Claude: I'll add rating validation. Currently checking:
- Rating must be 1-5
- Member can rate each book only once

Should I also add:
- Require comment for ratings below 3?
- Calculate average rating?

User: Yes, add both

Claude: Adding enhanced validation:
[Shows code with both features]

Tests added for edge cases. Should I also update the API to return average rating?
```

## 🛠️ Utility Prompts

### Code Generation

```
Generate boilerplate for [ENTITY] domain:
- Use .claude/examples/README.md as template
- Include entity.go, service.go, repository.go, service_test.go
- Follow our naming conventions
- Add package-level documentation

Just generate the files, I'll fill in the specifics.
```

### Documentation

```
Document [FEATURE] for future Claude instances:
- Add practical examples to .claude/examples/
- Add common operations to .claude/recipes.md
- Update .claude/cheatsheet.md if frequently used
- Keep it concise and actionable
```

### Verification

```
Verify my implementation of [FEATURE]:
- Run through .claude/checklist.md
- Check against .claude/standards.md
- Verify tests pass (make test)
- Check linter (make lint)
- Verify Swagger docs updated

Provide a summary with ✓/✗ for each item.
```

## 💭 Tips for Effective Prompting

### Be Specific About Scope
```
✅ "Add validation for ISBN-13 format to Book domain service"
❌ "Validate books"
```

### Reference Documentation
```
✅ "Following the pattern in .claude/examples/README.md, add..."
❌ "Add a domain..." (without context)
```

### Specify Testing Requirements
```
✅ "Add tests with 100% coverage using table-driven approach"
❌ "Add tests" (unclear expectations)
```

### Clarify Layer Placement
```
✅ "Add business rule to domain service, not use case"
❌ "Add validation" (unclear where)
```

### Request Explanations
```
✅ "Explain why this follows Clean Architecture"
❌ Just accepting code without understanding
```

## 📚 Quick Reference

| Task | Prompt Template |
|------|----------------|
| New Domain | "Add [entity] domain with [requirements]. Follow Clean Architecture." |
| New Endpoint | "Add [method] [path] endpoint following existing handler patterns." |
| Fix Bug | "Fix [issue]. Check .claude/troubleshooting.md first." |
| Add Tests | "Add table-driven tests for [component] with 100% coverage." |
| Refactor | "Refactor [code] to follow .claude/standards.md." |
| Review | "Review [changes] against .claude/checklist.md." |
| Document | "Add [feature] examples to .claude/examples/." |

---

**Remember: Clear prompts → Better code. Take 30 seconds to write a good prompt.** 🎯
