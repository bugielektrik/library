# Architecture Decision Records (ADRs)

**Documenting significant architectural decisions and their rationale.**

## What are ADRs?

Architecture Decision Records (ADRs) document important architectural decisions made during the development of the Library Management System. Each ADR captures:
- **Context**: Why the decision was needed
- **Decision**: What was decided
- **Consequences**: Impact of the decision (positive and negative)
- **Alternatives**: Other options considered

## ADR Index

| Number | Title | Status | Date |
|--------|-------|--------|------|
| [001](./001-clean-architecture.md) | Adopt Clean Architecture Pattern | Accepted | 2025-10-04 |
| [002](./002-domain-services.md) | Introduce Domain Services for Business Logic | Accepted | 2025-10-04 |
| [003](./003-dependency-injection.md) | Constructor-Based Dependency Injection | Accepted | 2025-10-04 |

## When to Create an ADR

Create an ADR when making decisions about:

### Architecture & Design
- ✅ Architectural patterns (Clean Architecture, Microservices, etc.)
- ✅ Layer organization and boundaries
- ✅ Design patterns (Repository, Factory, etc.)
- ✅ Domain modeling approaches

### Technology Choices
- ✅ Framework selection (Gin, Echo, etc.)
- ✅ Database choice (PostgreSQL, MySQL, etc.)
- ✅ Infrastructure tools (Docker, Kubernetes, etc.)
- ✅ Third-party libraries

### Process & Standards
- ✅ Testing strategies
- ✅ Deployment approaches
- ✅ Error handling conventions
- ✅ API design standards

### What NOT to Document
- ❌ Implementation details (code-level decisions)
- ❌ Temporary workarounds
- ❌ Routine bug fixes
- ❌ Configuration changes

## ADR Template

```markdown
# ADR XXX: [Title]

**Date**: YYYY-MM-DD
**Status**: [Proposed | Accepted | Deprecated | Superseded]
**Decision Makers**: [Team/Individual]

## Context

[Why is this decision needed? What problem are we solving?]

## Decision

[What did we decide to do?]

## Consequences

### Positive

✅ [Benefit 1]
✅ [Benefit 2]

### Negative

❌ [Drawback 1]
❌ [Drawback 2]

### Mitigations

[How we address the negatives]

## Alternatives Considered

### 1. [Alternative Option]
- [Pros]
- [Cons]
- [Why rejected]

## References

- [Link to documentation]
- [Related articles]

## Review Notes

[When to review this decision]
```

## ADR Statuses

- **Proposed**: Under discussion, not yet implemented
- **Accepted**: Decision made and implemented
- **Deprecated**: No longer recommended, but still in use
- **Superseded**: Replaced by a newer ADR (link to new ADR)

## Creating a New ADR

### 1. Determine ADR Number

```bash
# Next number is 004
ls docs/adr/*.md | wc -l
```

### 2. Create ADR File

```bash
# Use template
cp docs/adr/000-template.md docs/adr/004-your-decision.md
```

### 3. Fill in Template

- **Context**: Explain the situation and constraints
- **Decision**: Clearly state what was decided
- **Consequences**: List pros and cons honestly
- **Alternatives**: Show you considered other options
- **References**: Link to supporting materials

### 4. Review Process

1. Draft ADR and share with team
2. Discuss in team meeting or async
3. Incorporate feedback
4. Mark as "Accepted" when consensus reached
5. Update ADR index in this README

### 5. Implementation

- Reference ADR in relevant code comments
- Update architecture documentation
- Communicate decision to stakeholders

## Example ADRs

### Good ADR Example

```markdown
# ADR 001: Adopt Clean Architecture Pattern

**Context**: We need maintainable, testable architecture...

**Decision**: We will adopt Clean Architecture with 4 layers...

**Consequences**:
✅ Testability without external dependencies
❌ More complexity than MVC

**Alternatives**:
1. MVC - simpler but less testable
2. Microservices - overkill for our size
```

**Why it's good**:
- Clear context
- Specific decision
- Honest about tradeoffs
- Shows alternatives were considered

### Poor ADR Example

```markdown
# ADR: Use PostgreSQL

**Decision**: We're using PostgreSQL.

**Why**: Because it's good.
```

**Why it's poor**:
- No context
- No consequences
- No alternatives
- Too vague

## Referencing ADRs

### In Code Comments

```go
// Book domain service implements business rules.
// See ADR-002 for rationale on domain services.
type Service struct {}
```

### In Documentation

```markdown
The project uses Clean Architecture (see [ADR-001](./adr/001-clean-architecture.md))
with domain services ([ADR-002](./adr/002-domain-services.md)) to encapsulate
business logic.
```

### In Pull Requests

```markdown
## Summary
Refactored subscription pricing to domain service

## References
- Implements ADR-002 (Domain Services)
- Follows ADR-003 (Dependency Injection)
```

## Updating ADRs

ADRs are **immutable** once accepted. To change a decision:

### 1. Create New ADR

```markdown
# ADR 004: Switch from PostgreSQL to MySQL

**Status**: Accepted
**Supersedes**: ADR-001 (database choice section)

[Explain why we're changing and what's different]
```

### 2. Update Old ADR

```markdown
# ADR 001: Original Database Choice

**Status**: Superseded by ADR-004
**Date Superseded**: 2025-11-01
```

### 3. Update Index

Mark old ADR as superseded in index table.

## ADR Governance

### Review Cycle

- **New Projects**: Review all ADRs at 3 months, 6 months
- **Mature Projects**: Annual review
- **As Needed**: When facing similar decisions

### Deprecation Process

1. Create new ADR documenting new approach
2. Mark old ADR as "Deprecated"
3. Plan migration timeline
4. Update all references

### Team Process

1. **Major Decisions**: Team discussion + ADR
2. **Minor Decisions**: Lead approval + ADR
3. **Urgent Decisions**: Create ADR retroactively

## Benefits of ADRs

✅ **Knowledge Preservation**: Rationale survives beyond original team
✅ **Onboarding**: New developers understand "why" not just "what"
✅ **Decision Quality**: Forces thorough consideration of alternatives
✅ **Vibecoding**: Claude Code can reference ADRs for context
✅ **Avoid Repeating**: Reference past decisions when similar issues arise

## Tools

### Create ADR
```bash
# Create from template
./scripts/new-adr.sh "Your Decision Title"
```

### List ADRs
```bash
# List all ADRs
ls -1 docs/adr/*.md | grep -v README
```

### Search ADRs
```bash
# Find ADRs about a topic
grep -r "PostgreSQL" docs/adr/
```

## Related Resources

- [ADR GitHub](https://adr.github.io/) - ADR best practices
- [Documenting Architecture Decisions](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions) - Original proposal
- [ADR Tools](https://github.com/npryce/adr-tools) - Command-line tools

## Contributing

When making significant architectural changes:
1. Create an ADR **before** implementation
2. Discuss with team
3. Get approval
4. Implement
5. Reference ADR in code and docs

**Remember**: ADRs are living documentation. Keep them updated and refer to them often!
