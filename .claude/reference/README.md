# Reference Documentation

> **Advanced guides for debugging, performance, and refactoring**

These files provide detailed information for specific scenarios. Most developers won't need these unless working on advanced topics.

## 📚 What's Here

- **[debugging-guide.md](./debugging-guide.md)** - Advanced debugging with delve, profiling, tracing
- **[performance.md](./performance.md)** - Profiling, benchmarking, optimization techniques
- **[refactoring.md](./refactoring.md)** - Safe refactoring patterns and guidelines
- **[checklist.md](./checklist.md)** - Pre-commit code review checklist
- **[recipes.md](./recipes.md)** - Quick command snippets and copy-paste solutions
- **[prompts.md](./prompts.md)** - Effective prompts for Claude Code

## 🎯 When to Read These

**You probably don't need to read these unless:**
- Debugging complex issues → `debugging-guide.md`
- Optimizing performance → `performance.md`
- Refactoring code → `refactoring.md`
- Looking for quick commands → `recipes.md`

**Instead, start with:**
1. [../architecture.md](../architecture.md) - Clean Architecture and directory layout
2. [../development-guide.md](../development-guide.md) - Setup and commands
3. [../common-tasks.md](../common-tasks.md) - Step-by-step guides
4. [../coding-standards.md](../coding-standards.md) - Go best practices

## 🔍 Quick Search

Looking for something specific? Use grep:

```bash
# Search all reference docs
grep -r "your search term" .claude/reference/

# Search all docs including core
grep -r "your search term" .claude/
```

## 📖 Complete Documentation Map

```
.claude/
├── README.md                    # Start here - entry point
├── architecture.md              # **Core**: Clean Architecture & layout
├── development-guide.md         # **Core**: Setup & commands
├── coding-standards.md          # **Core**: Go conventions
├── common-tasks.md              # **Core**: Step-by-step guides
│
├── glossary.md                  # Business domain template
├── testing.md                   # Testing strategy
├── security.md                  # Security best practices
│
├── adrs/                        # Architecture decisions (8 files)
│
└── reference/                   # **You are here** (7 files)
    ├── debugging-guide.md
    ├── performance.md
    ├── refactoring.md
    ├── checklist.md
    ├── recipes.md
    ├── prompts.md
    └── README.md
```

---

**Note**: These files are advanced reference material. Start with the core 4 files in the parent directory first!
