# .claude - Library Management System Documentation

> **⚡ Quick navigation for Claude Code instances**

---

## 🚀 Start Here

**For New Claude Code Instances:**

1. **Read CLAUDE.md** (root) - 2 minutes
   - Project overview and quick commands
   - Entry point with links to everything

2. **Read `.claude-context/SESSION_MEMORY.md`** - 3 minutes
   - Essential architecture context
   - Current implementation state

3. **Read `.claude-context/CURRENT_PATTERNS.md`** - 3 minutes
   - Code patterns and conventions
   - Import aliases and naming

**Total time: 8 minutes → Fully productive** ✅

---

## 📁 Documentation Structure

### 📖 [guides/](./guides/) - Core Guides (7 files)

Daily development documentation:

- **[architecture.md](./guides/architecture.md)** - Clean Architecture principles & bounded contexts
- **[development.md](./guides/development.md)** - Setup, environment, daily commands
- **[common-tasks.md](./guides/common-tasks.md)** - Step-by-step implementation guides
- **[coding-standards.md](./guides/coding-standards.md)** - Go conventions & best practices
- **[testing.md](./guides/testing.md)** - Testing strategy & patterns
- **[security.md](./guides/security.md)** - Security best practices & JWT
- **[cache-warming.md](./guides/cache-warming.md)** - Cache warming implementation

### 📋 [adr/](./adr/) - Architecture Decision Records (13 files)

Why decisions were made:

1. **001-use-case-ops-suffix.md** - Use case package naming convention
2. **002-clean-architecture-boundaries.md** - Layer dependency rules
3. **003-domain-services-vs-infrastructure.md** - Service placement
4. **004-handler-subdirectories.md** - HTTP handler organization
5. **005-payment-gateway-interface.md** - Payment gateway abstraction
6. **006-postgresql.md** - Database choice
7. **007-jwt-authentication.md** - Auth strategy
8. **008-generic-repository-helpers.md** - Repository patterns
9. **009-payment-gateway-modularization.md** - Payment architecture
10. **010-domain-service-payment-status.md** - Payment domain logic
11. **011-base-repository-pattern.md** - Generic repository implementation
12. **012-bounded-context-organization.md** - Vertical slice architecture
13. **013-dto-colocation-and-token-optimization.md** - DTO organization

### 📚 [reference/](./reference/) - Reference Materials (4 files)

Quick lookup information:

- **[common-mistakes.md](./reference/common-mistakes.md)** - Gotchas and anti-patterns
- **[error-handling.md](./reference/error-handling.md)** - Error handling patterns
- **[migration-guide.md](./reference/migration-guide.md)** - Repository migration guide
- **[go-onboarding.md](./reference/go-onboarding.md)** - Universal Go patterns

### 📦 [archive/](./archive/) - Historical Documents

Completed refactoring documentation (for reference only).

---

## 🎯 Quick Navigation by Task

### I want to...

**Understand the project**
→ Read [guides/architecture.md](./guides/architecture.md)

**Set up my environment**
→ Read [guides/development.md](./guides/development.md)

**Add a new feature**
→ Read [guides/common-tasks.md](./guides/common-tasks.md)

**Write tests**
→ Read [guides/testing.md](./guides/testing.md)

**Understand why something was done this way**
→ Browse [adr/](./adr/)

**Avoid common pitfalls**
→ Read [reference/common-mistakes.md](./reference/common-mistakes.md)

**Learn Go patterns**
→ Read [reference/go-onboarding.md](./reference/go-onboarding.md)

---

## 📝 Related Documentation

- **[/examples/](../../examples/)** - Code pattern examples (handler, usecase, repository, testing)
- **[/docs/payments/](../../docs/payments/)** - Payment integration documentation
- **[/.claude-context/](../.claude-context/)** - Session context files (essential!)
- **[/CLAUDE.md](../../CLAUDE.md)** - Main entry point (start here!)

---

## 🏗️ Project Structure

```
.claude/
├── README.md                  # You are here
│
├── guides/                    # How to work with the project
│   ├── architecture.md
│   ├── development.md
│   ├── common-tasks.md
│   ├── coding-standards.md
│   ├── testing.md
│   ├── security.md
│   └── cache-warming.md
│
├── adr/                       # Why decisions were made
│   ├── 001-*.md through 013-*.md
│   └── README.md
│
├── reference/                 # Quick lookup info
│   ├── common-mistakes.md
│   ├── error-handling.md
│   ├── migration-guide.md
│   └── go-onboarding.md
│
└── archive/                   # Historical documents
    └── [refactoring history]
```

---

**Last Updated:** October 11, 2025
**Documentation Files:** 25 active + archive
