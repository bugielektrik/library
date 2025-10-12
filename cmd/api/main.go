package main

import (
	"log"

	"library-service/internal/app"
)

// @title Library Service API
// @version 2.0
// @description Library management system with clean architecture
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@library.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token

/*
Application Entry Point

This is the main entry point for the Library Service API server. The application follows
a strict boot sequence orchestrated by internal/app/app.go:

BOOT SEQUENCE:
1. Logger initialization (Zap logger with structured logging)
   - DEV mode: console output with color
   - PROD mode: JSON output

2. Configuration loading (internal/infrastructure/config/)
   - Environment variables (.env file + system env)
   - Validation of required variables
   - Default values for optional settings

3. Database initialization (PostgreSQL via sqlx)
   - Connection pool setup
   - Health check + retry logic
   - Migration status verification

4. Cache initialization (Redis or Memory fallback)
   - Redis connection if REDIS_HOST provided
   - Memory cache fallback if Redis unavailable

5. Infrastructure service (auth layer)
   - JWT service (token generation/validation)
   - Password service (bcrypt hashing)

6. Repository layer (adapters/repository/)
   - PostgreSQL implementations for all entities
   - BaseRepository[T] generic pattern

7. Use case container (usecase/container.go)
   - Domain service instantiation
   - Use case wiring with dependencies
   - Complete dependency injection graph

8. HTTP server (Chi router)
   - Middleware chain setup
   - Route registration
   - Graceful shutdown handler

REQUIRED ENVIRONMENT VARIABLES:
- POSTGRES_DSN: PostgreSQL connection string
  Example: "postgres://library:password@localhost:5432/library?sslmode=disable"

- JWT_SECRET: Secret key for JWT token signing (MUST change in production)
  Example: "your-256-bit-secret-key-change-this-in-production"

OPTIONAL ENVIRONMENT VARIABLES:
- APP_MODE: "dev" (default) or "prod" - Controls logging format
- APP_PORT: Server port (default: 8080)
- APP_TIMEOUT: Request timeout (default: 30s)
- REDIS_HOST: Redis server address (default: localhost:6379)
  Note: If Redis unavailable, memory cache is used automatically
- JWT_EXPIRY: Access token lifetime (default: 24h)

GRACEFUL SHUTDOWN:
The application handles SIGINT and SIGTERM signals gracefully:
1. Stop accepting new connections
2. Wait for in-flight requests to complete (max 30s)
3. Close database connections
4. Close Redis connections
5. Flush logs and exit

COMMON FAILURE MODES:

1. "Failed to create application: config error"
   Cause: Missing or invalid POSTGRES_DSN or JWT_SECRET
   Fix: Check .env file and ensure required variables are set

2. "Failed to create application: database error"
   Cause: Cannot connect to PostgreSQL
   Fix: Ensure PostgreSQL is running (docker-compose up -d)
        Verify connection string is correct
        Check network connectivity to database

3. "Failed to create application: migration status error"
   Cause: Database schema is not up to date
   Fix: Run migrations with: make migrate-up
        Or: POSTGRES_DSN="..." go run cmd/migrate/main.go up

4. "Application error: bind: address already in use"
   Cause: Port 8080 is already in use by another process
   Fix: Kill existing process: lsof -ti:8080 | xargs kill -9
        Or: Change APP_PORT in .env file

5. "Application error: context deadline exceeded"
   Cause: Database queries timing out (slow queries or connection issues)
   Fix: Check database performance
        Verify indexes are created
        Increase APP_TIMEOUT if needed

TESTING THE APPLICATION:

Start the server:
  make dev           # Full stack (docker + migrations + server)
  make run           # Server only (requires PostgreSQL already running)

Health check:
  curl http://localhost:8080/health

Register a user:
  curl -X POST http://localhost:8080/api/v1/auth/register \
    -H "Content-Type: application/json" \
    -d '{"email":"test@example.com","password":"Test123!@#","full_name":"Test User"}'

Login:
  curl -X POST http://localhost:8080/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email":"test@example.com","password":"Test123!@#"}'

API Documentation:
  http://localhost:8080/swagger/index.html

For detailed architecture and development workflow, see .claude/CLAUDE.md
*/

func main() {
	// Create application
	application, err := app.New()
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	// Run application
	if err := application.Run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
