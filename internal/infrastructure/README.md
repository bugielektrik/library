# Infrastructure Layer

Technical infrastructure and application bootstrap components.

## Structure

```
infrastructure/
├── app/         # Application initialization and bootstrap
├── auth/        # JWT token generation and validation
├── config/      # Configuration management
├── logger/      # Structured logging setup
├── server/      # HTTP server configuration
└── store/       # Database connection management
```

## Components

### `app/` - Application Bootstrap
- **Purpose**: Wire all dependencies and start the application
- **Key File**: `app.go` - Creates repositories, caches, services, and HTTP server
- **Boot Order**: Logger → Config → DB → Cache → Auth → Use Cases → Server

### `auth/` - Authentication Services
- **JWT Service**: Token generation, validation, refresh
- **Password Service**: Hashing (bcrypt) and verification
- **Infrastructure services** (not domain) as they depend on external libraries

### `store/` - Database Management
- **Connection pooling** for PostgreSQL
- **Migration support** via `migrate` package
- **Error mappings** (SQL errors to domain errors)

### `server/` - HTTP Server
- **Chi router** configuration
- **Middleware setup** (CORS, logging, recovery)
- **Graceful shutdown** handling

## Usage

```go
// Application startup in cmd/api/main.go
app, err := app.New(ctx)
if err != nil {
    log.Fatal(err)
}
app.Run(ctx)
```

## Dependencies

Infrastructure services are created here and passed to use cases:
- **External libraries**: JWT, bcrypt, database drivers
- **Configuration**: Environment variables, secrets
- **Technical concerns**: Not business logic

## Key Principles

1. **No business logic** - Only technical infrastructure
2. **Created at startup** - Single instances, shared across requests
3. **Passed to use cases** - Via dependency injection container
4. **Configurable** - Via environment variables

See `.claude/adr/003-domain-services-vs-infrastructure.md` for architecture decisions.