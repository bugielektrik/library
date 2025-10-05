# Claude Code Go Project Architecture Refactoring Guide

## Refactor the Go codebase architecture to optimize for vibecoding workflow with Claude Code.

## 📚 Reference Documentation
- **Google Go Style Guide**: https://google.github.io/styleguide/go/best-practices
- **Effective Go**: https://go.dev/doc/effective_go
- **Standard Go Project Layout**: https://github.com/golang-standards/project-layout

## Current Architecture Overview
- **Go version:** [e.g., 1.21.5]
- **Project type:** [e.g., REST API, gRPC service, CLI tool, web app]
- **Dependencies management:** [e.g., go.mod, vendor]
- **Codebase size:** [e.g., ~150 files, 40k LOC, largest file 3000+ lines]
- **Main pain points:** 
  - [e.g., "Business logic mixed with HTTP handlers"]
  - [e.g., "Database queries scattered throughout"]
  - [e.g., "No clear domain boundaries"]
  - [e.g., "Inconsistent error handling"]
- **Team size:** [number of developers]

## Specific Problems to Solve
1. **God Structs:** `internal/server/server.go` has 50+ methods doing everything
2. **Data Layer:** SQL queries hardcoded in handlers instead of repository pattern
3. **Dependencies:** Circular dependencies between `pkg/auth` and `pkg/user`
4. **Testing:** Can't unit test handlers without spinning up entire database
5. **Error Handling:** Inconsistent error types and handling across packages
6. **Configuration:** Config scattered across init() functions and global variables

## Goals (Priority Order)
1. Implement clean architecture with clear domain boundaries
2. Establish consistent error handling with error wrapping
3. Create testable code with dependency injection
4. Reduce file sizes to under 500 lines
5. Enable concurrent development without conflicts
6. Improve build and test execution time

## Target Architecture

```
.
├── cmd/
│   ├── api/
│   │   └── main.go           # HTTP server entry point
│   ├── worker/
│   │   └── main.go           # Background jobs/queue processor
│   └── migrate/
│       └── main.go           # Database migration tool
├── internal/
│   ├── domain/               # Core business logic (no external deps)
│   │   ├── user/
│   │   │   ├── user.go       # User entity & value objects
│   │   │   ├── repository.go # Repository interface
│   │   │   ├── service.go    # Domain service (business rules)
│   │   │   └── errors.go     # Domain-specific errors
│   │   ├── order/
│   │   │   ├── order.go      # Order aggregate root
│   │   │   ├── item.go       # Order item value object
│   │   │   ├── status.go     # Order status enum
│   │   │   ├── repository.go
│   │   │   └── service.go
│   │   ├── payment/
│   │   │   ├── payment.go
│   │   │   ├── method.go
│   │   │   └── repository.go
│   │   └── notification/
│   │       ├── notification.go
│   │       └── service.go
│   ├── usecase/              # Use case services (orchestration)
│   │   ├── user/
│   │   │   ├── create.go     # CreateUserUseCase
│   │   │   ├── authenticate.go
│   │   │   ├── update_profile.go
│   │   │   ├── change_password.go
│   │   │   └── interfaces.go # UseCase dependencies
│   │   ├── order/
│   │   │   ├── create.go
│   │   │   ├── checkout.go   # Complex multi-step use case
│   │   │   ├── cancel.go
│   │   │   ├── list.go
│   │   │   └── interfaces.go
│   │   └── reporting/
│   │       ├── sales_report.go
│   │       └── user_activity.go
│   ├── adapters/             # External interfaces implementation
│   │   ├── http/
│   │   │   ├── handlers/
│   │   │   │   ├── user_handler.go
│   │   │   │   ├── order_handler.go
│   │   │   │   ├── health_handler.go
│   │   │   │   └── base_handler.go
│   │   │   ├── middleware/
│   │   │   │   ├── auth.go
│   │   │   │   ├── cors.go
│   │   │   │   ├── ratelimit.go
│   │   │   │   ├── logging.go
│   │   │   │   └── recovery.go
│   │   │   ├── dto/
│   │   │   │   ├── user_dto.go
│   │   │   │   ├── order_dto.go
│   │   │   │   └── error_response.go
│   │   │   └── router.go
│   │   ├── grpc/
│   │   │   ├── server.go
│   │   │   └── services/
│   │   ├── repository/
│   │   │   ├── postgres/
│   │   │   │   ├── user_repository.go
│   │   │   │   ├── order_repository.go
│   │   │   │   ├── transaction.go
│   │   │   │   └── migrations/
│   │   │   ├── redis/
│   │   │   │   ├── cache_repository.go
│   │   │   │   └── session_store.go
│   │   │   └── mock/
│   │   │       └── repositories.go
│   │   ├── email/
│   │   │   ├── smtp_sender.go
│   │   │   └── templates/
│   │   ├── payment/
│   │   │   ├── stripe_gateway.go
│   │   │   └── paypal_gateway.go
│   │   └── storage/
│   │       ├── s3_storage.go
│   │       └── local_storage.go
│   └── infrastructure/       # Technical concerns
│       ├── config/
│       │   ├── config.go
│       │   ├── database.go
│       │   ├── redis.go
│       │   └── validators.go
│       ├── logger/
│       │   ├── logger.go
│       │   └── context.go
│       ├── database/
│       │   ├── postgres.go
│       │   ├── redis.go
│       │   └── connection_pool.go
│       ├── auth/
│       │   ├── jwt.go
│       │   ├── oauth.go
│       │   └── permissions.go
│       └── monitoring/
│           ├── metrics.go
│           ├── tracing.go
│           └── health.go
├── pkg/                      # Shared packages (can be extracted)
│   ├── errors/
│   │   ├── errors.go
│   │   ├── codes.go
│   │   └── http_errors.go
│   ├── validator/
│   │   ├── validator.go
│   │   └── custom_rules.go
│   ├── pagination/
│   │   ├── cursor.go
│   │   └── paginator.go
│   ├── crypto/
│   │   ├── hash.go
│   │   └── random.go
│   └── timeutil/
│       └── time.go
├── api/                      # API definitions
│   ├── openapi/
│   │   └── swagger.yaml
│   └── protobuf/
│       └── service.proto
├── scripts/
│   ├── setup.sh
│   ├── test.sh
│   └── build.sh
├── deployments/
│   ├── docker/
│   │   ├── Dockerfile
│   │   └── docker-compose.yml
│   ├── kubernetes/
│   │   ├── deployment.yaml
│   │   └── service.yaml
│   └── terraform/
├── test/
│   ├── integration/
│   │   ├── user_test.go
│   │   └── order_test.go
│   ├── e2e/
│   │   └── api_test.go
│   └── fixtures/
│       └── test_data.go
├── docs/
│   ├── architecture.md
│   └── api.md
├── .github/
│   └── workflows/
│       └── ci.yml
├── Makefile
├── go.mod
├── go.sum
├── .golangci.yml
├── .env.example
└── README.md
```

## Refactoring Requirements

### 1. **Domain-Driven Design Implementation**
- Extract pure domain logic with no framework dependencies
- Define clear aggregate roots and value objects
- Implement repository interfaces in domain layer
- Use dependency injection for all external dependencies

### 2. **Error Handling Strategy**
```go
// Implement consistent error types
type Error struct {
    Code    string
    Message string
    Err     error
    Details map[string]interface{}
}

// With error wrapping
return fmt.Errorf("failed to create user: %w", err)
```

### 3. **Dependency Injection Pattern**
```go
// Use struct embedding for dependencies
type UserService struct {
    repo UserRepository
    events EventPublisher
    logger Logger
}

// Constructor with explicit dependencies
func NewUserService(repo UserRepository, events EventPublisher, logger Logger) *UserService
```

### 4. **Testing Structure**
- Table-driven tests for all business logic
- Interfaces for all external dependencies
- Mock generation with `mockgen` or `testify/mock`
- Separate integration tests with build tags

### 5. **Configuration Management**
```go
// Single configuration struct
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Redis    RedisConfig
}

// Load with envconfig or viper
func LoadConfig() (*Config, error)
```

## Go-Specific Standards

### 📋 Comprehensive Refactoring Checklist

#### 1. Code Structure & Organization
- [ ] Package names: lowercase, single word, no underscores
- [ ] One package per directory
- [ ] Imports grouped: standard → external → internal (with blank lines)
- [ ] Consistent file naming: `snake_case.go`

#### 2. Naming Conventions
- [ ] **Exported**: `CamelCase` for public APIs
- [ ] **Unexported**: `camelCase` for internal use
- [ ] **Interfaces**: verb + "er" suffix (e.g., `Reader`, `Writer`, `UserCreator`)
- [ ] **Constants**: `CamelCase` or `SCREAMING_SNAKE_CASE` for exported
- [ ] **Acronyms**: Keep consistent case (e.g., `URLParser`, not `UrlParser`)
- [ ] **Receiver names**: 1-2 letters, consistent across methods
- [ ] **Struct naming**: PascalCase for exported, camelCase for unexported

#### 3. Function & Method Design
- [ ] Functions do one thing well
- [ ] Early returns for error cases
- [ ] Named return values only when they clarify
- [ ] Prefer value receivers unless mutation needed
- [ ] Context as first parameter in functions
- [ ] Errors as last return value

#### 4. Error Handling Best Practices
```go
// ✅ Good - Add context and wrap errors
if err != nil {
    return fmt.Errorf("processing user %d: %w", userID, err)
}

// ❌ Bad - No context
if err != nil {
    return err
}
```
- [ ] Check errors immediately
- [ ] Add context with `fmt.Errorf` and `%w` for wrapping
- [ ] Custom error types for API boundaries
- [ ] Never ignore errors (use `_` explicitly if intentional)

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
- [ ] Integration tests tagged with `//go:build integration`

#### 7. Documentation Requirements
- [ ] Package comment before `package` declaration
- [ ] Exported items have godoc comments
- [ ] Comments start with item name
- [ ] Use `// TODO(username):` for future work
- [ ] Code examples in godoc where helpful

### Code Organization
- **Package naming:** Lowercase, single word, no underscores
- **File naming:** lowercase with underscores (e.g., `user_service.go`)
- **Interface naming:** Verb+er pattern (e.g., `Reader`, `UserCreator`)
- **Struct naming:** PascalCase for exported, camelCase for unexported

### Best Practices
- **One interface per file** when interface is substantial
- **Prefer composition over inheritance**
- **Return early** to reduce nesting
- **Named returns** only for documentation purposes
- **Context as first parameter** in functions
- **Errors as last return value**
- **Prefer simplicity**: Clear code > clever code
- **Fail fast**: Return errors early
- **Make zero values useful**
- **Design APIs for testability**
- **Use interfaces for abstraction, not just for mocking**

### Performance Guidelines
- **Preallocate slices** when size is known
- **Use sync.Pool** for frequently allocated objects
- **Benchmark critical paths** with `testing.B`
- **Profile with pprof** for bottlenecks
- **Minimize allocations** in hot paths

## Constraints & Exclusions

### Must Maintain
- Must maintain backward compatibility with existing REST API
- Cannot break existing client SDKs
- Database schema changes require migration scripts
- Zero downtime deployment required
- Must support Go 1.19+ (no newer features)
- Keep compile time under 30 seconds
- Docker image size under 50MB

### Do Not Refactor
- `/vendor/` - External dependencies
- `*.pb.go` - Generated protobuf files
- `/generated/` - Any generated code
- `/.git/` - Version control
- Existing protobuf definitions
- Database schema (only add migrations)
- Third-party service integrations
- Authentication flow
- API versioning strategy

## Technical Specifications

### Dependencies to Use
```go
// Router
"github.com/gorilla/mux" or "github.com/gin-gonic/gin"

// Database
"github.com/jmoiron/sqlx"
"github.com/lib/pq"

// Migrations
"github.com/golang-migrate/migrate/v4"

// Logging
"go.uber.org/zap"

// Configuration
"github.com/kelseyhightower/envconfig"

// Testing
"github.com/stretchr/testify"
"github.com/golang/mocks"

// Validation
"github.com/go-playground/validator/v10"
```

### Linting and Formatting
```yaml
# .golangci.yml configuration
linters:
  enable:
    - gofmt
    - goimports
    - golint
    - govet
    - ineffassign
    - misspell
    - unconvert
    - prealloc
    - nakedret
    - gocritic
```

## Migration Strategy

### 🔄 Refactoring Strategy Overview
1. **Phase 1**: Fix critical issues (compilation, tests)
2. **Phase 2**: Apply naming conventions and code organization
3. **Phase 3**: Restructure packages if needed
4. **Phase 4**: Improve error handling
5. **Phase 5**: Add missing documentation
6. **Phase 6**: Optimize and add benchmarks

### Phase 1: Foundation (Week 1)
- [ ] Set up new directory structure
- [ ] Create domain entities and value objects
- [ ] Define repository interfaces
- [ ] Implement error package
- [ ] Set up dependency injection container

### Phase 2: Core Domain (Week 2)
- [ ] Extract business logic to domain services
- [ ] Implement repository pattern for data access
- [ ] Create usecase services (business orchestration)
- [ ] Add domain event system
- [ ] Write unit tests for domain logic

### Phase 3: Adapters (Week 3)
- [ ] Refactor HTTP handlers to thin controllers
- [ ] Implement DTO layer for request/response
- [ ] Create middleware pipeline
- [ ] Add OpenAPI documentation
- [ ] Integration tests for HTTP layer

### Phase 4: Infrastructure (Week 4)
- [ ] Centralize configuration management
- [ ] Implement structured logging
- [ ] Add metrics and tracing
- [ ] Create health check endpoints
- [ ] Performance testing and optimization

### 🛠️ Validation Commands

Run these commands after each refactoring phase:

```bash
# Format code
go fmt ./...
goimports -w .

# Vet for suspicious constructs
go vet ./...

# Lint for style issues
golangci-lint run

# Test with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Race detection
go test -race ./...

# Check for security issues
gosec ./...
```

## Success Metrics

### 📊 Success Criteria
- [ ] Zero linting errors (`golangci-lint run`)
- [ ] All tests passing (`go test ./...`)
- [ ] Coverage ≥ 80% for business logic, ≥ 60% overall
- [ ] No `go vet` warnings
- [ ] Documentation for all exported items
- [ ] Consistent style across entire codebase
- [ ] **Build Time:** <30 seconds for full build
- [ ] **Test Execution:** <2 minutes for unit tests
- [ ] **Cyclomatic Complexity:** <10 per function
- [ ] **Code Duplication:** <3% (measured by dupl)
- [ ] **Response Time:** p99 <100ms for API endpoints
- [ ] **Memory Usage:** <100MB for typical load
- [ ] **Zero data races** detected by race detector

## Current Code Problems

### Example 1: Fat Handler
```go
// CURRENT: Everything in handler
func (s *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
    // 300+ lines including:
    // - Request parsing
    // - Validation
    // - Business logic
    // - Database queries
    // - Error handling
    // - Response formatting
}
```

### Example 2: Global State
```go
// CURRENT: Global variables everywhere
var (
    DB     *sql.DB
    Cache  *redis.Client
    Config Configuration
)

func init() {
    // Initialization chaos
}
```

## Desired Outcome

### Clean Handler Example
```go
// DESIRED: Thin handler
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.respondError(w, errors.ErrInvalidInput.Wrap(err))
        return
    }

    user, err := h.userService.CreateUser(r.Context(), req.ToCommand())
    if err != nil {
        h.respondError(w, err)
        return
    }

    h.respondJSON(w, http.StatusCreated, UserResponse.FromDomain(user))
}
```

### Clean Service Example
```go
// DESIRED: Pure business logic
func (s *UserService) CreateUser(ctx context.Context, cmd CreateUserCommand) (*User, error) {
    if err := cmd.Validate(); err != nil {
        return nil, errors.ErrValidation.Wrap(err)
    }

    user := NewUser(cmd.Email, cmd.Name)
    
    if exists, err := s.repo.EmailExists(ctx, user.Email); err != nil {
        return nil, fmt.Errorf("checking email existence: %w", err)
    } else if exists {
        return nil, errors.ErrUserExists
    }

    if err := s.repo.Create(ctx, user); err != nil {
        return nil, fmt.Errorf("creating user: %w", err)
    }

    s.events.Publish(ctx, UserCreatedEvent{UserID: user.ID})
    
    return user, nil
}
```

## Performance Baseline
- **Current API Response Time:** [e.g., p50: 200ms, p99: 2s]
- **Database Query Performance:** [e.g., "User listing takes 500ms"]
- **Memory Usage:** [e.g., "API server uses 500MB idle, 2GB under load"]
- **Goroutine Leaks:** [e.g., "Grows to 10k goroutines after 24h"]
- **Build Time:** [e.g., "Full build takes 2 minutes"]

## Security Considerations
- **Authentication:** JWT tokens with refresh mechanism
- **Authorization:** Role-based access control (RBAC)
- **Input Validation:** All inputs sanitized and validated
- **SQL Injection:** Use parameterized queries only
- **Secrets Management:** Use environment variables, never hardcode
- **HTTPS:** TLS 1.3 only, proper certificate management
- **Rate Limiting:** Per-user and per-IP limits

## Developer Experience Requirements
- **Local Setup:** One command to run (`make dev`)
- **Hot Reload:** Code changes reflect immediately
- **Debugging:** Delve debugger configuration included
- **Documentation:** Swagger UI available locally
- **Database Seeding:** Test data available (`make seed`)
- **Pre-commit Hooks:** Linting and formatting automated

## Decision Points Needing Input

1. **HTTP Framework:** Keep `net/http`, use `gin`, `echo`, or `fiber`?
2. **Database Access:** Raw SQL, `sqlx`, `gorm`, or `ent`?
3. **API Style:** REST, GraphQL, gRPC, or hybrid?
4. **Monorepo vs Multi-repo:** for microservices evolution?
5. **Event System:** In-memory, Redis Pub/Sub, or Kafka?
6. **Caching Strategy:** Redis, in-memory, or both?
7. **Service Mesh:** Direct calls or through Istio/Linkerd?

## Collaboration Approach

1. **First:** Analyze current codebase structure and create detailed report
2. **Second:** Propose 2-3 architectural approaches with trade-offs
3. **Third:** Create proof-of-concept for one module
4. **Fourth:** Implement incrementally with review checkpoints
5. **Ask questions** about ambiguous business rules or technical constraints

## First Step

Analyze the `internal/` directory and provide:
1. Dependency graph visualization
2. Complexity metrics per package
3. List of circular dependencies
4. Identification of god objects/packages
5. Proposed refactoring order based on risk/impact

## Leverage Claude Code Capabilities

```bash
# Use these Claude Code commands effectively
/ask "Should we use repository pattern or active record?"
/explain "internal/server/server.go"  # Before refactoring
/test # After each module refactor
/debug # For migration issues
/docs # Generate package documentation
```

## Quick Start Commands

```bash
# Initialize refactoring
git checkout -b refactor/clean-architecture
mkdir -p internal/domain internal/adapters internal/usecase
go mod tidy

# Start with domain modeling
touch internal/domain/user/user.go
touch internal/domain/user/repository.go
touch internal/domain/user/service.go

# Run tests continuously
go test -race ./...

# Check for issues
golangci-lint run
go vet ./...
```

## Validation Checklist

### Phase 1 Complete When:
- [ ] New directory structure created
- [ ] Domain entities defined with validation
- [ ] Repository interfaces established
- [ ] Error handling package implemented
- [ ] Basic DI container working

### Phase 2 Complete When:
- [ ] All business logic in domain services
- [ ] Repository pattern fully implemented
- [ ] Use cases orchestrating domain logic
- [ ] Domain events publishing
- [ ] 80% test coverage on domain

### Phase 3 Complete When:
- [ ] HTTP handlers under 50 lines each
- [ ] All DTOs mapping correctly
- [ ] Middleware chain working
- [ ] OpenAPI spec generated
- [ ] Integration tests passing

### Phase 4 Complete When:
- [ ] Configuration centralized
- [ ] Structured logging throughout
- [ ] Metrics exposed on /metrics
- [ ] Health checks on /health
- [ ] Performance benchmarks met

---

Please analyze the current structure and implement improvements following Go best practices and idiomatic patterns. Focus on creating a maintainable, testable, and performant Go application architecture.

**Priority:** Start with [highest risk module] and create a working example before proceeding with full refactoring.

## 📝 Important Reminders

### Development Best Practices
- Run tests after each major change
- Commit frequently with descriptive messages
- Use feature branches for large refactors
- Document breaking changes in CHANGELOG.md

### Core Principles
**Remember**: Good code is not just working code, it's code that others (including future you) can understand, maintain, and extend.

- **Prefer simplicity**: Clear code > clever code
- **Fail fast**: Return errors early
- **Make zero values useful**: Design structs so their zero value is ready to use
- **Design for testability**: Make dependencies explicit and interfaces narrow
- **Use interfaces wisely**: For abstraction, not just for mocking

**Note:** If you encounter any unclear requirements or need to make architectural decisions, please ask for clarification rather than making assumptions.