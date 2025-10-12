# Architecture Decision Records (ADRs)

This directory contains Architecture Decision Records for the Library Management System.

## What are ADRs?

ADRs document significant architectural decisions made in the project, including:
- **Context:** Why was this decision needed?
- **Decision:** What did we decide?
- **Consequences:** What are the trade-offs?
- **Alternatives:** What else did we consider?

## Index

| ADR | Title | Status | Date |
|-----|-------|--------|------|
| [001](./001-use-case-ops-suffix.md) | Use Case Packages Use "ops" Suffix | Accepted | 2025-10-09 |
| [002](./002-clean-architecture-boundaries.md) | Clean Architecture Layer Boundaries | Accepted | 2025-10-09 |
| [003](./003-domain-services-vs-infrastructure.md) | Domain Services vs Infrastructure Services | Accepted | 2025-10-09 |
| [004](./004-handler-subdirectories.md) | Handler Organization by Domain | Accepted | 2025-10-09 |
| [005](./005-payment-gateway-interface.md) | Payment Gateway Interface Abstraction | Accepted | 2025-10-09 |

## Quick Navigation

### For New Team Members

Start here to understand the architecture:
1. **[ADR 002](./002-clean-architecture-boundaries.md)** - Overall architecture pattern
2. **[ADR 001](./001-use-case-ops-suffix.md)** - Why packages are named "bookops" not "book"
3. **[ADR 003](./003-domain-services-vs-infrastructure.md)** - Where to create services

### For Adding Features

Read these before implementing:
1. **[ADR 002](./002-clean-architecture-boundaries.md)** - Layer dependencies (critical!)
2. **[ADR 003](./003-domain-services-vs-infrastructure.md)** - Where to put your service
3. **[ADR 004](./004-handler-subdirectories.md)** - How to organize HTTP handlers

### For Understanding Code

Use these to understand existing patterns:
1. **[ADR 001](./001-use-case-ops-suffix.md)** - Why use cases have "ops" suffix
2. **[ADR 005](./005-payment-gateway-interface.md)** - How payment gateway works

## ADR Summary

### ADR 001: Use Case "ops" Suffix

**Problem:** Importing both `domain/book` and `usecase/book` creates naming conflicts

**Solution:** Use case packages use "ops" suffix: `usecase/bookops`

**Impact:** Clean imports, no aliases needed

```go
import (
    "library-service/internal/domain/book"
    "library-service/internal/usecase/bookops"  // No conflict!
)
```

---

### ADR 002: Clean Architecture Boundaries

**Problem:** Tight coupling between business logic, database, and HTTP

**Solution:** Four-layer architecture with strict dependency rules

**Impact:** Testable, flexible, maintainable

```
Domain → Use Case → Adapters → Infrastructure
(pure)   (orchestrate) (implement) (configure)
```

**Key Rule:** Dependencies point inward only. Domain has ZERO external dependencies.

---

### ADR 003: Domain vs Infrastructure Services

**Problem:** Confusion about where to create services

**Solution:** Two types with different creation locations

**Impact:** Clear separation of concerns

| Type | Created In | Dependencies |
|------|-----------|--------------|
| Domain Service | `container.go` | None (pure) |
| Infrastructure Service | `app.go` | External libraries |

**Examples:**
- Domain: ISBN validation, pricing calculation
- Infrastructure: JWT generation, password hashing

---

### ADR 004: Handler Subdirectories

**Problem:** 22 handler files in one flat directory

**Solution:** Organize by domain into subdirectories

**Impact:** Better navigation, reduced conflicts

```
handlers/
├── auth/      # Authentication endpoints
├── book/      # Book CRUD endpoints
├── payment/   # Payment endpoints
└── ...
```

---

### ADR 005: Payment Gateway Interface

**Problem:** Use cases directly coupled to epayment.kz gateway

**Solution:** Define gateway interface in use case layer

**Impact:** Easy to test, easy to switch gateways

```go
type PaymentGateway interface {  // In use case layer
    GetAuthToken(ctx context.Context) (string, error)
    CheckPaymentStatus(ctx context.Context, invoiceID string) (interface{}, error)
}

// epayment.Gateway implements this interface
```

## ADR Status

- **Accepted:** Decision is final and implemented
- **Proposed:** Under discussion
- **Deprecated:** No longer recommended
- **Superseded:** Replaced by another ADR

All current ADRs are **Accepted** and reflect implemented patterns.

## Creating New ADRs

Use this template:

```markdown
# ADR NNN: Title

**Status:** Proposed | Accepted | Deprecated

**Date:** YYYY-MM-DD

## Context

What problem are we solving? Why is this decision needed?

## Decision

What did we decide to do?

## Consequences

### Positive
What are the benefits?

### Negative
What are the downsides?

## Alternatives Considered

What else did we think about?

## Related Decisions

Links to other ADRs

## References

Implementation locations, commits, etc.
```

## Related Documentation

- **Architecture Guide:** `.claude/architecture.md`
- **Development Guide:** `.claude/development.md`
- **Quick Start:** `.claude/AI-QUICKSTART.md`
- **Container Documentation:** `internal/usecase/container.go` (comprehensive DI guide)

## Questions?

- Check `.claude/faq.md` for common questions
- See `.claude/troubleshooting.md` for issues
- Read `CLAUDE.md` for overall guidance
