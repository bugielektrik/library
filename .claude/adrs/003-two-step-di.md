# ADR-003: Two-Step Dependency Injection

**Status:** Accepted

**Date:** 2024-01-16

**Decision Makers:** Project Architecture Team

## Context

In Clean Architecture, we need to wire dependencies from outer layers (infrastructure, adapters) to inner layers (use cases, domain). The question was: how should we structure dependency injection?

**Problem:** Where and how should we initialize and wire all components (database, repositories, services, use cases, handlers)?

**Requirements:**
- Infrastructure (DB, Redis, JWT manager) must be initialized first
- Repositories need database connections
- Use cases need repositories and domain services
- Handlers need use cases
- Must be clear where each component is created
- Must be easy for AI to understand and modify

## Decision

We implemented **Two-Step Dependency Injection**:

### Step 1: `app.go` - Bootstrap Infrastructure

```go
// internal/infrastructure/app/app.go
type App struct {
    Config      *config.Config
    Store       *store.PostgresStore
    RedisClient *redis.Client
    JWTManager  *auth.JWTManager
    Logger      *logrus.Logger
}

func New(cfg *config.Config) (*App, error) {
    // Initialize ONLY infrastructure
    db, err := store.NewPostgres(cfg.DatabaseDSN)
    if err != nil {
        return nil, err
    }

    redis := redis.NewClient(&redis.Options{
        Addr: cfg.RedisAddr,
    })

    jwtManager := auth.NewJWTManager(cfg.JWTSecret)
    logger := logrus.New()

    return &App{
        Config:      cfg,
        Store:       db,
        RedisClient: redis,
        JWTManager:  jwtManager,
        Logger:      logger,
    }, nil
}
```

**Responsibilities:**
- Database connections
- External service clients (Redis, etc.)
- Shared utilities (logger, JWT manager)
- Configuration loading

### Step 2: `container.go` - Wire Application Components

```go
// internal/infrastructure/container/container.go
type Container struct {
    // Repositories
    BookRepo   book.Repository
    MemberRepo member.Repository
    AuthorRepo author.Repository

    // Services
    BookService   *book.Service
    MemberService *member.Service

    // Use Cases
    CreateBookUC *bookops.CreateBookUseCase
    GetBookUC    *bookops.GetBookUseCase
    LoginUC      *authops.LoginUseCase

    // Handlers
    BookHandler *handlers.BookHandler
    AuthHandler *handlers.AuthHandler
}

func New(app *app.App) *Container {
    // Step 1: Repositories (need app.Store)
    bookRepo := postgres.NewBookRepository(app.Store.DB)
    memberRepo := postgres.NewMemberRepository(app.Store.DB)

    // Step 2: Domain Services (pure, no dependencies)
    bookService := book.NewService()
    memberService := member.NewService()

    // Step 3: Use Cases (need repos + services)
    createBookUC := bookops.NewCreateBookUseCase(bookRepo, bookService)
    getBookUC := bookops.NewGetBookUseCase(bookRepo)
    loginUC := authops.NewLoginUseCase(memberRepo, app.JWTManager)

    // Step 4: Handlers (need use cases)
    bookHandler := handlers.NewBookHandler(createBookUC, getBookUC)
    authHandler := handlers.NewAuthHandler(loginUC)

    return &Container{
        BookRepo:     bookRepo,
        BookService:  bookService,
        CreateBookUC: createBookUC,
        GetBookUC:    getBookUC,
        BookHandler:  bookHandler,
        AuthHandler:  authHandler,
    }
}
```

**Responsibilities:**
- Create repositories with database connections
- Create domain services
- Wire use cases with their dependencies
- Create handlers with use cases

## Consequences

### Positive

1. **Clear Separation of Concerns:**
   - `app.go`: Infrastructure (database, redis, external services)
   - `container.go`: Application logic (repos, services, use cases, handlers)

2. **Easy to Find Components:**
   ```
   Q: Where is database initialized?     A: app.go
   Q: Where is BookRepository created?   A: container.go
   Q: Where is LoginUseCase wired?       A: container.go
   ```

3. **Initialization Order is Explicit:**
   ```go
   // Clear dependency chain
   app.Store (DB connection)
   → bookRepo (needs DB)
   → bookService (no deps)
   → createBookUC (needs repo + service)
   → bookHandler (needs use case)
   ```

4. **Easy to Add New Components:** Follow the pattern
   ```go
   // Adding a new "Loan" feature to container.go:

   // Step 1: Repository
   loanRepo := postgres.NewLoanRepository(app.Store.DB)

   // Step 2: Service
   loanService := loan.NewService()

   // Step 3: Use Cases
   createLoanUC := loanops.NewCreateLoanUseCase(loanRepo, loanService, bookRepo, memberRepo)
   returnBookUC := loanops.NewReturnBookUseCase(loanRepo, loanService)

   // Step 4: Handler
   loanHandler := handlers.NewLoanHandler(createLoanUC, returnBookUC)
   ```

5. **AI-Friendly:** Claude Code can easily:
   - Find where to add new dependencies
   - Understand initialization order
   - See all use case wiring in one place

6. **Testable:** Can create test containers with mocks
   ```go
   func NewTestContainer() *Container {
       mockBookRepo := &mocks.MockBookRepository{}
       bookService := book.NewService()
       createBookUC := bookops.NewCreateBookUseCase(mockBookRepo, bookService)
       // ...
   }
   ```

### Negative

1. **Large container.go File:** As the app grows, container.go can become hundreds of lines
   - Mitigation: This is acceptable. Having all wiring in one place is better than scattered.
   - Future: Can split into `container_*.go` files if needed (e.g., `container_book.go`, `container_auth.go`)

2. **Manual Dependency Wiring:** No automatic DI framework (like Wire, Dig)
   - Mitigation: Manual wiring is explicit and easy to understand. No magic.
   - Benefit: AI can understand and modify (unlike reflection-based DI frameworks)

3. **All Dependencies Created Eagerly:** Even if not used
   - Mitigation: App starts in <1 second, so this is not a performance issue
   - Lazy loading would add complexity without benefit

4. **Circular Dependencies Not Prevented:** Compiler won't catch them
   - Mitigation: Clean Architecture rules prevent this (dependencies point inward)
   - If it happens, will get runtime panic (good for catching mistakes)

## Alternatives Considered

### Alternative 1: Single app.go with everything

```go
// ❌ Everything in one place
type App struct {
    // Infrastructure
    DB          *sql.DB
    Redis       *redis.Client

    // Repositories
    BookRepo    book.Repository

    // Services
    BookService *book.Service

    // Use Cases
    CreateBookUC *bookops.CreateBookUseCase

    // Handlers
    BookHandler *handlers.BookHandler
}
```

**Why not chosen:**
- Mixes infrastructure and application concerns
- Harder to test (can't easily swap out infrastructure)
- app.go becomes hundreds of lines
- Not clear where to find things

### Alternative 2: DI Framework (google/wire, uber/dig)

```go
// wire.go
//go:build wireinject

func InitializeApp(cfg *config.Config) (*App, error) {
    wire.Build(
        NewDatabase,
        NewRedis,
        NewBookRepository,
        NewBookService,
        NewCreateBookUseCase,
        // ...
    )
    return &App{}, nil
}
```

**Why not chosen:**
- Adds dependency and learning curve
- Uses code generation or reflection (harder for AI to understand)
- Compile-time safety not worth the complexity
- Manual wiring is simple and explicit for our size

### Alternative 3: Service Locator Pattern

```go
// ❌ Service locator anti-pattern
type ServiceLocator struct {
    services map[string]interface{}
}

func (sl *ServiceLocator) Get(name string) interface{} {
    return sl.services[name]
}

// In handler:
bookRepo := serviceLocator.Get("bookRepo").(book.Repository)
```

**Why not chosen:**
- Loses compile-time type safety
- Dependencies not explicit (hidden inside service locator)
- Anti-pattern in most modern architectures
- Harder to test and reason about

### Alternative 4: Dependency Injection per Layer

```go
// Each layer has its own container
type RepositoryContainer struct { /* ... */ }
type UseCaseContainer struct { /* ... */ }
type HandlerContainer struct { /* ... */ }
```

**Why not chosen:**
- More complex (multiple containers to manage)
- Wiring split across multiple files
- Harder to see full dependency graph
- Doesn't provide enough benefit over single container

## Implementation Guidelines

**When to modify app.go:**
- ✅ Adding new database connection
- ✅ Adding new external service client (Redis, S3, etc.)
- ✅ Adding shared infrastructure (logger, metrics)
- ❌ Adding repository (goes in container.go)
- ❌ Adding use case (goes in container.go)

**When to modify container.go:**
- ✅ Adding new repository
- ✅ Adding new domain service
- ✅ Adding new use case
- ✅ Adding new handler
- ✅ Wiring new dependencies
- ❌ Adding database connection (goes in app.go)

**Initialization order in container.go:**
```go
func New(app *app.App) *Container {
    // 1. Repositories (need infrastructure)
    repos := createRepositories(app)

    // 2. Domain Services (no dependencies)
    services := createDomainServices()

    // 3. Use Cases (need repos + services)
    useCases := createUseCases(repos, services, app)

    // 4. Handlers (need use cases)
    handlers := createHandlers(useCases)

    return &Container{/* ... */}
}
```

## Real-World Flow

```go
// cmd/api/main.go
func main() {
    cfg := config.Load()

    // Step 1: Bootstrap infrastructure
    app, err := app.New(cfg)
    if err != nil {
        log.Fatal(err)
    }
    defer app.Close()

    // Step 2: Wire application components
    container := container.New(app)

    // Step 3: Setup HTTP router
    router := routes.NewRouter(container)

    // Step 4: Start server
    server := &http.Server{
        Addr:    cfg.HTTPAddr,
        Handler: router,
    }

    log.Fatal(server.ListenAndServe())
}
```

## Validation

After 6 months:
- ✅ Adding new feature takes 5-10 minutes (follow the pattern)
- ✅ Container.go is 300 lines (manageable for 8 domains)
- ✅ Zero circular dependency issues (Clean Architecture prevents them)
- ✅ Easy to create test containers with mocks
- ✅ AI can wire new components correctly 95% of the time

## References

- [Dependency Injection in Go](https://github.com/google/wire/blob/main/docs/best-practices.md)
- [Clean Architecture Dependency Rule](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- `.claude/examples/README.md` - Examples of adding new components
- `.claude/flows.md` - Visual diagram of two-step DI bootstrap

## Related ADRs

- [ADR-001: Clean Architecture](./001-clean-architecture.md) - Why we need DI
- [ADR-002: Domain Services](./002-domain-services.md) - What gets wired in container.go

---

**Last Reviewed:** 2024-01-16

**Next Review:** 2024-07-16 (or if considering DI framework)
