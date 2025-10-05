# Context Loading Guide for Claude Code

> **What to read FIRST for maximum productivity in minimum time**

## Purpose

This guide tells you EXACTLY which files to read for specific tasks. Don't read everything - read what you need, when you need it.

**Goal:** Get from "I just opened this project" to "I'm making productive changes" in under 5 minutes.

---

## 🚀 Universal "Cold Start" (Read These First)

**For ANY task, read these 3 files first (2-3 minutes):**

1. **[README.md](./README.md)** - 30-second overview, quick commands
2. **[cheatsheet.md](./cheatsheet.md)** - Where things are, common patterns
3. **[glossary.md](./glossary.md)** - Business domain terms (what is a loan, subscription, etc.)

Then use the task-specific sections below.

---

## 📋 Task-Specific Context Loading

### Adding a New Domain Entity (e.g., "Add Loan feature")

**Read in this order (10 minutes):**
1. ✅ **[adrs/001-clean-architecture.md](./adrs/001-clean-architecture.md)** - Why we separate layers (2 min)
2. ✅ **[adrs/002-domain-services.md](./adrs/002-domain-services.md)** - Where business logic goes (2 min)
3. ✅ **[adrs/004-ops-suffix-convention.md](./adrs/004-ops-suffix-convention.md)** - Package naming (1 min)
4. ✅ **[examples/README.md](./examples/README.md)** - Complete working example (5 min)
5. ✅ **[codebase-map.md](./codebase-map.md)** - Find similar code to reference

**Then:**
- Create domain layer (entity, service, repository interface, tests)
- Create use case layer (with "ops" suffix)
- Create adapter layer (handler, repository impl, DTOs)
- Wire in `container.go`
- Add routes in `router.go`

**Don't read:** Troubleshooting, gotchas (read these if you hit issues)

---

### Fixing a Bug

**Read in this order (5 minutes):**
1. ✅ **[troubleshooting.md](./troubleshooting.md)** - Search for your error message
2. ✅ **[gotchas.md](./gotchas.md)** - Common mistakes that cause bugs
3. ✅ **[flows.md](./flows.md)** - Understand request flow for your bug's area
4. ✅ **[codebase-map.md](./codebase-map.md)** - Find which files to investigate

**Then:**
- Use `grep -r "error message"` to find the source
- Read the file containing the bug
- Check tests for that component
- Understand the flow (handler → use case → domain → repository)

**Don't read:** Examples, ADRs (unless bug involves architecture)

---

### Writing Tests

**Read in this order (5 minutes):**
1. ✅ **[testing.md](./testing.md)** - Testing strategy and patterns (3 min)
2. ✅ **[examples/README.md](./examples/README.md#testing-examples)** - Test examples (2 min)
3. ✅ **Look at existing test files** in the same package

**Testing strategy quick reference:**
```
Domain tests:     100% coverage, NO mocks (pure functions)
Use case tests:   80%+ coverage, mock repositories
Repository tests: Integration tests with real database
Handler tests:    Mock use cases, test HTTP layer
```

**Then:**
- Follow table-driven test pattern
- Use mocks for repositories in use case tests
- Use real services (no mocks) in use case tests

**Don't read:** Architecture docs (unless new to project)

---

### Adding an API Endpoint

**Read in this order (8 minutes):**
1. ✅ **[api.md](./api.md)** - API design standards (2 min)
2. ✅ **[adrs/007-jwt-authentication.md](./adrs/007-jwt-authentication.md)** - If endpoint needs auth (2 min)
3. ✅ **[examples/README.md](./examples/README.md#adding-a-new-api-endpoint)** - Complete example (4 min)

**Steps:**
1. Create use case (if doesn't exist)
2. Create handler in `internal/adapters/http/handlers/{entity}.go`
3. Create DTOs in `internal/adapters/http/dto/`
4. Add route in `internal/adapters/http/routes/router.go`
5. Add Swagger annotations
6. Run `make gen-docs`

**Swagger annotation template:**
```go
// @Summary      Brief description
// @Tags         entity-name
// @Security     BearerAuth  (if protected)
// @Accept       json
// @Produce      json
// @Param        request body dto.CreateBookRequest true "Request body"
// @Success      201 {object} dto.BookResponse
// @Failure      400 {object} dto.ErrorResponse
// @Router       /books [post]
```

**Don't read:** Domain services, testing (add tests after endpoint works)

---

### Refactoring Existing Code

**Read in this order (10 minutes):**
1. ✅ **[refactoring.md](./refactoring.md)** - Safe refactoring steps (3 min)
2. ✅ **[adrs/](./adrs/)** - Read ADRs related to what you're refactoring (5 min)
3. ✅ **[gotchas.md](./gotchas.md)** - Avoid introducing anti-patterns (2 min)

**Safe refactoring process:**
1. Read existing tests (understand current behavior)
2. Add missing tests if coverage < 80%
3. Run tests: `go test -v ./path/to/package/`
4. Make refactoring changes
5. Run tests again (must still pass)
6. Run full CI: `make ci`

**Don't read:** Examples (you're changing existing code, not adding new)

---

### Debugging Performance Issues

**Read in this order (5 minutes):**
1. ✅ **[troubleshooting.md](./troubleshooting.md#performance-issues)** - Performance troubleshooting (3 min)
2. ✅ **[recipes.md](./recipes.md#performance-profiling)** - Profiling commands (2 min)

**Quick performance checks:**
```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=. ./internal/domain/book/
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof -bench=. ./internal/domain/book/
go tool pprof mem.prof

# Find N+1 queries
# Check for loops with repository calls inside
grep -A 10 "for.*range" internal/adapters/repository/postgres/*.go
```

**Don't read:** Architecture docs (focus on profiling)

---

### Database Migrations

**Read in this order (5 minutes):**
1. ✅ **[recipes.md](./recipes.md#database-migrations)** - Migration commands (2 min)
2. ✅ **[adrs/006-postgresql.md](./adrs/006-postgresql.md)** - Why PostgreSQL, SQL best practices (3 min)

**Migration workflow:**
```bash
# Create migration
make migrate-create name=add_loans_table

# Edit files:
# migrations/postgres/XXXXXX_add_loans_table.up.sql
# migrations/postgres/XXXXXX_add_loans_table.down.sql

# Test up
make migrate-up

# Test down (rollback)
make migrate-down

# Apply again
make migrate-up
```

**Always include in up migration:**
- Create table with all columns
- Add indexes (especially on foreign keys)
- Add constraints (NOT NULL, UNIQUE, FOREIGN KEY, CHECK)

**Always include in down migration:**
- DROP TABLE (or reverse of up migration)

**Don't read:** Examples (migrations are straightforward)

---

### Setting Up Development Environment

**Read in this order (10 minutes):**
1. ✅ **[setup.md](./setup.md)** - Complete setup instructions (10 min)

**Quick start:**
```bash
# Install dependencies
make install-tools

# Start infrastructure (PostgreSQL, Redis)
make up

# Run migrations
make migrate-up

# Start API
make run

# In another terminal, run tests
make test
```

**If something fails:** Read [troubleshooting.md](./troubleshooting.md)

**Don't read:** Architecture, examples (get environment working first)

---

### Reviewing Code (PR Review)

**Read in this order (5 minutes):**
1. ✅ **[checklist.md](./checklist.md)** - Complete review checklist (5 min)

**Or run automated checks:**
```bash
.claude/scripts/review.sh
```

**Review checklist summary:**
- ✅ Tests added (100% domain, 80%+ use case)
- ✅ Layer boundaries respected (no domain importing use case)
- ✅ Errors wrapped with `%w`
- ✅ No hardcoded secrets
- ✅ Swagger docs updated
- ✅ Database migrations have up AND down
- ✅ `make ci` passes

**Don't read:** Everything else (checklist covers all requirements)

---

### Understanding Architectural Decisions

**Read in this order (15 minutes):**
1. ✅ **[architecture.md](./architecture.md)** - High-level overview (5 min)
2. ✅ **[adrs/001-clean-architecture.md](./adrs/001-clean-architecture.md)** - Why Clean Arch (3 min)
3. ✅ **[adrs/002-domain-services.md](./adrs/002-domain-services.md)** - Where logic goes (2 min)
4. ✅ **[adrs/003-two-step-di.md](./adrs/003-two-step-di.md)** - Dependency injection (2 min)
5. ✅ **[flows.md](./flows.md)** - Visual diagrams (3 min)

**Key takeaways:**
- Domain → Use Case → Adapters → Infrastructure (dependencies point inward)
- Business logic in domain services (not use cases)
- Interfaces defined in domain, implemented in adapters
- Use case packages have "ops" suffix (bookops, authops)

**Don't read:** Examples, recipes (you're learning concepts, not implementing)

---

### Working with Authentication

**Read in this order (5 minutes):**
1. ✅ **[adrs/007-jwt-authentication.md](./adrs/007-jwt-authentication.md)** - Why JWT, how it works (3 min)
2. ✅ **[recipes.md](./recipes.md#authentication)** - How to get tokens (2 min)

**Quick reference:**
```bash
# Get access token
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test123!@#"}' \
  | jq -r '.tokens.access_token')

# Use token
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/books
```

**Protecting endpoints:**
```go
// @Security BearerAuth
// @Router /books [post]
func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
    claims := auth.GetClaimsFromContext(r.Context())
    memberID := claims.MemberID
    // ...
}
```

**Don't read:** Domain docs (authentication is infrastructure concern)

---

## 🎯 Priority Reading Matrix

| Your Task | Priority 1 (MUST read) | Priority 2 (Should read) | Priority 3 (Optional) |
|-----------|------------------------|--------------------------|----------------------|
| **Add new domain** | examples/, adrs/001,002,004 | codebase-map.md, glossary.md | gotchas.md |
| **Fix bug** | troubleshooting.md, gotchas.md | flows.md, codebase-map.md | testing.md |
| **Write tests** | testing.md, examples/ | Existing test files | - |
| **Add API endpoint** | api.md, examples/ | adrs/007, recipes.md | - |
| **Refactoring** | refactoring.md, related ADRs | gotchas.md, testing.md | - |
| **Setup environment** | setup.md | troubleshooting.md | - |
| **Database migration** | recipes.md, adrs/006 | - | - |
| **Code review** | checklist.md | standards.md | - |
| **Understanding architecture** | architecture.md, adrs/001-003 | flows.md | examples/ |
| **Authentication work** | adrs/007 | recipes.md | api.md |

---

## ⚡ Emergency Quick Reference

**"I have 60 seconds before I need to start coding"**

1. Read **[cheatsheet.md](./cheatsheet.md)** (1 page, 60 seconds)
2. Find relevant section in **[examples/README.md](./examples/README.md)**
3. Copy-paste code, adapt to your needs

**"I'm stuck on an error"**

1. Search error message in **[troubleshooting.md](./troubleshooting.md)**
2. If not found, search in **[gotchas.md](./gotchas.md)**
3. If still stuck, search in **[faq.md](./faq.md)**

**"I need a specific command"**

1. Check **[recipes.md](./recipes.md)** (copy-paste solutions)
2. If not there, check **[commands.md](./commands.md)**

**"I don't understand why we do X this way"**

1. Check **[adrs/](./adrs/)** directory
2. Read the relevant ADR (3-5 minutes)

---

## 📊 Reading Time Estimates

| Document | Lines | Reading Time | When to Read |
|----------|-------|--------------|--------------|
| README.md | 116 | 30 sec | Always (first thing) |
| cheatsheet.md | 303 | 2 min | Always (second thing) |
| glossary.md | ~200 | 2 min | Always (third thing) |
| examples/ | 606 | 5-10 min | When adding code |
| adrs/001 | ~300 | 3 min | Understanding architecture |
| flows.md | 680 | 5 min | Understanding request flow |
| troubleshooting.md | 604 | 3 min (scan) | When stuck |
| gotchas.md | 650 | 3 min (scan) | Before committing |
| checklist.md | 431 | 5 min | Before PR |
| testing.md | varies | 5 min | Writing tests |
| api.md | varies | 3 min | Adding endpoints |

**Total to become productive:** ~15 minutes (README + cheatsheet + glossary + examples for your task)

---

## 💡 Pro Tips for Maximum Productivity

1. **Use Grep to Find Examples**
   ```bash
   # Find how we create use cases
   grep -r "NewCreateBookUseCase" .

   # Find similar handlers
   grep -r "http.Handler" internal/adapters/http/handlers/

   # Find tests for pattern
   grep -r "TestService" internal/domain/*/
   ```

2. **Use the Codebase Map**
   - Before writing code, check **[codebase-map.md](./codebase-map.md)**
   - Find similar existing code
   - Copy pattern, adapt to your needs

3. **Read Tests First**
   - Tests show how code is ACTUALLY used
   - Tests reveal edge cases
   - Tests are executable documentation

4. **Don't Read Everything**
   - Only read what you need for your task
   - Use this guide to filter what's relevant
   - Come back for more context when needed

5. **Search Before Reading**
   ```bash
   # Find where book status is used
   grep -r "book.Status" internal/

   # Find all use cases for loans
   find . -path "*/usecase/*loan*" -name "*.go"
   ```

---

## 🔄 Feedback Loop

**As you work, update your knowledge:**

| Did This | Now Read This | Why |
|----------|---------------|-----|
| Got an error | troubleshooting.md | Solutions |
| Made a mistake | gotchas.md | Avoid repeating |
| Before committing | checklist.md | Ensure quality |
| Confused about decision | adrs/ | Understand why |

---

**Remember:** You don't need to read everything. Read what you need, when you need it. This guide ensures you read the RIGHT things at the RIGHT time.
