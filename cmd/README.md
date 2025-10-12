# Command Line Applications

**Entry points for the Library Management System.**

## Purpose

This directory contains all executable entry points:
- **API Server** (`cmd/api/`) - REST API service
- **Worker** (`cmd/worker/`) - Background job processor
- **Migration Tool** (`cmd/migrate/`) - Database migration utility

Each command is a standalone application with its own `main.go`.

## Directory Structure

```
cmd/
├── api/              # REST API Server
│   └── main.go
│
├── worker/           # Background Worker
│   └── main.go
│
└── migrate/          # Database Migration Tool
    └── main.go
```

## API Server

**Entry Point**: `cmd/api/main.go`

### Purpose
- Serve REST API endpoints
- Handle HTTP requests
- Manage application lifecycle
- Set up dependency injection

### Running

```bash
# Development
make run
# or: go run ./cmd/api

# Production build
make build-api
./bin/library-api

# Debug (VSCode)
# Use "Debug API Server" configuration
```

### Configuration

Environment variables (`.env`):
```env
# Server
SERVER_PORT=8080
SERVER_HOST=0.0.0.0

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=library
DB_USER=library
DB_PASSWORD=library123

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379

# Logging
LOG_LEVEL=info  # debug, info, warn, error
```

### Structure

```go
func main() {
    // 1. Load configuration
    cfg := loadConfig()

    // 2. Initialize infrastructure
    db := initDatabase(cfg)
    redis := initRedis(cfg)

    // 3. Create repositories
    bookRepo := repository.NewPostgresBookRepository(db)
    bookCache := cache.NewRedisBookCache(redis)

    // 4. Create domain service
    bookService := book.NewService()

    // 5. Create use cases
    createBookUC := usecase.NewCreateBookUseCase(bookRepo, bookService, bookCache)

    // 6. Create handler
    bookHandler := http.NewBookHandler(createBookUC, ...)

    // 7. Setup routes
    router := gin.Default()
    setupRoutes(router, bookHandler)

    // 8. Start server
    router.Run(":8080")
}
```

### Endpoints

- `GET /health` - Health check
- `GET /swagger/index.html` - API documentation
- `POST /api/v1/books` - Create book
- `GET /api/v1/books` - List books
- `GET /api/v1/books/:id` - Get book
- ... (see [OpenAPI spec](../api/openapi/swagger.yaml))

## Worker

**Entry Point**: `cmd/worker/main.go`

### Purpose
- Process background jobs
- Handle async tasks
- Scheduled operations
- Email notifications, reports, etc.

### Running

```bash
# Development
make run-worker
# or: go run ./cmd/worker

# Production build
make build-worker
./bin/library-worker

# Debug (VSCode)
# Use "Debug Worker" configuration
```

### Job Types

- **Email Notifications**: Send subscription reminders
- **Data Cleanup**: Archive old records
- **Report Generation**: Generate usage reports
- **Scheduled Tasks**: Subscription expiry checks

### Structure

```go
func main() {
    // 1. Connect to job queue (Redis)
    queue := initQueue()

    // 2. Register job handler
    queue.Register("send_email", handleEmailJob)
    queue.Register("cleanup", handleCleanupJob)

    // 3. Start workers
    queue.Start(4) // 4 concurrent workers

    // 4. Wait for shutdown signal
    waitForShutdown()

    // 5. Graceful shutdown
    queue.Shutdown()
}
```

## Migration Tool

**Entry Point**: `cmd/migrate/main.go`

### Purpose
- Run database migrations
- Rollback migrations
- Create new migration files
- Check migration status

### Running

```bash
# Apply migrations
make migrate-up
# or: go run ./cmd/migrate up

# Rollback last migration
make migrate-down
# or: go run ./cmd/migrate down

# Create new migration
make migrate-create name=add_books_table

# Debug (VSCode)
# Use "Debug Migration" configuration
```

### Commands

```bash
# Up - Apply all pending migrations
./bin/library-migrate up

# Down - Rollback last migration
./bin/library-migrate down

# Create - Create new migration files
./bin/library-migrate create add_index_to_books

# Status - Show migration status
./bin/library-migrate status

# Force - Set migration version (use with caution)
./bin/library-migrate force 000005
```

### Migration Files

Located in `migrations/`:
```
migrations/
├── 000001_initial_schema.up.sql
├── 000001_initial_schema.down.sql
├── 000002_add_books_table.up.sql
└── 000002_add_books_table.down.sql
```

## Building

### Build All

```bash
make build
# Creates:
# - bin/library-api
# - bin/library-worker
# - bin/library-migrate
```

### Build Individual

```bash
make build-api       # API server only
make build-worker    # Worker only
make build-migrate   # Migration tool only
```

### Cross-Platform Build

```bash
# Linux
GOOS=linux GOARCH=amd64 make build-api

# macOS
GOOS=darwin GOARCH=arm64 make build-api

# Windows
GOOS=windows GOARCH=amd64 make build-api
```

### Docker Build

```bash
# Build images
make docker-build

# Run specific service
docker run -p 8080:8080 library-api:latest
docker run library-worker:latest
docker run library-migrate:latest up
```

## Deployment

### Development

```bash
# Start all service
make dev
# Runs: docker-up → migrate-up → run (API)
```

### Production

```bash
# Build optimized binaries
CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bin/api ./cmd/api

# Or use Makefile
make build

# Deploy
scp bin/library-api server:/opt/library/
scp bin/library-worker server:/opt/library/
```

### Docker Compose

```bash
# Start all service
make up

# Stop all service
make down

# View logs
make docker-logs
```

## Configuration

### Environment Variables

```env
# Common
GO_ENV=development          # development, production, test
LOG_LEVEL=info             # debug, info, warn, error

# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=library
DB_USER=library
DB_PASSWORD=library123
DB_SSL_MODE=disable        # disable, require, verify-full

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# Server (API only)
SERVER_PORT=8080
SERVER_HOST=0.0.0.0
SERVER_READ_TIMEOUT=30s
SERVER_WRITE_TIMEOUT=30s

# Worker (Worker only)
WORKER_CONCURRENCY=4
WORKER_QUEUE=library_jobs
```

### Configuration File

```go
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Redis    RedisConfig
    Worker   WorkerConfig
}

func LoadConfig() (*Config, error) {
    // Load from .env or environment variables
}
```

## Monitoring

### Health Checks

```bash
# API health
curl http://localhost:8080/health

# Database connection
curl http://localhost:8080/health/db

# Redis connection
curl http://localhost:8080/health/redis
```

### Metrics

- Request latency
- Database query time
- Cache hit rate
- Job queue length (worker)

## Troubleshooting

### API Won't Start

```bash
# Check port availability
lsof -i :8080

# Check store connection
psql -h localhost -U library -d library

# Check logs
./bin/library-api 2>&1 | tee api.log
```

### Migration Failed

```bash
# Check migration status
./bin/library-migrate status

# Force to specific version
./bin/library-migrate force 000003

# Rollback and retry
./bin/library-migrate down
./bin/library-migrate up
```

### Worker Not Processing Jobs

```bash
# Check Redis connection
redis-cli ping

# Check job queue
redis-cli llen library_jobs

# Check worker logs
./bin/library-worker 2>&1 | tee worker.log
```

## References

- [Makefile Commands](../Makefile)
- [Development Guide](../docs/guides/DEVELOPMENT.md)
- [Deployment Guide](../docs/guides/DEPLOYMENT.md)
