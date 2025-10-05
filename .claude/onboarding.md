# Onboarding Checklist

> **First 15 minutes with this codebase - a guided path to productivity**

## ⚡ Quick Start (5 minutes)

### 1. Read Core Documentation

- [ ] Read [CLAUDE.md](../CLAUDE.md) - Main guidance document
- [ ] Skim [README.md](../README.md) - Project overview
- [ ] Review [.claude/README.md](./README.md) - Quick reference

**Key Concepts to Understand:**
- Clean Architecture (Domain → Use Case → Adapters → Infrastructure)
- "ops" suffix convention (bookops, authops, subops)
- Two-step dependency injection (app.go → container.go)

### 2. Verify Environment

```bash
# Check tools
command -v go >/dev/null 2>&1 && echo "✓ Go" || echo "✗ Go missing"
command -v docker >/dev/null 2>&1 && echo "✓ Docker" || echo "✗ Docker missing"
command -v make >/dev/null 2>&1 && echo "✓ Make" || echo "✗ Make missing"

# Check Go version (need 1.25+)
go version
```

### 3. Start the Project

```bash
# One-time setup
make init          # Download dependencies
make up            # Start PostgreSQL + Redis
make migrate-up    # Run migrations

# Start API
make run           # or `make dev` for full stack
```

### 4. Verify Everything Works

```bash
# In another terminal:
# Check health
curl http://localhost:8080/health

# Register a user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test123!@#","full_name":"Test User"}'

# Run tests
make test
```

**If anything fails:** See [troubleshooting.md](./troubleshooting.md)

## 📚 Understand the Architecture (10 minutes)

### 1. Explore the Directory Structure

```bash
# Domain layer (business logic)
ls -la internal/domain/
ls -la internal/domain/book/

# Use case layer (orchestration)
ls -la internal/usecase/
ls -la internal/usecase/bookops/

# Adapters (external interfaces)
ls -la internal/adapters/http/handlers/
ls -la internal/adapters/repository/postgres/

# Infrastructure (technical concerns)
ls -la internal/infrastructure/
```

### 2. Trace a Request Through the System

Follow the flow of **"Create a Book"**:

```bash
# 1. HTTP Request arrives
cat internal/adapters/http/handlers/book.go | grep -A 20 "func.*CreateBook"

# 2. Handler calls Use Case
cat internal/usecase/bookops/create_book.go | grep -A 30 "func.*Execute"

# 3. Use Case calls Domain Service
cat internal/domain/book/service.go | grep -A 10 "func.*ValidateISBN"

# 4. Use Case persists via Repository
cat internal/adapters/repository/postgres/book.go | grep -A 10 "func.*Create"
```

### 3. Understand Dependency Injection

```bash
# Step 1: Bootstrap (app.go)
cat internal/infrastructure/app/app.go | grep -A 50 "func New"

# Step 2: Container (container.go)
cat internal/usecase/container.go | grep -A 30 "func NewContainer"
```

### 4. Check Current State

```bash
# See what use cases exist
grep -r "type.*UseCase struct" internal/usecase/*/

# See what domains exist
ls -la internal/domain/

# See what handlers exist
ls -la internal/adapters/http/handlers/
```

## 🎯 Know Your Tools (Quick Reference)

### Essential Commands

```bash
# Development
make dev           # Start everything
make run           # Run API only
make test          # Run all tests
make ci            # Before commit (fmt, vet, lint, test, build)

# Database
make migrate-up    # Apply migrations
make migrate-down  # Rollback migration
make migrate-create name=your_migration  # Create new migration

# Code Quality
make fmt           # Format code
make lint          # Run linter
make vet           # Run go vet

# Docker
make up            # Start services
make down          # Stop services
```

### Key Files to Know

| Purpose | File Path |
|---------|-----------|
| Wire new use cases | `internal/usecase/container.go` |
| Bootstrap app | `internal/infrastructure/app/app.go` |
| Add HTTP routes | `internal/adapters/http/router.go` |
| Main entry point | `cmd/api/main.go` |
| Swagger metadata | `cmd/api/main.go` (annotations) |
| Makefile | `Makefile` |
| Environment vars | `.env` |
| Docker setup | `deployments/docker/docker-compose.yml` |

## 🔨 Practice Task (5-10 minutes)

Try adding a simple feature to solidify your understanding:

### Task: Add "Get All Authors" Endpoint

**Expected time:** 5-10 minutes

#### Step 1: Domain (already exists)

```bash
# Author domain already exists
cat internal/domain/author/repository.go | grep "List"
```

#### Step 2: Use Case

```bash
# Create new use case
cat > internal/usecase/authorops/list_authors.go << 'EOF'
package authorops

import (
    "context"
    "library-service/internal/domain/author"
)

type ListAuthorsUseCase struct {
    repo author.Repository
}

func NewListAuthorsUseCase(repo author.Repository) *ListAuthorsUseCase {
    return &ListAuthorsUseCase{repo: repo}
}

func (uc *ListAuthorsUseCase) Execute(ctx context.Context) ([]author.Entity, error) {
    return uc.repo.List(ctx)
}
EOF
```

#### Step 3: Wire in Container

```go
// Edit internal/usecase/container.go
// Add to Container struct:
ListAuthors *authorops.ListAuthorsUseCase

// Add to NewContainer return:
ListAuthors: authorops.NewListAuthorsUseCase(repos.Author),
```

#### Step 4: Add HTTP Handler

```go
// Edit internal/adapters/http/handlers/author.go (if exists)
// Or create it

// @Summary List all authors
// @Tags authors
// @Produce json
// @Security BearerAuth
// @Success 200 {array} dto.AuthorResponse
// @Router /authors [get]
func (h *AuthorHandler) ListAuthors(w http.ResponseWriter, r *http.Request) {
    authors, err := h.listAuthorsUC.Execute(r.Context())
    if err != nil {
        respondError(w, err, http.StatusInternalServerError)
        return
    }
    respondJSON(w, authors, http.StatusOK)
}
```

#### Step 5: Add Route

```go
// Edit internal/adapters/http/router.go
r.Route("/authors", func(r chi.Router) {
    r.Use(authMiddleware)
    r.Get("/", handlers.Author.ListAuthors)
})
```

#### Step 6: Test

```bash
# Regenerate Swagger
make gen-docs

# Restart API
make run

# Test endpoint
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test123!@#"}' \
  | jq -r '.tokens.access_token')

curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/authors | jq
```

**Did it work?** ✓ You understand the architecture!
**Didn't work?** Check [troubleshooting.md](./troubleshooting.md)

## 📖 Deep Dive Resources

Now that you have the basics, explore these for deeper understanding:

### Architecture & Patterns

- [ ] [architecture.md](./architecture.md) - Detailed architecture guide
- [ ] [standards.md](./standards.md) - Code standards and conventions
- [ ] [examples/](./examples/) - Code examples and patterns

### Development Workflow

- [ ] [development.md](./development.md) - Daily development workflow
- [ ] [testing.md](./testing.md) - Testing strategies
- [ ] [recipes.md](./recipes.md) - Copy-paste solutions

### API & Integration

- [ ] [api.md](./api.md) - API design and endpoints
- [ ] Swagger UI: http://localhost:8080/swagger/index.html

## 🎓 Common Patterns to Know

### 1. Adding a New Domain

```
1. Create internal/domain/entity/
2. Add entity.go, service.go, repository.go
3. Write tests (100% coverage for domain)
4. Create use cases in internal/usecase/entityops/
5. Add repository implementation in adapters/repository/postgres/
6. Add HTTP handler in adapters/http/handlers/
7. Wire in container.go
8. Add routes in router.go
9. Regenerate Swagger (make gen-docs)
```

**Reference:** [examples/README.md](./examples/README.md)

### 2. Package Naming Convention

```
Domain:   internal/domain/book      → package book
Use Case: internal/usecase/bookops  → package bookops ("ops" suffix!)
```

**Why?** Avoids import conflicts and removes need for aliases.

### 3. Error Handling

```go
// Always wrap errors with context
if err := repo.Create(ctx, entity); err != nil {
    return fmt.Errorf("creating entity: %w", err)
}
```

### 4. Testing Pattern

```go
// Table-driven tests are standard
func TestValidate(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"valid", "valid-input", false},
        {"invalid", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test code
        })
    }
}
```

## ⚡ Quick Wins Checklist

Things you can do immediately to be helpful:

- [ ] Run `make ci` before any commit
- [ ] Add Swagger annotations to new endpoints
- [ ] Write tests for new code (especially domain layer)
- [ ] Follow package naming convention ("ops" suffix)
- [ ] Use structured logging: `log.Info("message", "key", value)`
- [ ] Wrap errors with context: `fmt.Errorf("operation: %w", err)`
- [ ] Check [recipes.md](./recipes.md) before writing boilerplate

## 🚨 Things to Avoid

Common pitfalls for new Claude instances:

❌ **DON'T** import from outer layers in domain
✅ **DO** keep domain pure (no external dependencies)

❌ **DON'T** use package name `book` for use cases
✅ **DO** use `bookops` suffix to avoid conflicts

❌ **DON'T** skip tests (especially domain layer)
✅ **DO** aim for 100% coverage in domain, 80%+ in use cases

❌ **DON'T** put business logic in handlers
✅ **DO** keep handlers thin, delegate to use cases

❌ **DON'T** skip Swagger annotations
✅ **DO** document all endpoints with `@Security BearerAuth` for protected routes

❌ **DON'T** commit without running `make ci`
✅ **DO** run full CI locally before committing

## 📝 Productivity Tips

### Use Aliases

```bash
# Add to ~/.bashrc or ~/.zshrc
alias ldev='cd ~/path/to/library && make dev'
alias ltest='make test'
alias lci='make ci'
alias lauth='export TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login -H "Content-Type: application/json" -d "{\"email\":\"test@example.com\",\"password\":\"Test123!@#\"}" | jq -r ".tokens.access_token")'
```

### Keep This Open

- Swagger UI: http://localhost:8080/swagger/index.html
- [recipes.md](./recipes.md) - for quick copy-paste
- [troubleshooting.md](./troubleshooting.md) - when things break

### Hot Reload (Optional)

```bash
# Install Air
go install github.com/cosmtrek/air@latest

# Run with hot reload
air
```

## ✅ Onboarding Complete!

You're ready when you can:

- [ ] Explain the 4 Clean Architecture layers
- [ ] Add a new API endpoint
- [ ] Write a table-driven test
- [ ] Run the full CI pipeline locally
- [ ] Debug a failing test
- [ ] Use Swagger UI to test endpoints

**Next steps:**
1. Pick a task from the backlog
2. Read relevant documentation in `.claude/`
3. Use [recipes.md](./recipes.md) for common patterns
4. Ask for help when stuck (provide full error messages!)

**Welcome to the Library Management System! 🎉**
