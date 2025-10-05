# Claude Code Guide

> **Quick navigation for AI-assisted development in this repository**

## Project Overview

Library Management System - A Go REST API service built with Clean Architecture principles. The project is optimized for vibecoding with fast feedback loops and clear separation of concerns.

**Module Name:** `library-service`
**Go Version:** 1.25.0
**Architecture:** Clean Architecture (Onion/Hexagonal)

## 📚 Documentation Index

### 🚨 **NEW CLAUDE CODE INSTANCE? [START HERE →](./CLAUDE-START.md)**

### 🎯 Essential Reading (First 3 Minutes)
- **[CLAUDE-START.md](./CLAUDE-START.md)** - Boot sequence for new AI instances ⭐ NEW
- **[Context Guide](./context-guide.md)** - What to read for your specific task ⭐ NEW
- **[Glossary](./glossary.md)** - Business domain terms and concepts ⭐ NEW
- **[Codebase Map](./codebase-map.md)** - Find any file in 10 seconds ⭐ NEW
- **[Cheat Sheet](./cheatsheet.md)** - Single-page quick reference

### Active Documentation
- **[Commands](./commands.md)** - All essential commands and quick reference
- **[Setup](./setup.md)** - First-time setup and troubleshooting
- **[Architecture](./architecture.md)** - Clean architecture patterns and design
- **[Development](./development.md)** - Development workflow and productivity tips
- **[Testing](./testing.md)** - Testing patterns and strategies
- **[API](./api.md)** - API endpoints and design standards
- **[Standards](./standards.md)** - Code standards and conventions

### Practical Guides
- **[Quick Wins](./quick-wins.md)** - Safe improvements Claude can suggest ⭐ NEW
- **[Development Workflows](./development-workflows.md)** - Complete workflows start to finish
- **[Debugging Guide](./debugging-guide.md)** - Advanced debugging techniques
- **[Performance](./performance.md)** - Baselines, profiling, optimization
- **[Onboarding](./onboarding.md)** - 15-minute quick start guide
- **[Visual Flows](./flows.md)** - ASCII diagrams of request flows
- **[Gotchas](./gotchas.md)** - Common mistakes to avoid
- **[Examples](./examples/)** - Copy-paste code examples for common tasks
- **[Recipes](./recipes.md)** - Quick solutions and command snippets
- **[Troubleshooting](./troubleshooting.md)** - Solutions to common problems
- **[FAQ](./faq.md)** - Frequently asked questions
- **[Checklist](./checklist.md)** - Pre-commit code review checklist
- **[Prompts](./prompts.md)** - Effective Claude Code prompts

### Reference Material
- **[Architecture Decision Records (ADRs)](./adrs/)** - Why key decisions were made
- **[Security Guide](./security.md)** - Security best practices and common vulnerabilities
- **[Testing Checklist](./TESTING-CHECKLIST.md)** - Validate documentation effectiveness ⭐ NEW
- **[Refactoring Guide](./refactoring.md)** - Generic Go Clean Architecture refactoring template

## ⚡ 30-Second Quick Reference

**I just want to...**

| Task | Command |
|------|---------|
| **Start everything** | `make dev` |
| **Run API only** | `go run ./cmd/api` |
| **Run tests** | `make test` |
| **Fix my code** | `make fmt && make lint` |
| **Reset database** | `make down && make up && make migrate-up` |
| **Before committing** | `make ci` |

**Emergency Fixes:**
```bash
# Nothing works after git pull
make down && make up && make migrate-up && go mod tidy && make dev

# Port 8080 already in use
lsof -ti:8080 | xargs kill -9

# Tests fail randomly
go clean -testcache && make test

# Linter errors
make fmt && go mod tidy
```

**File Locations (Quick Lookup):**
```
Add business logic?     → internal/domain/{entity}/service.go
Add API endpoint?       → internal/adapters/http/handlers/{entity}.go
Add use case?           → internal/usecase/{entity}/{operation}.go
Wire dependencies?      → internal/infrastructure/app/app.go
Add migration?          → make migrate-create name=your_migration_name
Add test?               → {file}_test.go (same directory as file being tested)
```

## 🏗️ Architecture at a Glance

```
internal/
├── domain/              # Business logic (ZERO external dependencies)
│   ├── book/           # Book entity, service, repository interface
│   ├── member/         # Member entity, service (subscriptions)
│   └── author/         # Author entity
├── usecase/            # Application orchestration
│   ├── book/           # CreateBook, UpdateBook, etc.
│   ├── auth/           # Register, Login, RefreshToken
│   └── subscription/   # SubscribeMember
├── adapters/           # External interfaces (HTTP, DB, cache)
│   ├── http/           # Handlers, middleware, DTOs
│   ├── repository/     # PostgreSQL/MongoDB/Memory
│   └── cache/          # Redis/Memory cache
└── infrastructure/     # Technical concerns
    ├── auth/           # JWT token management
    ├── store/          # Database connections
    └── server/         # HTTP server configuration
```

**Critical Rule:** Domain → Use Case → Adapters → Infrastructure (dependencies point inward only)

## 🚀 Getting Started

**First time setup:**
```bash
make init && make up && make migrate-up && make run
```

**Daily development:**
```bash
make dev                # Start everything
make test               # Run tests
make ci                 # Before commit
```

For detailed setup instructions, see [Setup Guide](./setup.md).
