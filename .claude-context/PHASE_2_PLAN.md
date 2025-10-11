# Phase 2 Implementation Plan: Bounded Context Organization

**Date:** October 11, 2025
**Status:** Planning
**Expected Effort:** 1-2 weeks
**Risk Level:** Medium (significant file moves, but incremental approach reduces risk)

---

## Overview

**Goal:** Reorganize codebase from horizontal layering to bounded context organization, achieving 40-50% additional token reduction while maintaining Clean Architecture principles.

**Approach:** Hybrid bounded context organization (less radical than full vertical slicing)

**Key Principle:** Existing projects with established patterns shouldn't do big-bang rewrites. We'll adopt a pragmatic hybrid approach that groups code by bounded context while maintaining layer separation within each context.

---

## Current State vs Target State

### Current Structure (Horizontal Layering)
```
internal/
├── domain/              # All domains mixed
│   ├── book/
│   ├── member/
│   ├── payment/
│   └── reservation/
├── usecase/             # All use cases mixed
│   ├── bookops/
│   ├── authops/
│   ├── paymentops/
│   └── reservationops/
├── adapters/            # All adapters mixed
│   ├── http/handlers/
│   │   ├── book/
│   │   ├── auth/
│   │   ├── payment/
│   │   └── reservation/
│   ├── repository/postgres/
│   └── payment/epayment/
└── infrastructure/      # Technical concerns
    ├── auth/
    ├── store/
    └── server/
```

**Token Cost for Feature Change:** 5,000-8,000 tokens (traverse 3-4 directories)

### Target Structure (Bounded Context Organization)
```
internal/
├── books/                    # Books bounded context
│   ├── domain/              # Entities, services, repository interfaces
│   ├── operations/          # Use cases (create, get, update, delete, list)
│   ├── http/                # HTTP handlers, DTOs
│   └── repository/          # PostgreSQL implementation
├── members/                  # Members bounded context
│   ├── domain/
│   ├── operations/          # Auth + member operations
│   ├── http/
│   └── repository/
├── payments/                 # Payments bounded context
│   ├── domain/              # Payment, SavedCard, Receipt entities
│   ├── operations/          # Payment, SavedCard, Receipt use cases
│   ├── http/                # Handlers for all payment features
│   ├── repository/          # DB implementations
│   └── gateway/             # epayment.kz adapter
├── reservations/             # Reservations bounded context
│   ├── domain/
│   ├── operations/
│   ├── http/
│   └── repository/
├── shared/                   # Shared infrastructure
│   ├── auth/                # JWT service
│   ├── store/               # Database connections
│   ├── server/              # HTTP server
│   └── middleware/          # HTTP middleware
└── pkg/                      # Shared utilities (unchanged)
    ├── errors/
    ├── httputil/
    ├── logutil/
    └── strutil/
```

**Token Cost for Feature Change:** 2,500-4,000 tokens (40-50% reduction)

---

## Benefits

### Token Efficiency
- **Reduced directory traversal:** All book-related code in one tree
- **Clear context boundaries:** Loading context stays within bounded context 90% of time
- **Fewer imports:** Related code is closer together
- **Expected savings:** 40-50% token reduction per task

### Code Organization
- **Improved discoverability:** All book code in `internal/books/`
- **Clear ownership:** Each bounded context is self-contained
- **Easier navigation:** Feature changes touch 1-2 directories instead of 3-4
- **Maintains Clean Architecture:** Layers preserved within each context

### Development Velocity
- **Faster feature additions:** All related code in one place
- **Reduced cognitive load:** Less mental context switching
- **Better IDE navigation:** Go-to-definition stays within bounded context
- **Clearer dependencies:** Bounded contexts have explicit interfaces

---

## Migration Strategy

### Phase 2.1: Pilot with Books Bounded Context (Week 1)

**Goal:** Validate approach with smallest, most stable domain

**Steps:**
1. Create `internal/books/` directory structure
2. Move files from current locations to new structure
3. Update import paths across codebase
4. Update container.go to use new paths
5. Update router.go to use new paths
6. Run tests to ensure no breakage
7. Update documentation and examples

**Files to Move:**
- `internal/domain/book/` → `internal/books/domain/`
- `internal/domain/author/` → `internal/books/domain/` (authors are part of books context)
- `internal/usecase/bookops/` → `internal/books/operations/`
- `internal/adapters/http/handlers/book/` → `internal/books/http/`
- `internal/adapters/repository/postgres/book.go` → `internal/books/repository/book.go`
- `internal/adapters/repository/postgres/author.go` → `internal/books/repository/author.go`

**Validation Criteria:**
- ✅ All tests pass
- ✅ Build succeeds
- ✅ API endpoints work (manual test)
- ✅ Token consumption reduced by 40-50% for book feature changes
- ✅ No performance degradation

**Rollback Plan:** Keep original files until validation complete, then delete

### Phase 2.2: Migrate Members Bounded Context (Week 1-2)

**Goal:** Migrate auth and member operations (tightly coupled)

**Steps:**
1. Create `internal/members/` directory structure
2. Move member and auth domain entities
3. Move auth and member use cases to operations/
4. Move HTTP handlers
5. Move repository implementations
6. Update imports, container, router
7. Run tests and validate

**Files to Move:**
- `internal/domain/member/` → `internal/members/domain/`
- `internal/usecase/authops/` → `internal/members/operations/auth/`
- `internal/usecase/memberops/` → `internal/members/operations/profile/`
- `internal/usecase/subops/` → `internal/members/operations/subscription/`
- `internal/adapters/http/handlers/auth/` → `internal/members/http/auth/`
- `internal/adapters/http/handlers/member/` → `internal/members/http/profile/`
- `internal/adapters/repository/postgres/member.go` → `internal/members/repository/member.go`

**Note:** Auth is part of members context (members authenticate)

### Phase 2.3: Migrate Payments Bounded Context (Week 2)

**Goal:** Migrate largest, most complex domain

**Steps:**
1. Create `internal/payments/` directory structure
2. Move payment, saved card, receipt domains
3. Move all payment operations
4. Move HTTP handlers
5. Move repository and gateway implementations
6. Update imports, container, router
7. Run tests and validate

**Files to Move:**
- `internal/domain/payment/` → `internal/payments/domain/`
- `internal/usecase/paymentops/` → `internal/payments/operations/payment/`
- `internal/usecase/savedcardops/` → `internal/payments/operations/savedcard/` (if exists)
- `internal/usecase/receiptops/` → `internal/payments/operations/receipt/` (if exists)
- `internal/adapters/http/handlers/payment/` → `internal/payments/http/payment/`
- `internal/adapters/http/handlers/receipt/` → `internal/payments/http/receipt/`
- `internal/adapters/http/handlers/savedcard/` → `internal/payments/http/savedcard/`
- `internal/adapters/repository/postgres/payment*.go` → `internal/payments/repository/`
- `internal/adapters/payment/epayment/` → `internal/payments/gateway/epayment/`

**Special Considerations:**
- Payment has 18+ use cases (largest domain)
- External gateway integration needs careful testing
- Background worker depends on payment domain

### Phase 2.4: Migrate Reservations Bounded Context (Week 2)

**Goal:** Migrate final bounded context

**Steps:**
1. Create `internal/reservations/` directory structure
2. Move reservation domain
3. Move reservation operations
4. Move HTTP handlers
5. Move repository implementations
6. Update imports, container, router
7. Run tests and validate

**Files to Move:**
- `internal/domain/reservation/` → `internal/reservations/domain/`
- `internal/usecase/reservationops/` → `internal/reservations/operations/`
- `internal/adapters/http/handlers/reservation/` → `internal/reservations/http/`
- `internal/adapters/repository/postgres/reservation.go` → `internal/reservations/repository/reservation.go`

### Phase 2.5: Cleanup and Documentation (Week 2)

**Steps:**
1. Delete old empty directories
2. Update all documentation files:
   - CLAUDE.md
   - .claude-context/SESSION_MEMORY.md
   - .claude-context/CURRENT_PATTERNS.md
   - examples/ (update import paths)
   - .claude/architecture.md
3. Update .claudeignore if needed
4. Regenerate Swagger docs
5. Update README.md
6. Create Phase 2 completion report

---

## Detailed File Moves (Books Pilot)

### Step-by-Step Migration

**1. Create Directory Structure:**
```bash
mkdir -p internal/books/domain
mkdir -p internal/books/operations
mkdir -p internal/books/http
mkdir -p internal/books/repository
```

**2. Move Domain Layer:**
```bash
# Move book domain
git mv internal/domain/book/* internal/books/domain/
# Move author domain (authors belong to books context)
git mv internal/domain/author/* internal/books/domain/
```

**3. Move Use Case Layer:**
```bash
# Move all book operations
git mv internal/usecase/bookops/* internal/books/operations/
```

**4. Move HTTP Handlers:**
```bash
# Move book handlers
git mv internal/adapters/http/handlers/book/* internal/books/http/
```

**5. Move Repository Layer:**
```bash
# Move book repository
git mv internal/adapters/repository/postgres/book.go internal/books/repository/book.go
git mv internal/adapters/repository/postgres/author.go internal/books/repository/author.go
```

**6. Update Import Paths:**

Files to update (search and replace):
- All files in `internal/books/` (update relative imports)
- `internal/usecase/container.go` (import book operations)
- `internal/adapters/http/router.go` (import book handlers)
- Any test files that import book code

Search pattern: `"library-service/internal/domain/book"`
Replace with: `"library-service/internal/books/domain"`

Search pattern: `"library-service/internal/usecase/bookops"`
Replace with: `"library-service/internal/books/operations"`

Search pattern: `"library-service/internal/adapters/http/handlers/book"`
Replace with: `"library-service/internal/books/http"`

**7. Update Container:**

In `internal/usecase/container.go`:
```go
// Before
import "library-service/internal/usecase/bookops"

// After
import bookops "library-service/internal/books/operations"
```

**8. Update Router:**

In `internal/adapters/http/router.go`:
```go
// Before
import "library-service/internal/adapters/http/handlers/book"

// After
import bookhandler "library-service/internal/books/http"
```

---

## Testing Strategy

### Pre-Migration Baseline
```bash
# Record baseline metrics
make test                  # All tests pass
make build                 # Builds successfully
make lint                  # No lint errors

# Record API test results
./scripts/test-api.sh      # All endpoints work

# Record token consumption (estimate from TOKEN_LOG.md)
# Feature change: 5,000-8,000 tokens currently
```

### Post-Migration Validation (After Each Bounded Context)
```bash
# Run full test suite
make test

# Build all binaries
make build

# Run linter
make lint

# Manual API testing
make dev
# Test all endpoints in migrated bounded context
# Example: For books, test:
# - POST /api/v1/books (create)
# - GET /api/v1/books/:id (get)
# - GET /api/v1/books (list)
# - PUT /api/v1/books/:id (update)
# - DELETE /api/v1/books/:id (delete)

# Integration tests (if applicable)
make test-integration

# Verify imports
go list -f '{{.ImportPath}} {{.Imports}}' ./internal/books/...
# Ensure no broken imports
```

### Token Consumption Measurement
After each migration, measure token consumption for typical tasks:
- Add new book field: Before ~5,000-8,000 tokens → Target ~2,500-4,000 tokens
- Fix book validation bug: Before ~3,000-5,000 tokens → Target ~1,500-2,500 tokens
- Add new book endpoint: Before ~6,000-10,000 tokens → Target ~3,000-5,000 tokens

**Record results in TOKEN_LOG.md**

---

## Risks and Mitigation

### Risk 1: Import Path Breakage
**Impact:** High (compilation failure)
**Probability:** High (many imports to update)
**Mitigation:**
- Use git mv to preserve history
- Use global search/replace for import paths
- Run `go build` frequently during migration
- Keep original directory until validation complete

### Risk 2: Circular Dependencies
**Impact:** High (compilation failure)
**Probability:** Low (Clean Architecture prevents this)
**Mitigation:**
- Bounded contexts should not import each other
- Shared code goes in `shared/` or `pkg/`
- Review import graph: `go list -f '{{.ImportPath}} {{.Imports}}' ./internal/...`

### Risk 3: Test Failures
**Impact:** High (broken functionality)
**Probability:** Medium (tests may hardcode paths)
**Mitigation:**
- Run tests after each file move
- Update test fixtures with new paths
- Update integration test configurations

### Risk 4: Container Wiring Errors
**Impact:** High (runtime panic)
**Probability:** Medium (complex dependency graph)
**Mitigation:**
- Update container.go carefully
- Test DI container initialization explicitly
- Manual smoke test after container changes

### Risk 5: Token Savings Don't Materialize
**Impact:** Low (wasted effort, but no breakage)
**Probability:** Low (research-backed approach)
**Mitigation:**
- Measure token consumption before and after pilot
- If savings < 30%, reconsider full migration
- Pilot with Books validates approach before full commitment

### Risk 6: Regression in Functionality
**Impact:** Critical (production issues)
**Probability:** Low (thorough testing)
**Mitigation:**
- Comprehensive test coverage (current: 60%+)
- Manual API testing checklist
- Staged rollout (one bounded context at time)
- Keep git history for easy rollback

---

## Success Criteria

### Must Have (Phase 2 Success)
- ✅ All tests pass (unit + integration)
- ✅ Build succeeds with no errors
- ✅ Linter passes with no new warnings
- ✅ All API endpoints work correctly
- ✅ Token consumption reduced by 40-50% for feature changes
- ✅ No performance degradation
- ✅ Documentation updated (CLAUDE.md, examples/, .claude-context/)

### Nice to Have (Stretch Goals)
- ✅ Token consumption reduced by 50-60%
- ✅ Improved build times (fewer imports to resolve)
- ✅ Clearer bounded context boundaries
- ✅ Updated architecture diagrams in .claude/

---

## Decision Points

### After Books Pilot (Week 1)
**Decision:** Continue with full migration or revert?

**Continue if:**
- ✅ Token savings ≥ 40%
- ✅ All tests pass
- ✅ Team (user) approves new structure
- ✅ No significant development velocity impact

**Revert if:**
- ❌ Token savings < 30%
- ❌ Tests fail and cannot be fixed easily
- ❌ Significant regressions
- ❌ Team prefers current structure

### After Week 1 (Books + Members Complete)
**Decision:** Continue with Payments and Reservations?

**Continue if:**
- ✅ Consistent token savings across both contexts
- ✅ Migration process is smooth
- ✅ No major blockers

**Pause/Adjust if:**
- ⚠️ Token savings inconsistent
- ⚠️ Migration taking longer than expected
- ⚠️ Team velocity impacted

---

## Rollback Plan

If migration fails at any stage:

1. **Revert Git Commits:**
   ```bash
   git revert <commit-hash>  # Revert import updates
   git revert <commit-hash>  # Revert file moves
   ```

2. **Restore Original Structure:**
   ```bash
   git mv internal/books/* <original-locations>
   # Or git reset --hard <pre-migration-commit>
   ```

3. **Validate Rollback:**
   ```bash
   make test
   make build
   make lint
   ```

4. **Document Learnings:**
   - Update TOKEN_LOG.md with findings
   - Document why migration failed
   - Adjust plan if retry is warranted

---

## Timeline

**Week 1:**
- Day 1-2: Books bounded context migration + validation
- Day 3-4: Members bounded context migration + validation
- Day 5: Review progress, decide to continue

**Week 2:**
- Day 1-2: Payments bounded context migration + validation
- Day 3: Reservations bounded context migration + validation
- Day 4: Cleanup, documentation updates
- Day 5: Final validation, completion report

**Total Effort:** 8-10 days of focused work

---

## Open Questions

1. **Should `shared/` be `internal/shared/` or `pkg/shared/`?**
   - Recommendation: `internal/shared/` (not exported)

2. **How to handle cross-context dependencies?**
   - Recommendation: Bounded contexts should NOT import each other
   - Use events or shared DTOs in `internal/shared/` if needed

3. **Should we keep `pkg/` utilities or move to `internal/shared/pkg/`?**
   - Recommendation: Keep current `pkg/` (already well-organized)

4. **Update examples/ to use new structure?**
   - Recommendation: Yes, update import paths in Phase 2.5

---

## Next Steps

**Immediate:**
1. Review this plan with user for approval
2. Create git branch: `feature/phase-2-bounded-contexts`
3. Start with Books pilot migration

**After Approval:**
1. Execute Phase 2.1 (Books pilot)
2. Measure token consumption
3. Validate success criteria
4. Decide to continue or adjust

---

**Created:** October 11, 2025
**Status:** Awaiting approval
**Estimated Impact:** 40-50% token reduction for feature changes
**Risk:** Medium (significant refactoring, but incremental approach reduces risk)
