# Documentation Testing Checklist

> **How to test if this documentation actually helps future Claude Code instances**

## Purpose

This checklist validates that the 37 documentation files we created actually enable Claude Code instances to become productive quickly.

**Use this:**
- After major documentation updates
- Before claiming "documentation is complete"
- When onboarding new team members (test if humans can use it too)
- Every 3-6 months to ensure docs stay relevant

---

## 🧪 Test Setup

### Prerequisites

1. **Fresh Claude Code instance** - New conversation, no context from this session
2. **Clean git state** - No uncommitted changes that would confuse
3. **Working environment** - `make test` and `make dev` should work
4. **Timing tools** - Stopwatch to measure how long tasks take

### How to Run Tests

1. Open a new Claude Code conversation
2. Give Claude the task from test scenario
3. **Do NOT provide extra context** - Let documentation guide Claude
4. Observe behavior and time to completion
5. Record results in table below

---

## 📋 Test Scenarios

### Test 1: Cold Start - Orientation (Target: < 5 minutes)

**Task:**
```
"I'm a new Claude Code instance working on this project.
Please orient yourself to the codebase."
```

**Expected Behavior:**
- [ ] Reads `.claude/CLAUDE-START.md` first
- [ ] Identifies the 3 critical rules (dependencies, business logic location, ops suffix)
- [ ] Reads `.claude/glossary.md` to understand domain
- [ ] Reads `.claude/README.md` or `.claude/cheatsheet.md` for quick reference
- [ ] States it's ready to work (< 5 minutes)

**Success Criteria:**
- ✅ Reads CLAUDE-START.md within first 2 files
- ✅ Can explain what a "loan" vs "subscription" is
- ✅ Knows that business logic goes in domain services
- ✅ Knows use case packages have "ops" suffix
- ✅ Completes orientation in < 5 minutes

**Failure Modes to Watch:**
- ❌ Reads code files before documentation
- ❌ Doesn't read CLAUDE-START.md
- ❌ Takes > 10 minutes
- ❌ Can't explain basic business concepts

**Record Results:**
| Metric | Target | Actual | Pass/Fail |
|--------|--------|--------|-----------|
| Time to complete | < 5 min | ___ min | ___ |
| Read CLAUDE-START.md | Yes | ___ | ___ |
| Understood domain | Yes | ___ | ___ |
| Stated readiness | Yes | ___ | ___ |

---

### Test 2: Simple Task - Add API Endpoint (Target: < 30 minutes)

**Task:**
```
"Add a GET /api/v1/authors/{id} endpoint that returns a single author by ID.
The endpoint should:
- Require authentication
- Return 404 if author not found
- Include Swagger documentation"
```

**Expected Behavior:**
- [ ] Reads `.claude/context-guide.md` to find "Adding an API endpoint" section
- [ ] Follows `.claude/api.md` for API standards
- [ ] Checks `.claude/examples/` or `.claude/codebase-map.md` for similar endpoint
- [ ] Creates handler with proper Swagger annotations including `@Security BearerAuth`
- [ ] Adds route with auth middleware
- [ ] Runs `make gen-docs`
- [ ] Tests the endpoint

**Success Criteria:**
- ✅ Finds correct documentation within 5 minutes
- ✅ Handler has correct Swagger annotations
- ✅ Route uses auth middleware
- ✅ Returns 404 for not found
- ✅ Runs `make gen-docs` to update Swagger
- ✅ Completes in < 30 minutes
- ✅ Code follows project patterns (no architectural violations)

**Failure Modes:**
- ❌ Doesn't add `@Security BearerAuth`
- ❌ Forgets to add auth middleware to route
- ❌ Doesn't regenerate Swagger docs
- ❌ Takes > 45 minutes

**Record Results:**
| Metric | Target | Actual | Pass/Fail |
|--------|--------|--------|-----------|
| Time to complete | < 30 min | ___ min | ___ |
| Correct Swagger | Yes | ___ | ___ |
| Auth middleware | Yes | ___ | ___ |
| Regenerated docs | Yes | ___ | ___ |
| Follows patterns | Yes | ___ | ___ |

---

### Test 3: Complex Task - Add New Domain Feature (Target: < 2 hours)

**Task:**
```
"Add a Reservation feature allowing members to reserve books.

Requirements:
- Members can reserve a book that's currently loaned
- Reservation is held for 48 hours when book becomes available
- Only 1 reservation per member per book
- Basic tier: max 1 reservation, Premium: max 5, VIP: unlimited

Include:
- Domain layer (entity, service, repository interface)
- Use case layer (create reservation, cancel reservation)
- Database migration
- API endpoints (POST /reservations, DELETE /reservations/{id})
- Tests for domain service (100% coverage)
"
```

**Expected Behavior:**
- [ ] Reads `.claude/context-guide.md` → "Adding a New Feature" section
- [ ] Reads `.claude/development-workflows.md` → Complete workflow
- [ ] Reads `.claude/examples/README.md` → Follows pattern
- [ ] Reads `.claude/glossary.md` → Understands reservation business rules
- [ ] Follows 11-phase workflow from development-workflows.md:
  1. Planning (creates feature branch)
  2. Domain layer (entity, service, repository interface)
  3. Use case layer (with "reservationops" package - ops suffix!)
  4. Repository implementation
  5. Database migration (up AND down)
  6. HTTP handlers
  7. Wiring in container.go
  8. Routes with auth
  9. Swagger docs
  10. Tests
  11. Pre-commit checks

**Success Criteria:**
- ✅ Creates `internal/domain/reservation/` directory
- ✅ Business logic in domain service (not use case)
- ✅ Use case package is `reservationops` (NOT `reservation`)
- ✅ Repository interface in domain layer
- ✅ Migration has both up AND down
- ✅ Indexes on foreign keys
- ✅ Swagger annotations include `@Security BearerAuth`
- ✅ Domain tests have 100% coverage
- ✅ Runs `make ci` before suggesting commit
- ✅ Completes in < 2 hours

**Failure Modes:**
- ❌ Use case package named `reservation` (should be `reservationops`)
- ❌ Business logic in use case instead of domain service
- ❌ Migration missing down file
- ❌ No indexes on foreign keys
- ❌ Domain tests < 100% coverage
- ❌ Doesn't run `make ci`

**Record Results:**
| Metric | Target | Actual | Pass/Fail |
|--------|--------|--------|-----------|
| Time to complete | < 2 hours | ___ hrs | ___ |
| Correct package naming | Yes | ___ | ___ |
| Business logic location | Domain | ___ | ___ |
| Migration complete | Yes | ___ | ___ |
| Domain test coverage | 100% | ___% | ___ |
| Ran make ci | Yes | ___ | ___ |

---

### Test 4: Bug Fix (Target: < 30 minutes)

**Task:**
```
"There's a bug: books can be borrowed even when status is 'maintenance'.
Find and fix it."
```

**Expected Behavior:**
- [ ] Reads `.claude/context-guide.md` → "Fixing a Bug" section
- [ ] Reads `.claude/troubleshooting.md` or `.claude/debugging-guide.md`
- [ ] Writes a failing test that reproduces the bug
- [ ] Finds the bug location (likely in CreateLoanUseCase or domain)
- [ ] Fixes the bug
- [ ] Verifies test now passes
- [ ] Runs `make test` to ensure no regressions

**Success Criteria:**
- ✅ Writes failing test first
- ✅ Finds root cause
- ✅ Fixes bug correctly
- ✅ All tests pass
- ✅ Completes in < 30 minutes

**Record Results:**
| Metric | Target | Actual | Pass/Fail |
|--------|--------|--------|-----------|
| Time to complete | < 30 min | ___ min | ___ |
| Test-first approach | Yes | ___ | ___ |
| Bug fixed | Yes | ___ | ___ |
| All tests pass | Yes | ___ | ___ |

---

### Test 5: Security Review (Target: < 15 minutes)

**Task:**
```
"Review the authentication code for security issues."
```

**Expected Behavior:**
- [ ] Reads `.claude/security.md`
- [ ] Checks for hardcoded secrets
- [ ] Checks SQL queries for injection vulnerabilities
- [ ] Checks password hashing (should use bcrypt)
- [ ] Checks JWT token validation
- [ ] Suggests running `.claude/scripts/review.sh`

**Success Criteria:**
- ✅ Identifies if secrets are hardcoded
- ✅ Verifies SQL uses parameterized queries
- ✅ Confirms bcrypt is used for passwords
- ✅ Checks JWT secret is from environment
- ✅ Completes in < 15 minutes

**Record Results:**
| Metric | Target | Actual | Pass/Fail |
|--------|--------|--------|-----------|
| Time to complete | < 15 min | ___ min | ___ |
| Found security issues | All | ___ | ___ |
| Suggested review.sh | Yes | ___ | ___ |

---

### Test 6: Performance Optimization (Target: < 1 hour)

**Task:**
```
"The GET /books endpoint is slow. Investigate and optimize it."
```

**Expected Behavior:**
- [ ] Reads `.claude/context-guide.md` → "Optimizing Performance"
- [ ] Reads `.claude/performance.md` for baseline metrics
- [ ] Checks if indexes exist on frequently queried columns
- [ ] Suggests adding index if missing
- [ ] Creates migration for index
- [ ] Benchmarks before and after

**Success Criteria:**
- ✅ Identifies missing indexes
- ✅ Creates proper migration (up AND down)
- ✅ Measures performance improvement
- ✅ Completes in < 1 hour

**Record Results:**
| Metric | Target | Actual | Pass/Fail |
|--------|--------|--------|-----------|
| Time to complete | < 1 hour | ___ min | ___ |
| Identified issue | Yes | ___ | ___ |
| Created migration | Yes | ___ | ___ |
| Measured improvement | Yes | ___ | ___ |

---

### Test 7: Proactive Improvements (Target: Ongoing)

**Task:**
```
"Add a GetBookByISBN use case."
```

**Expected Behavior:**
- [ ] Completes the primary task correctly
- [ ] Reads `.claude/quick-wins.md`
- [ ] Notices opportunities for improvement (e.g., missing tests, no error wrapping)
- [ ] Asks user: "I noticed X while working on this. Would you like me to fix it?"
- [ ] If user says yes, applies quick win correctly

**Success Criteria:**
- ✅ Completes primary task first
- ✅ Identifies at least 1 quick win opportunity
- ✅ Asks before making extra changes
- ✅ Applies quick win correctly if approved

**Record Results:**
| Metric | Target | Actual | Pass/Fail |
|--------|--------|--------|-----------|
| Primary task complete | Yes | ___ | ___ |
| Quick wins identified | ≥ 1 | ___ | ___ |
| Asked permission | Yes | ___ | ___ |

---

### Test 8: Architectural Decision Understanding (Target: < 10 minutes)

**Task:**
```
"Why do we use the 'ops' suffix for use case packages?"
```

**Expected Behavior:**
- [ ] Reads `.claude/adrs/004-ops-suffix-convention.md`
- [ ] Explains the reason (avoid package naming conflicts)
- [ ] Provides code example showing the problem and solution
- [ ] References alternatives that were considered

**Success Criteria:**
- ✅ Finds correct ADR
- ✅ Explains reason correctly
- ✅ Provides example
- ✅ Completes in < 10 minutes

**Record Results:**
| Metric | Target | Actual | Pass/Fail |
|--------|--------|--------|-----------|
| Time to answer | < 10 min | ___ min | ___ |
| Found ADR | Yes | ___ | ___ |
| Correct explanation | Yes | ___ | ___ |

---

## 📊 Overall Assessment

### Scoring

**Grade each test:**
- **A (Excellent):** All success criteria met, completed within target time
- **B (Good):** Most criteria met, within 1.5x target time
- **C (Acceptable):** Some criteria met, within 2x target time
- **D (Poor):** Few criteria met or took > 2x target time
- **F (Fail):** Didn't complete or major issues

### Results Summary

| Test | Grade | Time | Notes |
|------|-------|------|-------|
| 1. Cold Start | ___ | ___ | ___ |
| 2. API Endpoint | ___ | ___ | ___ |
| 3. New Feature | ___ | ___ | ___ |
| 4. Bug Fix | ___ | ___ | ___ |
| 5. Security Review | ___ | ___ | ___ |
| 6. Performance | ___ | ___ | ___ |
| 7. Proactive | ___ | ___ | ___ |
| 8. ADR Understanding | ___ | ___ | ___ |

**Overall Grade:** ___

---

## 🎯 Success Thresholds

### Excellent Documentation (A grade)
- ≥ 6 tests score A or B
- All critical tests (1, 2, 3) score A or B
- Average time within targets
- No architectural violations
- Claude reads documentation before code

### Good Documentation (B grade)
- ≥ 5 tests score A, B, or C
- Most critical tests score B or better
- Average time within 1.5x targets
- Minor violations that Claude catches

### Needs Improvement (C grade)
- ≥ 4 tests score C or better
- Some critical tests fail
- Average time > 1.5x targets
- Frequent violations

### Documentation Failing (D/F grade)
- < 4 tests pass
- Critical tests fail
- Claude doesn't use documentation
- Frequent architectural violations

---

## 🔍 What to Observe During Testing

### Positive Signals
- ✅ Claude reads CLAUDE-START.md first
- ✅ Claude uses context-guide.md to navigate
- ✅ Claude references specific docs by name
- ✅ Claude follows workflows from development-workflows.md
- ✅ Claude checks ADRs before suggesting changes
- ✅ Claude suggests quick wins appropriately
- ✅ Claude runs automated checks (make ci, review.sh)

### Warning Signals
- ⚠️ Claude reads code files before documentation
- ⚠️ Claude doesn't reference documentation
- ⚠️ Claude asks questions already answered in docs
- ⚠️ Claude suggests changes that violate ADRs
- ⚠️ Claude takes > 2x target time

### Critical Failures
- 🚨 Claude never reads documentation
- 🚨 Claude violates critical architectural rules
- 🚨 Claude hardcodes secrets or has SQL injection
- 🚨 Claude can't complete tasks
- 🚨 Claude produces code that doesn't compile/test

---

## 📝 Documentation Effectiveness Metrics

### Quantitative Metrics

**Time Efficiency:**
```
Baseline (no docs): 2 hours to productivity
Target (with docs): 6 minutes to productivity
Improvement: 20x faster

Measure: Did we achieve < 10 minutes on Test 1?
```

**Task Completion Rate:**
```
Target: 100% of tasks completed correctly
Acceptable: ≥ 75% of tasks completed

Measure: How many tests passed?
```

**Architectural Compliance:**
```
Target: 0 violations
Acceptable: ≤ 2 violations that Claude catches and fixes

Measure: Count violations in Test 3 (complex feature)
```

### Qualitative Metrics

**Documentation Usage:**
- Does Claude cite specific docs?
- Does Claude read docs before code?
- Does Claude use the right doc for each task?

**Code Quality:**
- Follows project patterns?
- Proper error handling?
- Appropriate tests?
- Security best practices?

---

## 🔧 What to Do Based on Results

### If Overall Grade is A
✅ **Documentation is working!**
- Continue using
- Update as codebase evolves
- Consider this validated

### If Overall Grade is B
⚠️ **Minor improvements needed:**
- Review failed tests
- Identify which docs were missed
- Add cross-references
- Clarify confusing sections

### If Overall Grade is C
🔧 **Significant improvements needed:**
- Interview Claude: "What was confusing?"
- Reorganize navigation (context-guide.md)
- Add missing examples
- Consolidate redundant info

### If Overall Grade is D or F
🚨 **Major problems:**
- Documentation may not be discoverable
- CLAUDE-START.md needs to be more prominent
- Context-guide.md may not be clear
- Consider restructuring

---

## 📋 Post-Test Actions

After completing all tests:

1. **Record Results**
   - Fill in all tables above
   - Calculate overall grade
   - Note patterns (which tests consistently pass/fail?)

2. **Identify Gaps**
   ```
   Questions to ask:
   - Which docs were never read?
   - Which docs were read but not helpful?
   - What questions did Claude ask that docs should answer?
   - Where did Claude get stuck?
   ```

3. **Improve Documentation**
   - Fix issues found in testing
   - Add cross-references where navigation failed
   - Clarify confusing sections
   - Remove unused documentation

4. **Retest**
   - After improvements, run critical tests again
   - Verify improvements actually helped
   - Iterate until grade is A or B

---

## 🎓 Human Testing (Bonus)

Want to test if humans can use the docs too?

**Give a new team member this task:**
```
"Read the .claude/ documentation and add a 'Rating' feature
allowing members to rate books 1-5 stars. You have 3 hours."
```

**Observe:**
- Do they find CLAUDE-START.md or README.md first?
- Do they use the documentation or ask questions?
- How long does it take?
- What do they find confusing?

Human feedback is valuable because if humans can't use it, Claude might struggle too.

---

## 📅 Testing Schedule

**Recommended frequency:**

- **After major documentation changes:** Test critical paths (Tests 1, 2, 3)
- **Monthly:** Run full test suite (all 8 tests)
- **After codebase refactoring:** Verify docs are still accurate
- **When onboarding new developers:** Use as validation

---

## ✅ Completion Checklist

After running this testing checklist:

- [ ] All 8 tests completed
- [ ] Results recorded in tables
- [ ] Overall grade calculated
- [ ] Gaps identified
- [ ] Improvement plan created (if needed)
- [ ] Documentation updated based on findings
- [ ] Retested critical failures (if any)

---

**Last Updated:** 2025-01-19

**Next Test Date:** ___________

**Tester:** ___________

**Overall Grade:** ___________

**Key Findings:** ___________
