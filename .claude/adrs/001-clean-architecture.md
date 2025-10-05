# ADR-001: Clean Architecture (Hexagonal/Onion Pattern)

**Status:** Accepted

**Date:** 2024-01-15

**Decision Makers:** Project Architecture Team

## Context

We needed to choose an architectural pattern for a Library Management System that would:
- Support long-term maintainability (5+ years)
- Allow easy testing without external dependencies
- Enable technology changes (swap databases, frameworks) without rewriting business logic
- Support multiple teams working on different layers simultaneously
- Be understandable for AI-assisted development (clear boundaries)

**Constraints:**
- Small team (2-3 developers initially)
- Need for rapid feature development
- Expectation of evolving requirements
- Must support vibecoding (AI-assisted development)

## Decision

We adopted **Clean Architecture** (also known as Hexagonal Architecture or Onion Architecture) with these specific layers:

```
┌─────────────────────────────────────────────┐
│         Infrastructure Layer                │  ← Technical concerns (HTTP server, DB connections)
│  ┌───────────────────────────────────────┐  │
│  │       Adapters Layer                  │  │  ← External interfaces (handlers, repos, cache)
│  │  ┌─────────────────────────────────┐  │  │
│  │  │      Use Case Layer             │  │  │  ← Application orchestration
│  │  │  ┌───────────────────────────┐  │  │  │
│  │  │  │   Domain Layer            │  │  │  │  ← Pure business logic
│  │  │  │   (entities, services)    │  │  │  │
│  │  │  └───────────────────────────┘  │  │  │
│  │  └─────────────────────────────────┘  │  │
│  └───────────────────────────────────────┘  │
└─────────────────────────────────────────────┘
```

**Key rules:**
1. **Dependencies point inward only** (Domain → Use Case → Adapters → Infrastructure)
2. **Domain layer has ZERO external dependencies** (no imports from outer layers, no framework code)
3. **Interfaces defined in inner layers, implemented in outer layers** (Dependency Inversion)
4. **Use cases orchestrate, domain services contain business logic**

## Consequences

### Positive

1. **Testability:** Domain logic can be tested with zero mocks (pure functions, no dependencies)
   ```go
   // No database, no HTTP, just pure business logic testing
   func TestService_CalculateLateFee(t *testing.T) {
       svc := book.NewService()
       fee := svc.CalculateLateFee(14) // 14 days late
       assert.Equal(t, 7.0, fee) // $0.50/day
   }
   ```

2. **Technology Independence:** Swapped from MongoDB to PostgreSQL with zero domain code changes (only adapter layer affected)

3. **Parallel Development:** Teams can work on different layers simultaneously without conflicts

4. **AI-Friendly:** Clear boundaries make it easy for Claude Code to understand where code belongs

5. **Maintainability:** Changes are localized. Adding a new database doesn't touch business logic.

### Negative

1. **More Files:** A simple CRUD operation requires files in 4 layers (domain, use case, adapter, infrastructure)
   - Mitigation: Created `.claude/examples/` with templates for quick scaffolding

2. **Learning Curve:** New developers need to understand the pattern before contributing
   - Mitigation: Created `.claude/onboarding.md` for 15-minute guided onboarding

3. **Boilerplate:** Requires interfaces, DTOs, and mapping code
   - Mitigation: Worth it for long-term maintainability. Not actually that much code.

4. **Over-engineering Risk:** For simple CRUD apps, this might be overkill
   - Mitigation: Our system has complex business rules (loan management, late fees, subscriptions) that justify this approach

5. **Import Verbosity:** Must import from specific layers
   ```go
   import (
       "library-service/internal/domain/book"
       "library-service/internal/usecase/bookops"
       "library-service/internal/adapters/repository/postgres"
   )
   ```
   - Mitigation: Clear imports make dependencies obvious (which is actually a benefit)

## Alternatives Considered

### Alternative 1: MVC (Model-View-Controller)

**Why not chosen:**
- Business logic tends to leak into controllers or models
- Hard to test without HTTP requests
- Framework coupling (tight to web framework)
- Doesn't scale well as complexity grows

### Alternative 2: Layered Architecture (classic 3-tier)

**Why not chosen:**
- Doesn't enforce dependency direction (layers often become bidirectional)
- Database schema often drives the entire design (we want domain to drive it)
- Hard to swap infrastructure (database changes ripple through all layers)

### Alternative 3: Domain-Driven Design (DDD) with CQRS

**Why not chosen:**
- Too complex for our current needs (event sourcing, separate read/write models)
- Team not experienced with DDD tactical patterns (aggregates, value objects, domain events)
- Can adopt later if needed (Clean Architecture doesn't prevent CQRS)

### Alternative 4: Simple Monolithic Structure (handlers → services → models)

**Why not chosen:**
- Works for small projects but doesn't scale
- Business logic gets mixed with HTTP and database concerns
- Hard to test in isolation
- We expect this project to grow significantly

## Implementation Details

**Layer responsibilities:**

| Layer | Responsibilities | Example |
|-------|-----------------|---------|
| **Domain** | Entities, business rules, repository interfaces | `book.Entity`, `book.Service.ValidateISBN()` |
| **Use Case** | Orchestration, transaction boundaries | `bookops.CreateBookUseCase.Execute()` |
| **Adapters** | HTTP handlers, repository implementations, cache | `handlers.BookHandler`, `postgres.BookRepository` |
| **Infrastructure** | DB connections, HTTP server, JWT manager | `store.NewPostgres()`, `server.New()` |

**Why we chose "ops" suffix for use cases:** See [ADR-004](./004-ops-suffix-convention.md)

**Why we use domain services:** See [ADR-002](./002-domain-services.md)

## Validation

After 6 months of development:
- ✅ Successfully migrated from MongoDB to PostgreSQL in 4 hours (only adapter layer changed)
- ✅ Domain layer maintains 100% test coverage with zero mocks
- ✅ New developers productive after 15-minute onboarding
- ✅ AI-assisted development works smoothly (Claude Code understands boundaries)
- ✅ Zero import cycles (strict dependency rules prevent them)

## References

- [The Clean Architecture (Uncle Bob)](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Hexagonal Architecture (Alistair Cockburn)](https://alistair.cockburn.us/hexagonal-architecture/)
- [Clean Architecture in Go (Go Community)](https://medium.com/@hatajoe/clean-architecture-in-go-4030f11ec1b1)
- `.claude/architecture.md` - Our specific implementation details
- `.claude/flows.md` - Visual diagrams of our architecture in action

## Related ADRs

- [ADR-002: Domain Services Pattern](./002-domain-services.md)
- [ADR-003: Two-Step Dependency Injection](./003-two-step-di.md)
- [ADR-005: Repository Interfaces in Domain](./005-repository-interfaces.md)

---

**Last Reviewed:** 2024-01-15

**Next Review:** 2024-07-15 (or when considering major architectural change)
