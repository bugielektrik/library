# Code Examples - Library Management System

Practical, step-by-step guides for common development tasks in this Clean Architecture project.

**Target Audience:** Developers working with this Go-based library management system

**Philosophy:** Learn by doing - each guide provides complete, working examples you can follow immediately.

---

## 📚 Available Guides

### 1. [Adding a Domain Entity](./adding-domain-entity.md)

**What you'll learn:** How to add a complete new domain entity from scratch

**Example:** Adding a "Review" domain (book reviews with ratings)

**Covers:**
- Domain layer: Entity, service, repository interface
- Use case layer: CRUD operations with business logic
- Adapter layer: PostgreSQL repository, HTTP handlers, DTOs
- Database migrations
- Dependency injection wiring
- Unit tests with mocks
- Integration tests

**Time Estimate:** 2-3 hours

**Skill Level:** ⭐⭐ Intermediate

**When to use:**
- Adding a new core business concept (e.g., Loan, Fine, Notification)
- Need full CRUD + custom business logic
- Want to understand the complete Clean Architecture flow

---

### 2. [Adding an API Endpoint](./adding-api-endpoint.md)

**What you'll learn:** How to quickly add a new endpoint to an existing domain

**Example:** Adding `GET /books/{id}/availability` to check book availability

**Covers:**
- Creating a use case for the endpoint
- Adding DTO for HTTP response
- Implementing HTTP handler with Swagger annotations
- Registering route in router
- Wiring use case in container
- Testing with curl
- Common endpoint patterns (query params, path params, authentication)

**Time Estimate:** 30-45 minutes

**Skill Level:** ⭐ Beginner/Intermediate

**When to use:**
- Adding functionality to existing domains
- Quick API additions
- Learning handler patterns

---

### 3. [Integration Testing](./integration-testing.md)

**What you'll learn:** How to write integration tests with real PostgreSQL database

**Covers:**
- Test database setup and configuration
- Reusable test fixtures
- Repository integration tests
- Use case integration tests
- HTTP handler integration tests
- Test isolation strategies (truncate tables, transactions)
- Running tests in parallel
- CI/CD integration with GitHub Actions

**Time Estimate:** 30 minutes setup + 15 minutes per test suite

**Skill Level:** ⭐⭐ Intermediate

**When to use:**
- Testing complex SQL queries
- Verifying multi-table operations
- Testing transaction handling
- Ensuring repository implementations work correctly

---

### 4. [Common Tasks - Quick Reference](./common-tasks.md)

**What you'll learn:** Quick answers to frequently asked questions

**Format:** FAQ-style with immediate answers and commands

**Covers:**
- Database & migrations
- Testing (unit, integration, coverage)
- Code quality (formatting, linting, CI)
- Building & running
- API development
- Error handling
- Logging
- Dependency injection
- Architecture questions
- Docker & deployment
- Git & version control
- Troubleshooting
- Performance

**Time Estimate:** 2-5 minutes per task

**Skill Level:** ⭐ All levels

**When to use:**
- Need a quick command or solution
- First time setting up project
- Debugging common issues
- Can't remember exact syntax

---

## 🎯 Quick Navigation

**I want to...**

- **Add a new domain concept (e.g., Loan, Fine)** → [Adding a Domain Entity](./adding-domain-entity.md)
- **Add an endpoint to existing domain** → [Adding an API Endpoint](./adding-api-endpoint.md)
- **Write tests with real database** → [Integration Testing](./integration-testing.md)
- **Find a specific command** → [Common Tasks](./common-tasks.md)
- **Understand the architecture** → [.claude/architecture.md](../.claude/architecture.md)
- **Set up the project** → [.claude/setup.md](../.claude/setup.md)

---

## 📖 Learning Path

### Beginner (New to this codebase)

1. **Setup:** Read [.claude/setup.md](../.claude/setup.md) - Set up development environment
2. **Quick Reference:** Skim [Common Tasks](./common-tasks.md) - Familiarize with commands
3. **First Feature:** Follow [Adding an API Endpoint](./adding-api-endpoint.md) - Learn basic flow
4. **Architecture:** Read [.claude/architecture.md](../.claude/architecture.md) - Understand structure

### Intermediate (Comfortable with basics)

1. **Full Feature:** Follow [Adding a Domain Entity](./adding-domain-entity.md) - Complete workflow
2. **Testing:** Follow [Integration Testing](./integration-testing.md) - Write integration tests
3. **Standards:** Read [.claude/standards.md](../.claude/standards.md) - Code quality practices
4. **Workflows:** Read [.claude/development-workflows.md](../.claude/development-workflows.md) - Advanced patterns

### Advanced (Contributing significant features)

1. **Architecture Deep Dive:** Study [.claude/architecture.md](../.claude/architecture.md) + [.claude/development.md](../.claude/development.md)
2. **All Examples:** Work through all examples in this directory
3. **Testing Strategy:** Read [.claude/testing.md](../.claude/testing.md) - Comprehensive testing
4. **Debugging:** Read [.claude/debugging-guide.md](../.claude/debugging-guide.md) - Advanced techniques

---

## 🛠️ Example Code Quality

All code examples in these guides:

- ✅ Follow project coding standards
- ✅ Use consistent naming conventions
- ✅ Include error handling
- ✅ Have logging instrumentation
- ✅ Include Swagger annotations
- ✅ Follow Clean Architecture principles
- ✅ Are production-ready (with minor customization)

**Note:** Examples use placeholder names (e.g., "Review", "Availability"). Replace with your actual feature names.

---

## 📋 Quick Command Cheatsheet

```bash
# Start developing
make dev                           # Full stack (Docker + migrations + API)

# Testing
make test                          # All tests with coverage
make test-unit                     # Unit tests only (fast)
make test-integration              # Integration tests (requires DB)

# Code quality
make ci                            # Full CI pipeline locally
make lint                          # Run linter
make fmt                           # Format code

# Database
make migrate-up                    # Apply migrations
make migrate-down                  # Rollback last migration
make migrate-create name=my_migration  # Create new migration

# API
make run                           # Run API server
make gen-docs                      # Regenerate Swagger docs

# Building
make build                         # Build all binaries
make build-api                     # Build API server only
```

---

## 🎓 Additional Resources

### Project Documentation (`.claude/` directory)

**Essential:**
- [CLAUDE-START.md](../.claude/CLAUDE-START.md) - 60-second quickstart
- [README.md](../.claude/README.md) - Quick navigation
- [architecture.md](../.claude/architecture.md) - Clean Architecture patterns
- [standards.md](../.claude/standards.md) - Code standards

**Development:**
- [development.md](../.claude/development.md) - Development practices
- [development-workflows.md](../.claude/development-workflows.md) - Complete workflows
- [testing.md](../.claude/testing.md) - Testing strategies
- [debugging-guide.md](../.claude/debugging-guide.md) - Debugging techniques

**Reference:**
- [api.md](../.claude/api.md) - API documentation
- [commands.md](../.claude/commands.md) - Command reference
- [cheatsheet.md](../.claude/cheatsheet.md) - Single-page reference
- [faq.md](../.claude/faq.md) - Frequently asked questions

**Problem Solving:**
- [troubleshooting.md](../.claude/troubleshooting.md) - Common issues
- [gotchas.md](../.claude/gotchas.md) - Common mistakes
- [quick-wins.md](../.claude/quick-wins.md) - Easy improvements

### External Resources

- [Clean Architecture (Uncle Bob)](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Go Documentation](https://go.dev/doc/)
- [Chi Router](https://go-chi.io/)
- [Swagger/OpenAPI](https://swagger.io/specification/)
- [Go Validator](https://pkg.go.dev/github.com/go-playground/validator/v10)

---

## 💡 Tips for Using Examples

1. **Don't blindly copy-paste** - Understand each step and adapt to your needs
2. **Follow the order** - Examples build on each other (domain → use case → handler)
3. **Run tests frequently** - Verify each step works before moving forward
4. **Use version control** - Commit after each major step
5. **Read comments** - Code examples include explanatory comments
6. **Check time estimates** - Plan your work session accordingly
7. **Refer to Common Tasks** - Quick answers while following examples

---

## 🤝 Contributing Examples

Have an example that would help other developers? Consider adding it!

**Good example topics:**
- Adding middleware
- Implementing caching
- Adding background jobs
- Working with transactions
- Error handling patterns
- Authentication flows
- File uploads
- WebSocket endpoints
- Rate limiting

**Example structure:**
```markdown
# Title - Quick Guide

Brief description

**Scenario:** Specific example
**Time Estimate:** X minutes

## Step 1: ...
## Step 2: ...
## Summary Checklist
```

---

## 📞 Getting Help

**Stuck? Try these:**

1. Check [Common Tasks](./common-tasks.md) - Covers 90% of questions
2. Review [Troubleshooting](../.claude/troubleshooting.md) - Common issues
3. Read [FAQ](../.claude/faq.md) - Frequently asked questions
4. Check [Gotchas](../.claude/gotchas.md) - Common mistakes
5. Review example code in `internal/` - Working examples
6. Run `make ci` - Verify your changes compile

**Still stuck?**
- Verify Docker containers are running: `docker ps`
- Check logs: `make logs`
- Reset database: `make migrate-down && make migrate-up`
- Clear test cache: `go clean -testcache`

---

## 📊 Progress Tracking

Use this checklist when following examples:

```
Adding a Domain Entity:
□ Step 1: Domain layer (entity, service, repository)
□ Step 2: Use case layer (CRUD operations)
□ Step 3: Repository implementation
□ Step 4: Database migration
□ Step 5: HTTP layer (handlers, DTOs)
□ Step 6: Dependency wiring
□ Step 7: Testing

Adding an API Endpoint:
□ Step 1: Create use case
□ Step 2: Create DTO
□ Step 3: Add HTTP handler
□ Step 4: Add route
□ Step 5: Wire in container
□ Step 6: Test endpoint
□ Step 7: Add tests (optional)

Integration Testing:
□ Step 1: Database setup (one-time)
□ Step 2: Test fixtures (one-time)
□ Step 3: Write repository tests
□ Step 4: Write use case tests
□ Step 5: Write handler tests
□ Step 6: Run integration tests
```

---

**Happy Coding! 🚀**

These examples are designed to get you productive quickly. Start with [Common Tasks](./common-tasks.md) for quick wins, then dive into [Adding an API Endpoint](./adding-api-endpoint.md) for your first feature.
