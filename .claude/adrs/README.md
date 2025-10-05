# Architecture Decision Records (ADRs)

> **Why these decisions were made, not just what they are**

## What are ADRs?

Architecture Decision Records document the key architectural decisions made in this project, including:
- The context and problem
- The decision made
- The consequences (both positive and negative)
- Alternatives considered

**Purpose for AI-assisted development:** Understanding WHY decisions were made prevents suggesting changes that contradict the project's philosophy.

## Index of Decisions

### Core Architecture

- **[ADR-001: Clean Architecture](./001-clean-architecture.md)** - Why we use Clean Architecture/Hexagonal pattern
- **[ADR-002: Domain Services Pattern](./002-domain-services.md)** - Why business logic lives in domain services
- **[ADR-003: Two-Step Dependency Injection](./003-two-step-di.md)** - Why app.go + container.go pattern

### Naming and Organization

- **[ADR-004: "ops" Suffix Convention](./004-ops-suffix-convention.md)** - Why use case packages end with "ops"
- **[ADR-005: Repository Interfaces in Domain](./005-repository-interfaces.md)** - Why interfaces are defined in domain layer

### Technical Decisions

- **[ADR-006: PostgreSQL as Primary Database](./006-postgresql.md)** - Why PostgreSQL over other databases
- **[ADR-007: JWT for Authentication](./007-jwt-authentication.md)** - Why JWT tokens over sessions

## How to Read ADRs

**New to the project?** Read in order (001 → 007) to understand the foundational thinking.

**Considering a change?** Check if there's an ADR about it. If the ADR still applies, don't fight it. If context has changed, propose updating the ADR.

**Adding a new major decision?** Create a new ADR following the template below.

## ADR Template

```markdown
# ADR-XXX: Title

**Status:** Accepted | Deprecated | Superseded by ADR-YYY

**Date:** YYYY-MM-DD

**Context:** What problem are we trying to solve? What constraints exist?

**Decision:** What did we decide to do?

**Consequences:**
- **Positive:** What benefits does this bring?
- **Negative:** What costs or limitations does this create?

**Alternatives Considered:**
- Alternative 1: Why we didn't choose this
- Alternative 2: Why we didn't choose this

**References:**
- Links to discussions, articles, etc.
```

## Status Meanings

- **Accepted:** This is currently how we do things
- **Deprecated:** No longer recommended, but not yet removed
- **Superseded:** Replaced by a newer ADR

---

**Remember:** ADRs are not rules set in stone. They're snapshots of our best thinking at the time. If context changes, we can change our decisions—but we should document WHY.
