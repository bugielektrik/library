# Claude Context Directory

**Purpose:** Persistent context files that reduce cold-start token requirements for Claude Code sessions from 5,000-8,000 tokens to 1,500-2,500 tokens.

---

## Files in This Directory

### 1. `SESSION_MEMORY.md` (~1,200 tokens)

**Purpose:** Accumulated architectural knowledge across sessions

**Contents:**
- Current architecture decisions
- Recent refactoring history
- Known patterns and conventions
- Common pitfalls
- Token optimization notes

**When to Load:** Start of every Claude Code session

**When to Update:**
- After major architectural changes
- After significant refactoring
- When discovering new patterns
- When adding/changing conventions

### 2. `CURRENT_PATTERNS.md` (~1,500 tokens)

**Purpose:** Quick reference for active code patterns

**Contents:**
- File organization structure
- Code templates (handler, use case, repository, test)
- Common operations workflows
- Validation patterns
- Error handling patterns
- Logging patterns

**When to Load:** When creating new code or following existing patterns

**When to Update:**
- When patterns change
- When adding new patterns
- After team code review sessions
- When standardizing approaches

### 3. `TOKEN_LOG.md` (~800 tokens)

**Purpose:** Track token consumption to measure optimization impact

**Contents:**
- Baseline measurements (pre-optimization)
- Post-optimization targets
- Recent tasks log
- Token efficiency metrics
- Optimization wins tracking

**When to Load:** When reviewing session efficiency or planning optimizations

**When to Update:**
- After completing significant tasks
- Weekly for pattern review
- Monthly for comprehensive analysis
- After implementing optimizations

---

## Usage Guidelines

### For Claude Code

**Session Start Protocol:**
```
1. Load SESSION_MEMORY.md (1,200 tokens)
2. If creating code: Load CURRENT_PATTERNS.md (1,500 tokens)
3. Reference examples/ directory as needed (600-800 tokens per example)

Total cold start: 1,500-2,500 tokens
vs. Previous: 5,000-8,000 tokens
Savings: 60-70%
```

**During Development:**
```
- Reference CURRENT_PATTERNS.md for templates
- Check examples/ for detailed implementations
- Update TOKEN_LOG.md after major tasks
```

**Session End:**
```
- Update SESSION_MEMORY.md if architecture changed
- Log token consumption in TOKEN_LOG.md
- Note any new patterns discovered
```

### For Developers

**Weekly Maintenance:**
- Review TOKEN_LOG.md for trends
- Update CURRENT_PATTERNS.md if conventions changed
- Ensure SESSION_MEMORY.md reflects current state

**Monthly Review:**
- Calculate average tokens per task type
- Identify inefficiencies
- Plan optimizations
- Update targets in TOKEN_LOG.md

**After Major Refactoring:**
- Update all three files
- Document changes and rationale
- Update token impact estimates

---

## Token Optimization Strategy

### Three-Tier Context Loading

**Tier 1: Always Load (1,500-2,500 tokens)**
- SESSION_MEMORY.md
- CURRENT_PATTERNS.md (if creating code)
- Root CLAUDE.md (high-level overview)

**Tier 2: Load as Needed (600-1,200 tokens)**
- Relevant example from examples/ directory
- Specific .claude/ documentation

**Tier 3: Load for Deep Dives (2,000-5,000 tokens)**
- Actual codebase files
- Multiple related files
- Integration test setups

### Context Selection Rules

**Use Tier 1 when:**
- Starting new session
- Understanding project structure
- Following established patterns

**Use Tier 2 when:**
- Creating new components
- Following specific patterns
- Need implementation details

**Use Tier 3 when:**
- Debugging complex issues
- Refactoring existing code
- Understanding intricate dependencies

---

## Maintenance Schedule

### Daily
- ✅ Update TOKEN_LOG.md for significant tasks
- ✅ Note pattern violations or inefficiencies

### Weekly
- ✅ Review TOKEN_LOG.md trends
- ✅ Update CURRENT_PATTERNS.md if needed
- ✅ Check SESSION_MEMORY.md accuracy

### Monthly
- ✅ Comprehensive TOKEN_LOG.md analysis
- ✅ Calculate optimization impact
- ✅ Update targets and goals
- ✅ Clean up stale information

### Quarterly
- ✅ Major SESSION_MEMORY.md review
- ✅ Architecture alignment check
- ✅ Pattern consistency audit
- ✅ Token efficiency planning

---

## File Size Limits

**Critical for Token Efficiency:**

| File | Current | Maximum | Tokens |
|------|---------|---------|--------|
| SESSION_MEMORY.md | 600 lines | 1,000 lines | 1,200-2,000 |
| CURRENT_PATTERNS.md | 750 lines | 1,000 lines | 1,500-2,000 |
| TOKEN_LOG.md | 400 lines | 800 lines | 800-1,600 |
| **Total** | **1,750 lines** | **2,800 lines** | **3,500-5,600** |

**Why Limits Matter:**
- Keep context focused and relevant
- Faster loading and processing
- Easier to maintain
- Better signal-to-noise ratio

**When Approaching Limits:**
- Archive old information to `archive/` subdirectory
- Remove stale patterns
- Consolidate redundant information
- Split into more specific files if needed

---

## Integration with Other Documentation

### Documentation Hierarchy

```
Root Level (Quick Start)
├── CLAUDE.md                    # 5-minute project overview
├── README.md                    # User-facing documentation
└── refactoring.txt             # Token efficiency article

Context Layer (Session Start - THIS DIRECTORY)
├── .claude-context/
│   ├── SESSION_MEMORY.md       # Architecture context
│   ├── CURRENT_PATTERNS.md     # Active patterns
│   └── TOKEN_LOG.md            # Efficiency tracking

Examples Layer (Pattern Reference)
└── examples/
    ├── handler_example.go      # Handler template
    ├── usecase_example.go      # Use case template
    ├── repository_example.go   # Repository template
    └── test_example_test.go    # Test template

Documentation Layer (Deep Dive)
└── .claude/
    ├── architecture.md         # Detailed architecture
    ├── development-guide.md    # Development workflows
    ├── testing.md              # Testing strategies
    └── adr/                    # Architecture decisions
```

### When to Use Each Level

**Quick Start (1-2 minutes):**
- Read CLAUDE.md only
- Understand project at high level
- Know where to find details

**Session Start (3-5 minutes):**
- Read SESSION_MEMORY.md
- Skim CURRENT_PATTERNS.md
- Ready to code with context

**Deep Work (10-20 minutes):**
- Reference examples/ for patterns
- Read specific .claude/ docs
- Review actual codebase

**Expert Level (1-2 hours):**
- Study architecture.md
- Review all ADRs
- Understand system holistically

---

## Expected Impact

### Before Context Files

**Session Start:**
- Read CLAUDE.md: 2,100 tokens
- Infer architecture from code: 3,000-5,000 tokens
- Find pattern examples: 2,000-3,000 tokens
- **Total:** 7,100-10,100 tokens

**Per Task:**
- Search for patterns: 1,500-2,500 tokens
- Load related files: 2,500-4,000 tokens
- Understand dependencies: 1,000-2,000 tokens
- **Total:** 5,000-8,500 tokens per task

### After Context Files

**Session Start:**
- Read CLAUDE.md: 2,100 tokens
- Load SESSION_MEMORY.md: 1,200 tokens
- Load CURRENT_PATTERNS.md: 1,500 tokens
- **Total:** 4,800 tokens (52% reduction)

**Per Task:**
- Reference CURRENT_PATTERNS.md: 0 tokens (already loaded)
- Load 1 example: 600-800 tokens
- Load specific files: 1,500-3,000 tokens
- **Total:** 2,100-3,800 tokens per task (60% reduction)

### ROI Calculation

**Setup Investment:** 4-6 hours
**Maintenance:** 1-2 hours/month
**Savings per Task:** 2,500-4,700 tokens
**Break-even:** After 10-15 tasks
**Annual Savings:** 100,000-200,000 tokens (for active projects)

---

## Anti-Patterns to Avoid

❌ **Don't duplicate CLAUDE.md content**
- These files are SESSION context, not PROJECT overview
- Keep high-level overview in CLAUDE.md
- Keep session-specific knowledge here

❌ **Don't let files grow unbounded**
- Enforce size limits strictly
- Archive old information
- Remove redundant content

❌ **Don't include implementation details**
- Patterns and templates only
- Point to examples/ for full code
- Keep focused on context, not code

❌ **Don't forget to update**
- Stale context is worse than no context
- Update after every major change
- Review regularly

❌ **Don't over-optimize early**
- Start simple, measure impact
- Add complexity only if needed
- Focus on most common tasks first

---

**Created:** October 11, 2025
**Purpose:** Token-efficient AI coding assistance
**Maintenance:** Ongoing
