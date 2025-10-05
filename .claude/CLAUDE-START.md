# CLAUDE CODE - START HERE

> **Boot sequence for new Claude Code instances. Read this FIRST before any task.**

## 🚨 CRITICAL: Read This in Your First 60 Seconds

You are working on a **Library Management System** built with **Clean Architecture** in Go.

**3 Rules You Must Never Break:**

1. **Dependencies point INWARD only:** Domain ← Use Case ← Adapters ← Infrastructure
2. **Business logic lives in domain services** (NOT in use cases or handlers)
3. **Use case packages have "ops" suffix** (bookops, authops, loanops) to avoid conflicts

**Violating these rules = Architectural violation = Rejected code**

---

## ⚡ 60-Second Orientation

### What This Project Is

A REST API for library management with:
- **Entities:** Book, Member, Author, Subscription, (Loan - future)
- **Core Features:** Book borrowing, subscriptions, late fees, authentication
- **Tech Stack:** Go 1.25, PostgreSQL, Redis (future), JWT auth
- **Architecture:** Clean Architecture (Hexagonal/Onion)

### Current State

```
✅ Implemented: Books, Members, Authors, Subscriptions, Authentication
🚧 In Progress: (check git status)
❌ Not Yet: Loans, Reservations, Notifications
```

---

## 📖 Your First 3 Minutes (Read These Files)

**Don't read randomly. Read in this exact order:**

### Minute 1: Project Overview
1. Read **[README.md](./README.md)** (30 sec) - Quick commands
2. Read **[glossary.md](./glossary.md)** (90 sec) - Business terms (CRITICAL!)

### Minute 2: Find What You Need
3. Read **[context-guide.md](./context-guide.md)** (60 sec) - What to read for YOUR task

### Minute 3: Locate Code
4. Read **[codebase-map.md](./codebase-map.md)** (60 sec) - Where everything is

**After 3 minutes, you should know:**
- ✅ What a "loan" vs "subscription" is
- ✅ Which files to read for your specific task
- ✅ Where to find similar code to reference

---

## 🎯 Task-Specific Quick Start

**Read the section that matches your task, then follow the linked guide:**

### Adding a New Feature / Domain Entity

**Read (5 minutes):**
1. [adrs/001-clean-architecture.md](./adrs/001-clean-architecture.md) - Why we structure this way
2. [adrs/002-domain-services.md](./adrs/002-domain-services.md) - Where logic goes
3. [examples/README.md](./examples/README.md) - Complete code example

**Then:** Follow the 7-layer checklist in examples/

---

### Fixing a Bug

**Read (3 minutes):**
1. [troubleshooting.md](./troubleshooting.md) - Search your error
2. [gotchas.md](./gotchas.md) - Common mistakes
3. [flows.md](./flows.md) - Understand the flow

**Then:** Use grep to find the bug location, read that file + tests

---

### Writing Tests

**Read (2 minutes):**
1. [testing.md](./testing.md) - Testing strategy

**Then:** Look at existing `*_test.go` files in the same package

**Remember:**
- Domain tests: 100% coverage, NO mocks
- Use case tests: 80%+ coverage, mock repositories

---

### Adding an API Endpoint

**Read (3 minutes):**
1. [api.md](./api.md) - API standards
2. [examples/README.md](./examples/README.md#adding-a-new-api-endpoint) - Complete example

**Steps:** Use case → Handler → DTO → Route → Swagger → Test

---

### Database Work (Migrations/Queries)

**Read (2 minutes):**
1. [recipes.md](./recipes.md#database-migrations) - Migration commands
2. [adrs/006-postgresql.md](./adrs/006-postgresql.md) - PostgreSQL best practices

**Then:** `make migrate-create name=your_migration`

---

### Refactoring Existing Code

**Read (5 minutes):**
1. [refactoring.md](./refactoring.md) - Safe refactoring steps
2. Relevant ADRs in [adrs/](./adrs/) directory
3. [gotchas.md](./gotchas.md) - Avoid anti-patterns

**Rule:** Tests must pass before AND after refactoring

---

### Code Review / Pre-Commit

**Read (2 minutes):**
1. [checklist.md](./checklist.md) - Review checklist

**Or run:** `.claude/scripts/review.sh` (automated checks)

---

## 🧠 Mental Model (Memorize This)

### Layer Responsibilities

```
┌─────────────────────────────────────────────┐
│  Infrastructure (HTTP server, DB, JWT)      │  ← Technical plumbing
│  ┌───────────────────────────────────────┐  │
│  │  Adapters (Handlers, Repos, DTOs)    │  │  ← External interfaces
│  │  ┌─────────────────────────────────┐  │  │
│  │  │  Use Cases (Orchestration)      │  │  │  ← Application logic
│  │  │  ┌───────────────────────────┐  │  │  │
│  │  │  │  Domain (Business Logic)  │  │  │  │  ← Core business rules
│  │  │  └───────────────────────────┘  │  │  │
│  │  └─────────────────────────────────┘  │  │
│  └───────────────────────────────────────┘  │
└─────────────────────────────────────────────┘

Dependencies point INWARD ONLY ───────────────→
```

### "Where Does This Code Go?" Decision Tree

```
Is it business logic? (validation, calculation, rules)
    YES → Domain Service (internal/domain/{entity}/service.go)
    NO → ↓

Is it orchestration? (get data → validate → persist → cache)
    YES → Use Case (internal/usecase/{entity}ops/{operation}.go)
    NO → ↓

Is it HTTP-related? (request parsing, response formatting)
    YES → Handler (internal/adapters/http/handlers/{entity}.go)
    NO → ↓

Is it database-related? (SQL queries, transactions)
    YES → Repository (internal/adapters/repository/postgres/{entity}.go)
    NO → ↓

Is it infrastructure? (DB connection, JWT, config)
    YES → Infrastructure (internal/infrastructure/{concern}/)
```

### Package Naming Rule

```
Entity: Book
Domain package: book
Use case package: bookops  ← "ops" suffix to avoid conflict!

import (
    "internal/domain/book"        // Entity, Service, Repository interface
    "internal/usecase/bookops"    // Use cases
)

// No import aliases needed - names are distinct
book.Entity{}
bookops.CreateBookUseCase{}
```

---

## 🚫 Common Mistakes (Avoid These!)

### ❌ WRONG: Business Logic in Use Case

```go
// internal/usecase/bookops/create_book.go
func (uc *CreateBookUseCase) Execute(req Request) error {
    // ❌ WRONG - ISBN validation is business logic!
    if len(req.ISBN) != 13 {
        return errors.New("invalid ISBN")
    }
}
```

### ✅ CORRECT: Business Logic in Domain Service

```go
// internal/domain/book/service.go
func (s *Service) ValidateISBN(isbn string) error {
    // ✅ CORRECT - Business logic in domain
    if len(isbn) != 13 {
        return errors.New("invalid ISBN")
    }
    // ... checksum validation
}

// internal/usecase/bookops/create_book.go
func (uc *CreateBookUseCase) Execute(req Request) error {
    // ✅ CORRECT - Use case calls domain service
    if err := uc.bookService.ValidateISBN(req.ISBN); err != nil {
        return err
    }
}
```

### ❌ WRONG: Package Naming Conflict

```go
// internal/usecase/book/create_book.go
package book  // ❌ CONFLICTS with domain/book!
```

### ✅ CORRECT: "ops" Suffix

```go
// internal/usecase/bookops/create_book.go
package bookops  // ✅ Different from domain/book
```

### ❌ WRONG: Domain Importing Use Case

```go
// internal/domain/book/service.go
import "internal/usecase/bookops"  // ❌ FORBIDDEN!
```

### ✅ CORRECT: Use Case Importing Domain

```go
// internal/usecase/bookops/create_book.go
import "internal/domain/book"  // ✅ Allowed (points inward)
```

**See [gotchas.md](./gotchas.md) for 50+ more examples**

---

## 🎓 Key Concepts You Must Understand

### Domain Entity vs. Use Case vs. Handler

| Layer | Example | Responsibility | Dependencies |
|-------|---------|----------------|--------------|
| **Domain Entity** | `book.Entity` | Data structure | None |
| **Domain Service** | `book.Service.ValidateISBN()` | Business rules | None |
| **Use Case** | `bookops.CreateBookUseCase` | Orchestration | Domain + Repos |
| **Handler** | `BookHandler.CreateBook()` | HTTP I/O | Use Cases |

### Repository Pattern

```go
// Interface defined in DOMAIN (not adapter!)
// internal/domain/book/repository.go
package book

type Repository interface {
    Create(ctx context.Context, book Entity) error
}

// Implementation in ADAPTER
// internal/adapters/repository/postgres/book.go
package postgres

func (r *BookRepository) Create(ctx context.Context, book book.Entity) error {
    // PostgreSQL implementation
}
```

**Why?** Dependency Inversion Principle. Domain defines contract, adapters fulfill it.

### JWT Authentication Flow

```
1. POST /auth/login {email, password}
2. ← {access_token, refresh_token}
3. Request: Authorization: Bearer <access_token>
4. Middleware validates token → extracts claims
5. Handler gets memberID from claims
```

**Protected endpoints need:**
```go
// @Security BearerAuth
```

**Get user in handler:**
```go
claims := auth.GetClaimsFromContext(r.Context())
memberID := claims.MemberID
```

---

## 📋 Pre-Flight Checklist (Before Coding)

Before you start coding, verify:

- [ ] I understand the business domain terms (read [glossary.md](./glossary.md))
- [ ] I know which layer my code belongs to (domain/usecase/adapter/infrastructure)
- [ ] I found similar existing code to reference ([codebase-map.md](./codebase-map.md))
- [ ] I read the relevant ADR if touching architecture ([adrs/](./adrs/))
- [ ] I know how to test my changes ([testing.md](./testing.md))

**If any checkbox is unchecked, STOP and read the linked file first.**

---

## 🛠️ Essential Commands

```bash
# Development
make dev                  # Start everything (DB + API)
make test                 # Run all tests
make ci                   # Full CI checks (before commit)

# Database
make migrate-up           # Apply migrations
make migrate-down         # Rollback migration
make migrate-create name=add_loans  # Create new migration

# Code Quality
make fmt                  # Format code
make lint                 # Run linters
make gen-docs             # Regenerate Swagger docs

# Emergency
lsof -ti:8080 | xargs kill -9     # Kill port 8080
make down && make up              # Restart Docker
go clean -testcache               # Clear test cache
```

---

## 🔗 Quick Reference Links

| I need to... | Read this |
|--------------|-----------|
| Understand business terms | [glossary.md](./glossary.md) |
| Find where code lives | [codebase-map.md](./codebase-map.md) |
| Know what to read for my task | [context-guide.md](./context-guide.md) |
| See code examples | [examples/README.md](./examples/README.md) |
| Understand WHY decisions were made | [adrs/](./adrs/) |
| Fix an error | [troubleshooting.md](./troubleshooting.md) |
| Avoid mistakes | [gotchas.md](./gotchas.md) |
| Quick command | [recipes.md](./recipes.md) |
| Before committing | [checklist.md](./checklist.md) |

---

## 🎯 Success Criteria

You'll know you're ready to code when you can answer:

1. **"What's the difference between a Book and a Loan?"**
   → Answer in [glossary.md](./glossary.md)

2. **"Where does business logic go?"**
   → Domain services (NOT use cases!)

3. **"Where do I find the Book entity?"**
   → `internal/domain/book/entity.go`

4. **"What's the 'ops' suffix for?"**
   → Avoid package naming conflicts (bookops vs book)

5. **"Can domain import from use case layer?"**
   → NO! Never! Dependencies point inward only.

**If you can't answer these, read for 3 more minutes.**

---

## 💡 Pro Tips

1. **Read tests first** - They show how code is actually used
2. **Use grep liberally** - `grep -r "CreateBook" internal/`
3. **Follow existing patterns** - Look at `book/` implementation, copy structure
4. **Don't read everything** - Use [context-guide.md](./context-guide.md) to filter
5. **Check ADRs before changing architecture** - Understand WHY before changing

---

## ⚠️ If You're Stuck

1. **Error?** → Search [troubleshooting.md](./troubleshooting.md)
2. **Mistake?** → Check [gotchas.md](./gotchas.md)
3. **Question?** → Search [faq.md](./faq.md)
4. **Confused about decision?** → Read relevant [ADR](./adrs/)

---

## 🚀 Now You're Ready

**You've completed orientation. You should now:**
- ✅ Understand the domain (books, members, loans, subscriptions)
- ✅ Know the architecture (Clean Architecture, layers, dependencies)
- ✅ Know where to find code (codebase map)
- ✅ Know what to read next (context guide)

**Go forth and code!** 🎉

**Remember:** When in doubt, read an ADR. Every major decision has documentation explaining WHY.

---

**Last Updated:** 2025-01-19
**Next Review:** When onboarding process changes
