# Documentation Refactoring Plan

**Date:** October 11, 2025
**Current State:** 77 markdown files in `.claude/`, ~23,000 lines of documentation
**Goal:** Streamline documentation for immediate Claude Code productivity

---

## Problems Identified

### 1. **Excessive Historical Documentation** (30+ files, ~10,000 lines)
Many refactoring completion summaries that are no longer relevant:
- Multiple "PHASE" completion files
- Multiple "REFACTORING" summary files
- Pattern analysis files that have been completed

### 2. **Duplicate ADR Directories**
- `.claude/adr/` - 7 ADRs (newer, 2025)
- `.claude/adrs/` - 11 ADRs (older, 2024)
- These are NOT duplicates but different decisions
- Confusing for future instances

### 3. **Redundant Pattern Documentation**
- `CODE_PATTERN_STANDARDS.md`
- `USECASE_PATTERN_STANDARDS.md`
- `DOMAIN_PATTERN.md`
- Some overlap with `CURRENT_PATTERNS.md` in `.claude-context/`

### 4. **Duplicate Example Directories**
- `.claude/examples/` (4 files)
- `/examples/` (4 files)
- Likely duplicates

### 5. **Outdated Root-Level Assessment Files**
- `REFACTORING_ASSESSMENT.md` - Historical analysis
- `POST_REFACTORING_ASSESSMENT.md` - Validation completed
- These should be archived

### 6. **Too Many Entry Points**
- `CLAUDE.md` (main)
- `.claude/README.md`
- `.claude/AI-QUICKSTART.md`
- `.claude/QUICKSTART.md`
- Multiple paths to the same information

---

## Proposed Structure

```
/
├── CLAUDE.md                          # ✅ KEEP - Main entry point
├── README.md                          # ✅ KEEP - Project README
│
├── .claude/                           # Active documentation (reduced to ~15 files)
│   ├── README.md                      # Quick navigation guide
│   │
│   ├── guides/                        # Core guides (7 files)
│   │   ├── architecture.md
│   │   ├── development.md             # Renamed from development-guide.md
│   │   ├── testing.md
│   │   ├── common-tasks.md
│   │   ├── coding-standards.md
│   │   ├── security.md
│   │   └── cache-warming.md
│   │
│   ├── adr/                           # Architecture Decision Records (merged)
│   │   ├── README.md
│   │   ├── 001-clean-architecture.md
│   │   ├── 002-domain-services.md
│   │   ├── 003-two-step-di.md
│   │   ├── 004-ops-suffix-convention.md
│   │   ├── 005-repository-interfaces.md
│   │   ├── 006-postgresql.md
│   │   ├── 007-jwt-authentication.md
│   │   ├── 008-generic-repository-helpers.md
│   │   ├── 009-payment-gateway-modularization.md
│   │   ├── 010-domain-service-payment-status.md
│   │   ├── 011-base-repository-pattern.md
│   │   ├── 012-bounded-context-organization.md
│   │   └── 013-dto-colocation-and-token-optimization.md
│   │
│   └── reference/                     # Reference materials (4 files)
│       ├── README.md
│       ├── common-mistakes.md         # Renamed from COMMON-MISTAKES.md
│       ├── error-handling.md          # Renamed from ERROR-HANDLING-GUIDE.md
│       └── migration-guide.md         # Renamed from MIGRATION-GUIDE-REPOSITORIES.md
│
├── .claude-context/                   # Session context (reduced to 3 files)
│   ├── README.md
│   ├── SESSION_MEMORY.md              # ✅ KEEP - Essential
│   └── CURRENT_PATTERNS.md            # ✅ KEEP - Essential
│
├── docs/                              # Domain-specific documentation
│   ├── payments/                      # ✅ NEW - Organize payment docs
│   │   ├── README.md                  # Overview
│   │   ├── integration.md             # Renamed from PAYMENT_INTEGRATION.md
│   │   ├── quick-start.md             # Renamed from PAYMENT_QUICK_START.md
│   │   ├── features.md                # Renamed from PAYMENT_FEATURES.md
│   │   ├── api-integration.md         # Renamed from EPAYMENT_API_INTEGRATION.md
│   │   └── swagger-api.md             # Renamed from SWAGGER_PAYMENT_API.md
│   │
│   └── archive/                       # ✅ NEW - Historical documents
│       ├── README.md                  # Index of archived docs
│       ├── refactoring-assessment.md  # From root
│       └── post-refactoring.md        # From root
│
├── examples/                          # ✅ KEEP - Code pattern examples (4 files)
│   ├── README.md
│   ├── handler_pattern.md
│   ├── usecase_pattern.md
│   ├── repository_pattern.md
│   └── testing_pattern.md
│
└── [other directories unchanged]
```

---

## Files to Remove (33 files)

### Historical Refactoring Documents (28 files)
All completed, no longer needed:

```
.claude/CODEBASE-ANALYSIS-2025-10-09.md
.claude/CODEBASE_PATTERN_REFACTORING.md
.claude/COMPLETE_USECASE_REFACTORING.md
.claude/DOMAIN_REFACTORING_SUMMARY.md
.claude/HANDLER_PATTERN_ANALYSIS.md
.claude/HANDLER_PATTERN_COMPLETE.md
.claude/HANDLER_REFACTORING_FINAL.md
.claude/HANDLER_REFACTORING_SUMMARY.md
.claude/LEGACY_CODE_REMOVAL.md
.claude/PHASE1-QUICK-WINS-COMPLETE.md
.claude/PHASE2-COMPLETE.md
.claude/PHASE3A-CLEANUP-COMPLETE.md
.claude/PHASE3B-DUPLICATION-COMPLETE.md
.claude/PHASE4_REFACTORING_SUMMARY.md
.claude/REFACTORING-OPPORTUNITIES.md
.claude/REFACTORING-PHASE3-RECOMMENDATIONS.md
.claude/REFACTORING-RECOMMENDATIONS-2.md
.claude/REFACTORING-STATUS-2025-10-09.md
.claude/USECASE_REFACTORING_PLAN.md
.claude/USECASE_REFACTORING_SUMMARY.md
.claude/refactoring-complete-summary.md
.claude/refactoring-phase3c-summary.md
.claude/refactoring-phase4-plan.md
.claude/refactoring-phase4-progress.md
.claude/refactoring-phase4-visual.md
.claude/refactoring-phase4a-quickstart.md
.claude/refactoring-phase4a-summary.md
.claude/refactoring-phase4b-summary.md
.claude/refactoring-phase4c-summary.md
.claude/refactoring-phase4d-summary.md
```

### Redundant/Duplicate Files (5 files)
```
.claude/AI-QUICKSTART.md                # Redundant with README.md
.claude/QUICKSTART.md                   # Redundant with README.md
.claude/examples/                       # Duplicate of /examples/
.claude-context/PHASE_2.1_COMPLETE.md   # Historical
.claude-context/PHASE_2_PLAN.md         # Historical
.claude-context/TOKEN_LOG.md            # No longer useful
```

---

## Files to Move/Rename (28 files)

### To .claude/guides/
```
.claude/architecture.md                 → .claude/guides/architecture.md
.claude/development-guide.md            → .claude/guides/development.md
.claude/common-tasks.md                 → .claude/guides/common-tasks.md
.claude/coding-standards.md             → .claude/guides/coding-standards.md
.claude/testing.md                      → .claude/guides/testing.md
.claude/security.md                     → .claude/guides/security.md
.claude/cache-warming.md                → .claude/guides/cache-warming.md
```

### To .claude/reference/
```
.claude/COMMON-MISTAKES.md              → .claude/reference/common-mistakes.md
.claude/ERROR-HANDLING-GUIDE.md         → .claude/reference/error-handling.md
.claude/MIGRATION-GUIDE-REPOSITORIES.md → .claude/reference/migration-guide.md
.claude/GO-ONBOARDING.md                → .claude/reference/go-onboarding.md
```

### Merge ADR Directories
```
.claude/adrs/* (11 files)               → .claude/adr/ (merge & renumber if needed)
# Keep .claude/adr/ as the canonical location
# Add ADRs 006-011 from adrs/ with new numbers or merge logically
```

### To docs/payments/
```
docs/PAYMENT_INTEGRATION.md             → docs/payments/integration.md
docs/PAYMENT_QUICK_START.md             → docs/payments/quick-start.md
docs/PAYMENT_FEATURES.md                → docs/payments/features.md
docs/EPAYMENT_API_INTEGRATION.md        → docs/payments/api-integration.md
docs/SWAGGER_PAYMENT_API.md             → docs/payments/swagger-api.md
```

### To docs/archive/
```
REFACTORING_ASSESSMENT.md               → docs/archive/refactoring-assessment.md
POST_REFACTORING_ASSESSMENT.md          → docs/archive/post-refactoring.md
```

---

## Files to Keep As-Is (15 files)

```
CLAUDE.md                               # Main entry point
README.md                               # Project README
examples/*                              # All example files
.claude-context/README.md               # Context guide
.claude-context/SESSION_MEMORY.md       # Essential architecture
.claude-context/CURRENT_PATTERNS.md     # Essential patterns
internal/*/README.md                    # Package documentation
test/*/README.md                        # Test documentation
migrations/README.md                    # Migration guide
scripts/README.md                       # Script documentation
cmd/README.md                           # Command documentation
pkg/README.md                           # Package documentation
```

---

## Impact Analysis

### Before
- **77 files** in `.claude/` and `.claude-context/`
- **~23,000 lines** of documentation
- Multiple entry points
- Historical cruft

### After
- **~25 files** in `.claude/` and `.claude-context/` (68% reduction)
- **~8,000 lines** estimated (65% reduction)
- Clear structure with 3 main sections:
  - `guides/` - How to work with the project
  - `adr/` - Why decisions were made
  - `reference/` - Quick lookup info
- Single entry point: `CLAUDE.md` → `.claude/README.md` for navigation

### Benefits
1. **Faster onboarding** - Clear hierarchy, less noise
2. **Better organization** - Logical grouping by purpose
3. **Easier maintenance** - Clear what's current vs archived
4. **Token efficiency** - 65% reduction in documentation size
5. **Clear history** - ADRs preserved, refactoring docs archived

---

## Implementation Steps

1. **Create new directories**
   - `.claude/guides/`
   - `.claude/reference/`
   - `docs/payments/`
   - `docs/archive/`

2. **Move and rename files** (as specified above)

3. **Merge ADR directories**
   - Review for any conflicts
   - Renumber if necessary
   - Delete `.claude/adrs/` after merge

4. **Delete historical files** (33 files)

5. **Update all cross-references**
   - `CLAUDE.md` - Update file paths
   - `.claude/README.md` - Update navigation
   - All files with relative links

6. **Create new README files**
   - `.claude/guides/README.md`
   - `.claude/reference/README.md`
   - `docs/payments/README.md`
   - `docs/archive/README.md`

7. **Test all links** - Verify no broken references

---

## Recommendation

**Proceed with full refactoring.** The benefits (68% file reduction, clearer structure, better UX) far outweigh the effort. This will significantly improve the experience for future Claude Code instances.
