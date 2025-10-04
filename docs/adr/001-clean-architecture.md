# ADR 001: Adopt Clean Architecture Pattern

**Date**: 2025-10-04
**Status**: Accepted
**Decision Makers**: Development Team

## Context

The Library Management System needs a maintainable, testable architecture that:
- Separates business logic from infrastructure concerns
- Allows easy testing without external dependencies
- Supports changing frameworks/databases without rewriting core logic
- Scales as the application grows
- Enables vibecoding workflow with Claude Code

## Decision

We will adopt **Clean Architecture** (also known as Hexagonal Architecture or Onion Architecture) with the following layers:

### Layer Structure

```
┌─────────────────────────────────────┐
│   Infrastructure & Frameworks      │  ← External dependencies
│   (Gin, PostgreSQL, Redis)         │
├─────────────────────────────────────┤
│   Adapters (HTTP, Repository)      │  ← Interface implementations
├─────────────────────────────────────┤
│   Use Cases (Application Logic)    │  ← Orchestration
├─────────────────────────────────────┤
│   Domain (Business Logic)           │  ← Core business rules
└─────────────────────────────────────┘
```

### Dependency Rule

Dependencies point **inward only**:
- Domain has NO dependencies (pure business logic)
- Use Cases depend on Domain
- Adapters depend on Domain (via interfaces)
- Infrastructure depends on Adapters

### File Organization

```
internal/
├── domain/          # Business entities, rules, interfaces
├── usecase/         # Application use cases
└── adapters/        # External interfaces (HTTP, DB)
```

## Consequences

### Positive

✅ **Testability**: Domain and use cases can be tested without database/HTTP
✅ **Independence**: Business logic independent of frameworks
✅ **Flexibility**: Can swap PostgreSQL for MySQL without changing domain
✅ **Maintainability**: Clear separation of concerns
✅ **Vibecoding**: Claude Code can navigate layers intuitively
✅ **Team Productivity**: Developers can work on different layers in parallel

### Negative

❌ **Complexity**: More files and abstractions than MVC
❌ **Learning Curve**: Team needs to understand layer boundaries
❌ **Boilerplate**: DTOs and mappers add code overhead
❌ **Initial Effort**: Takes longer to set up than simple architecture

### Mitigations

- Comprehensive documentation and examples
- Code templates for new features
- Clear package-level documentation
- Regular architecture reviews

## Alternatives Considered

### 1. MVC (Model-View-Controller)
- ❌ Business logic often bleeds into controllers
- ❌ Hard to test without HTTP layer
- ❌ Tight coupling to framework

### 2. Layered Architecture
- ❌ Layers often depend on lower layers (database leaking up)
- ❌ Business logic scattered across layers
- ✅ Simpler to understand

### 3. Domain-Driven Design (DDD) with Aggregates
- ✅ Strong domain modeling
- ❌ Higher complexity for this project size
- ❌ Steeper learning curve

## Implementation Notes

1. **Domain Layer** (`internal/domain/`)
   - Entities, value objects, domain services
   - Repository and cache interfaces (defined here, implemented elsewhere)
   - No external dependencies

2. **Use Case Layer** (`internal/usecase/`)
   - One use case per file (single responsibility)
   - Orchestrates domain services and repositories
   - Input/output DTOs

3. **Adapter Layer** (`internal/adapters/`)
   - HTTP handlers (Gin)
   - Repository implementations (PostgreSQL)
   - Cache implementations (Redis)

4. **Dependency Injection**
   - Constructor injection throughout
   - Wire up in `cmd/api/main.go`

## References

- [The Clean Architecture (Uncle Bob)](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Hexagonal Architecture](https://alistair.cockburn.us/hexagonal-architecture/)
- [Package by Feature vs Package by Layer](https://medium.com/@benbjohnson/structuring-applications-in-go-3b04be4ff091)

## Review Notes

This decision was made to support:
- Long-term maintainability
- Easy onboarding for new developers
- Effective vibecoding with Claude Code
- Test-driven development workflow

**Next Review**: After 6 months or 50+ use cases
