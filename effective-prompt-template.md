# Effective Prompt Template for Code Refactoring

## Task: Refactor Project to Google Go Style Guide

### 📚 Reference Documentation
- **Primary**: https://google.github.io/styleguide/go/best-practices
- **Secondary**: https://go.dev/doc/effective_go

### 🎯 Objective
Refactor the entire codebase to strictly follow Google's Go style guide and best practices.

### 📋 Refactoring Checklist

#### 1. Code Structure & Organization
- [ ] Package names: lowercase, single word, no underscores
- [ ] One package per directory
- [ ] Imports grouped: standard → external → internal
- [ ] Consistent file naming: `snake_case.go`

#### 2. Naming Conventions
- [ ] **Exported**: `CamelCase` for public APIs
- [ ] **Unexported**: `camelCase` for internal use
- [ ] **Interfaces**: verb + "er" suffix (e.g., `Reader`, `Writer`)
- [ ] **Constants**: `CamelCase` or `SCREAMING_SNAKE_CASE` for exported
- [ ] **Acronyms**: Keep consistent case (e.g., `URLParser`, not `UrlParser`)

#### 3. Function & Method Design
- [ ] Functions do one thing well
- [ ] Early returns for error cases
- [ ] Named return values only when they clarify
- [ ] Receiver names: 1-2 letters, consistent across methods
- [ ] Prefer value receivers unless mutation needed

#### 4. Error Handling
```go
// ✅ Good
if err != nil {
    return fmt.Errorf("processing user %d: %w", userID, err)
}

// ❌ Bad
if err != nil {
    return err
}
```
- [ ] Check errors immediately
- [ ] Add context with `fmt.Errorf` and `%w`
- [ ] Custom error types for API boundaries
- [ ] Never ignore errors (use `_` explicitly if needed)

#### 5. Concurrency Patterns
- [ ] Goroutines must have clear lifecycle
- [ ] Use `context.Context` for cancellation
- [ ] Channels: sender closes, receiver checks
- [ ] Prefer `sync.Once` for one-time initialization

#### 6. Testing Standards
- [ ] Table-driven tests with subtests
- [ ] Test file naming: `*_test.go`
- [ ] Benchmark critical paths: `Benchmark*`
- [ ] Example tests for documentation: `Example*`
- [ ] Minimum 80% code coverage

#### 7. Documentation
- [ ] Package comment before `package` declaration
- [ ] Exported items have godoc comments
- [ ] Comments start with item name
- [ ] Use `// TODO(username):` for future work

### 🛠️ Validation Commands
```bash
# Format code
go fmt ./...

# Vet for suspicious constructs
go vet ./...

# Lint for style issues
golangci-lint run

# Test with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 📁 Project Structure Example
```
project/
├── cmd/           # Main applications
│   └── app/
│       └── main.go
├── internal/      # Private packages
│   ├── config/
│   ├── handler/
│   └── service/
├── pkg/           # Public packages
│   └── client/
├── test/          # Integration tests
├── docs/          # Documentation
└── scripts/       # Build/deploy scripts
```

### ⚠️ Exclusions
Do not refactor:
- `/vendor/` - External dependencies
- `*.pb.go` - Generated protobuf files
- `/generated/` - Any generated code
- `/.git/` - Version control

### 📊 Success Criteria
- [ ] Zero linting errors
- [ ] All tests passing
- [ ] Coverage ≥ 80%
- [ ] No `go vet` warnings
- [ ] Documentation for all exported items
- [ ] Consistent style across entire codebase

### 💡 Additional Best Practices
1. **Prefer simplicity**: Clear code > clever code
2. **Fail fast**: Return errors early
3. **Make zero values useful**
4. **Design APIs for testability**
5. **Use interfaces for abstraction, not just for mocking**

### 🔄 Refactoring Strategy
1. **Phase 1**: Fix critical issues (compilation, tests)
2. **Phase 2**: Apply naming conventions
3. **Phase 3**: Restructure packages if needed
4. **Phase 4**: Improve error handling
5. **Phase 5**: Add missing documentation
6. **Phase 6**: Optimize and add benchmarks

### 📝 Notes
- Run tests after each major change
- Commit frequently with descriptive messages
- Use feature branches for large refactors
- Document breaking changes in CHANGELOG.md

---

**Remember**: Good code is not just working code, it's code that others (including future you) can understand, maintain, and extend.