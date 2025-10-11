# .claude - Go Clean Architecture Project Documentation

> **⚡ Fast-track guide for Claude Code instances working with Go Clean Architecture projects**

---

## 🚀 Start Here (15 Minutes to Productivity)

### For New Claude Code Instances

**Step 1: Universal Go Patterns (10 min)**
→ Read **[GO-ONBOARDING.md](./GO-ONBOARDING.md)** first
- Works for ANY Go Clean Architecture project
- Essential patterns, conventions, and pitfalls
- Quick command reference

**Step 2: This Project Specifics (5 min)**
→ Continue below ↓

**Total time: 15 minutes → Fully productive** ✅

---

## 📋 Project: Library Management System

**What:** REST API for library operations (books, members, reservations, payments)
**Stack:** Go 1.25, PostgreSQL, Redis, Chi router, JWT auth
**Architecture:** Clean Architecture (Hexagonal/Onion pattern)

**3 Critical Architectural Rules (Never Break!):**

1. **Dependencies flow inward only:** `Domain ← Use Case ← Adapters ← Infrastructure`
2. **Business logic in domain services** (NOT in use cases or handlers)
3. **Use case packages use "ops" suffix** (`bookops`, `authops`, not `book`, `auth`)

---

## 📖 Documentation Structure

### 🎯 Essential Reading (Core Guides)

Start with these for daily development:

| File | Purpose | Read When |
|------|---------|-----------|
| **[architecture.md](./architecture.md)** | Clean Architecture principles & structure | Before making changes |
| **[development-guide.md](./development-guide.md)** | Setup, environment, daily commands | Setting up / daily work |
| **[common-tasks.md](./common-tasks.md)** | Step-by-step implementation guides | Adding features |
| **[coding-standards.md](./coding-standards.md)** | Go conventions & best practices | Writing code |
| **[testing.md](./testing.md)** | Testing strategy & patterns | Writing tests |
| **[security.md](./security.md)** | Security best practices | Auth, validation, sensitive data |

### 🏛️ Architecture Decision Records (ADRs)

Understand the "why" behind key decisions:

| ADR | Decision | Status |
|-----|----------|--------|
| [001](./adrs/001-clean-architecture.md) | Clean Architecture pattern | ✅ Active |
| [002](./adrs/002-domain-services.md) | Domain service pattern | ✅ Active |
| [003](./adrs/003-two-step-di.md) | Two-step dependency injection | ✅ Active |
| [004](./adrs/004-ops-suffix-convention.md) | "ops" suffix for use case packages | ✅ Active |
| [005](./adrs/005-repository-interfaces.md) | Repository pattern | ✅ Active |
| [006](./adrs/006-postgresql.md) | PostgreSQL as database | ✅ Active |
| [007](./adrs/007-jwt-authentication.md) | JWT authentication | ✅ Active |

Read these to understand architectural constraints.

### 🔧 Advanced Reference

For complex scenarios:

| File | Purpose | Read When |
|------|---------|-----------|
| [reference/debugging-guide.md](./reference/debugging-guide.md) | Advanced debugging techniques | Complex bugs |
| [reference/performance.md](./reference/performance.md) | Profiling & optimization | Performance issues |
| [reference/refactoring.md](./reference/refactoring.md) | Safe refactoring patterns | Large refactorings |
| [reference/checklist.md](./reference/checklist.md) | Pre-commit review checklist | Before PR |

### 🛠️ Scripts

| Script | Purpose | Usage |
|--------|---------|-------|
| [scripts/review.sh](./scripts/review.sh) | Pre-commit checks | `./scripts/review.sh` |

---

## 🎯 Quick Decision Tree

**"I'm new to this codebase"**
→ Read: GO-ONBOARDING.md → architecture.md → development-guide.md

**"I need to add a feature"**
→ Read: architecture.md (structure) → common-tasks.md (step-by-step)

**"I need to fix a bug"**
→ Read: reference/debugging-guide.md

**"I'm writing tests"**
→ Read: testing.md → reference/checklist.md

**"I need to optimize performance"**
→ Read: reference/performance.md

**"I need to refactor code"**
→ Read: reference/refactoring.md + relevant ADRs

---

## 🏗️ Project Structure Quick Reference

```
internal/
├── domain/              # 🎯 Business entities & rules (ZERO external dependencies)
│   ├── book/           # Entity, Service, Repository interface
│   ├── member/
│   ├── author/
│   ├── reservation/
│   └── payment/
│
├── usecase/            # 🔄 Application orchestration
│   ├── bookops/        # Note: "ops" suffix to avoid conflicts
│   ├── authops/
│   ├── reservationops/
│   └── paymentops/
│
├── adapters/           # 🔌 External interfaces
│   ├── http/           # HTTP handlers (Chi router)
│   ├── repository/     # DB implementations (PostgreSQL)
│   └── cache/          # Cache implementations (Redis)
│
└── infrastructure/     # ⚙️ Technical concerns
    ├── auth/           # JWT, password hashing
    ├── store/          # DB connections
    └── server/         # HTTP server

cmd/
├── api/               # API server entry point
├── worker/            # Background worker
└── migrate/           # Database migrations

pkg/                   # Shared packages
├── errors/            # Standard error types
├── httputil/          # HTTP helpers
└── logutil/           # Logging utilities
```

**Dependency Rule:** Inner layers NEVER import outer layers
- ✅ `usecase` imports `domain`
- ✅ `adapters` imports `usecase` and `domain`
- ❌ `domain` imports `adapters` (NEVER!)

---

## ⚡ Quick Start Commands

```bash
# Setup (first time)
make init && make up && make migrate-up

# Daily development
make dev              # Start everything (Docker + API)

# Testing
make test             # All tests with coverage
make test-unit        # Unit tests only (fast)

# Code quality
make ci               # Full CI: fmt + vet + lint + test + build

# Before commit
./scripts/review.sh   # Pre-commit checks
```

See [development-guide.md](./development-guide.md) for complete command reference.

---

## 📊 Current Implementation Status

**✅ Implemented Features:**
- Books management (CRUD)
- Authors management
- Members & authentication (JWT)
- Subscriptions
- Reservations (with domain service)
- Payments (integration with gateway)

**📈 Test Coverage:**
- Overall use case coverage: **45.6%**
- Domain services: **100%** (critical business logic)
- Individual packages: bookops (91%), authops (91%), reservationops (93%)

**🔍 Check Current Status:**
```bash
git status              # Current changes
git log --oneline -5   # Recent commits
go test ./... -cover   # Test coverage
```

---

## 🚨 Common Pitfalls (Avoid These!)

### 1. Import Cycle Violations
```go
// ❌ NEVER: Domain imports adapters
package book
import "myproject/internal/adapters/repository"  // WRONG!

// ✅ CORRECT: Domain defines interface, adapters import domain
package book
type Repository interface {
    Create(ctx context.Context, book Book) error
}
```

### 2. Business Logic in Wrong Layer
```go
// ❌ WRONG: Business validation in HTTP handler
func (h *Handler) CreateBook(w http.ResponseWriter, r *http.Request) {
    if book.Price < 0 {  // Business rule in handler!
        // ...
    }
}

// ✅ CORRECT: Business logic in domain service
func (s *Service) ValidateBook(book Book) error {
    if book.Price < 0 {
        return ErrInvalidPrice
    }
    return nil
}
```

### 3. Missing Context Parameter
```go
// ❌ BAD: No context
func (r *Repo) GetUser(id string) (User, error)

// ✅ GOOD: Context as first parameter
func (r *Repo) GetUser(ctx context.Context, id string) (User, error)
```

### 4. Package Naming Conflicts
```go
// ❌ PROBLEMATIC: Domain and use case both called "book"
internal/domain/book/       # package book
internal/usecase/book/      # package book (conflict!)

// ✅ CORRECT: Use "ops" suffix for use cases
internal/domain/book/       # package book
internal/usecase/bookops/   # package bookops (no conflict)
```

---

## 📚 Learning Resources

**For Claude Code instances:**
1. Start with GO-ONBOARDING.md (universal Go patterns)
2. Skim architecture.md (this project's structure)
3. Reference common-tasks.md as needed (how-to guides)
4. Check ADRs for architectural context

**For humans:**
1. Read GO-ONBOARDING.md (Go architecture overview)
2. Read development-guide.md (setup environment)
3. Read architecture.md (understand structure)
4. Use common-tasks.md during development

---

## 🔍 Finding Documentation

**Search all docs:**
```bash
grep -r "search term" .claude/
```

**Search excluding reference:**
```bash
grep -r "search term" .claude/ --exclude-dir=reference --exclude-dir=adrs
```

**Search ADRs only:**
```bash
grep -r "search term" .claude/adrs/
```

---

## ✅ You're Ready When You Can Answer:

1. ✓ **Where do I add business logic?** → Domain services (`internal/domain/*/service.go`)
2. ✓ **Where do I add API endpoints?** → HTTP handlers (`internal/adapters/http/handlers/`)
3. ✓ **Where do I orchestrate operations?** → Use cases (`internal/usecase/*ops/`)
4. ✓ **How do I test my code?** → Table-driven tests, mock repositories
5. ✓ **What's the dependency rule?** → Inner layers never import outer layers
6. ✓ **Why "ops" suffix?** → Avoids naming conflicts (ADR-004)

If you can answer these, you're ready to be productive! 🚀

---

**Next:** Check [CLAUDE.md](../CLAUDE.md) in project root for any custom instructions.
