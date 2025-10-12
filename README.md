# Library Management System

A modern library management system built with Go, following clean architecture principles and optimized for vibecoding with Claude Code.

📚 **[Quick Start Guide](./docs/guides/QUICKSTART.md)** | 🏗️ **[Architecture](./docs/architecture.md)** | 🧪 **[Development Guide](./docs/guides/DEVELOPMENT.md)** | 📦 **[Package Overview](./docs/package-overview.md)**

## Architecture

This project follows clean architecture with clear separation of concerns:

```
library/
├── cmd/                    # Application entry points
│   ├── api/               # REST API server
│   ├── worker/            # Background worker
│   └── migrate/           # Database migration tool
├── internal/              # Private application code
│   ├── domain/           # Business entities and logic
│   │   ├── book/         # Book domain
│   │   ├── user/         # User domain
│   │   └── ...
│   ├── usecase/          # Application business rules
│   │   ├── book/         # Book use cases
│   │   ├── user/         # User use cases
│   │   └── ...
│   ├── adapters/         # External service integrations
│   │   ├── grpc/         # gRPC server
│   │   ├── email/        # Email sending
│   │   ├── payment/      # Payment processing
│   │   └── storage/      # File storage
│   ├── infrastructure/   # Infrastructure concerns
│   │   ├── auth/         # Authentication
│   │   ├── config/       # Configuration
│   │   ├── database/     # Database connection
│   │   ├── http/         # HTTP server and routing
│   │   └── logger/       # Logging
│   └── app/             # Application initialization
├── pkg/                  # Shared utilities
│   ├── validator/        # Input validation
│   ├── pagination/       # Pagination helpers
│   ├── crypto/          # Cryptography utilities
│   └── timeutil/        # Time utilities
├── api/                  # API definitions
│   └── openapi/         # OpenAPI/Swagger specs
├── migrations/           # Database migrations
├── test/                # Tests
├── scripts/             # Build and deployment scripts
└── deployments/         # Deployment configurations
    ├── docker/          # Docker and docker-compose
    ├── kubernetes/      # Kubernetes manifests
    └── terraform/       # Terraform IaC
```

## Prerequisites

- Go 1.25 or higher
- PostgreSQL 15+
- Redis 7+
- Docker and Docker Compose (optional)

## Getting Started

### ⚡ Quick Setup (Recommended)

Get your development environment ready in minutes:

```bash
# Clone the repository
git clone <repository-url>
cd library

# Run automated setup (checks prerequisites, installs deps, starts service, seeds data)
./scripts/dev-setup.sh

# Start developing!
make run
```

The setup script will:
- ✅ Check prerequisites (Go, Docker, Make)
- ✅ Install dependencies and tools
- ✅ Configure git hooks for quality checks
- ✅ Start Docker services (PostgreSQL, Redis)
- ✅ Run database migrations
- ✅ Seed development data (test users & books)
- ✅ Build the project

**Test Accounts** (after seeding):
- `admin@library.com` / `Admin123!@#`
- `user@library.com` / `User123!@#`
- `premium@library.com` / `Premium123!@#`

### 📖 Manual Setup

If you prefer step-by-step setup:

#### 1. Clone the repository

```bash
git clone <repository-url>
cd library
```

#### 2. Set up environment variables

```bash
cp .env.example .env
# Edit .env with your configuration
```

#### 3. Install dependencies

```bash
go mod download
```

#### 4. Install git hooks (optional but recommended)

```bash
make install-hooks
```

Pre-commit hooks will automatically run:
- Code formatting (gofmt, goimports)
- Go vet checks
- Unit tests
- Log file detection

#### 5. Start Docker services

```bash
make up  # Starts PostgreSQL and Redis
```

#### 6. Run database migrations

```bash
make migrate-up
```

#### 7. Seed development data (optional)

```bash
./scripts/seed-data.sh
```

#### 8. Start the API server

```bash
make run
```

The API will be available at http://localhost:8080

**Swagger UI**: http://localhost:8080/swagger/index.html

## Development

### Running with Docker Compose

```bash
cd deployments/docker
docker-compose up
```

This will start:
- PostgreSQL database
- Redis cache
- API server
- Background worker

### Running tests

```bash
./scripts/test.sh

# With coverage report
./scripts/test.sh --html
```

### Building binaries

```bash
./scripts/build.sh
```

Binaries will be created in the `bin/` directory.

### Database Migrations

Create a new migration:

```bash
go run cmd/migrate/main.go create <migration_name>
```

Run migrations:

```bash
go run cmd/migrate/main.go up
```

Rollback migrations:

```bash
go run cmd/migrate/main.go down
```

## Project Components

### Domain Layer

Contains business entities and core business logic:
- **Book**: Book management, borrowing, reservations
- **User**: User accounts and profiles
- **Author**: Author information
- **Category**: Book categorization
- **Review**: Book reviews and ratings
- **Fine**: Overdue fines management

### Use Case Layer

Implements application-specific business rules and orchestrates domain entities.

### Adapters Layer

External service integrations:
- **gRPC**: gRPC server for service-to-service communication
- **Email**: SMTP email sending
- **Payment**: Stripe and PayPal integration
- **Storage**: S3 and local file storage

### Infrastructure Layer

Technical concerns:
- **Auth**: JWT authentication and authorization
- **Config**: Environment configuration
- **Database**: PostgreSQL connection and queries
- **HTTP**: HTTP server, middleware, and routing
- **Logger**: Structured logging with Zap

### Shared Packages

Reusable utilities:
- **Validator**: Input validation with custom rules
- **Pagination**: Cursor and offset pagination
- **Crypto**: Password hashing and token generation
- **Timeutil**: Time manipulation helpers

## API Documentation

API documentation is available via Swagger UI when running the server:

http://localhost:8080/swagger/

## Deployment

### Docker

```bash
cd deployments/docker
docker-compose up -d
```

### Kubernetes

```bash
kubectl apply -f deployments/kubernetes/
```

### Terraform

Infrastructure as Code configurations are in `deployments/terraform/`.

## Testing

The project includes:
- Unit tests for domain logic
- Integration tests for use cases
- API tests for HTTP endpoints

Run all tests:

```bash
go test ./...
```

## Documentation

### 📖 Getting Started
- **[Quick Start (5 min)](./docs/guides/QUICKSTART.md)** - Get up and running in 5 minutes
- **[Development Guide](./docs/guides/DEVELOPMENT.md)** - Comprehensive development workflow
- **[Contributing Guidelines](./docs/guides/CONTRIBUTING.md)** - How to contribute

### 🏗️ Architecture
- **[Architecture Overview](./docs/architecture.md)** - System architecture and design principles
- **[Package Overview](./docs/package-overview.md)** - Package structure and dependencies
- **[Architecture Decisions (ADRs)](./docs/adr/README.md)** - Key architectural decisions

### 💻 Code Examples
- **[Basic CRUD](./examples/basic_crud/)** - Complete CRUD workflow example
- **[Domain Services](./examples/domain_service/)** - Business logic patterns
- **[Testing Patterns](./examples/testing/)** - Testing strategies and mocks

### 📚 Layer Documentation
- **[Domain Layer](./internal/domain/README.md)** - Business logic and entities
- **[Use Case Layer](./internal/usecase/README.md)** - Application use cases
- **[Adapter Layer](./internal/adapters/README.md)** - External interfaces
- **[Command Line Apps](./cmd/README.md)** - Entry points (API, worker, migrate)
- **[Shared Packages](./pkg/README.md)** - Reusable utilities

### 🧪 Testing
- **[Test Fixtures](./test/fixtures/README.md)** - Shared test data
- **[Integration Tests](./test/integration/)** - Database integration tests
- **Benchmarks** - Performance benchmarks in `*_benchmark_test.go`

### 📡 API
- **[OpenAPI Specification](./api/openapi/swagger.yaml)** - API schema
- **Swagger UI**: http://localhost:8080/swagger/index.html (when running)

## Contributing

Please read our [Contributing Guidelines](./docs/guides/CONTRIBUTING.md) for:
- Code style and standards
- Testing requirements
- Pull request process
- Commit message format

**Quick Start for Contributors**:
```bash
# Automated setup (recommended)
./scripts/dev-setup.sh

# OR manual setup
make init && make up && make migrate-up
make install-hooks  # Install pre-commit hooks

# Development workflow
make test           # Run tests
make lint           # Check code quality
make ci             # Full CI pipeline: fmt → vet → lint → test → build

# Pre-commit hooks run automatically on git commit
# They ensure: formatting, vet checks, unit tests, no log files
```

**Pre-commit Hooks**:
Git hooks are automatically installed by the dev-setup script or via `make install-hooks`. They run quality checks before each commit to catch issues early.

See also:
- [Architecture Decisions](./docs/adr/README.md) - Understanding design choices
- [Development Guide](./docs/guides/DEVELOPMENT.md) - Development workflow

## License

[Add your license here]

## Support

For issues and questions, please open an issue on GitHub.

---

**Built with Clean Architecture** | [ADR-001](./docs/adr/001-clean-architecture.md) | [ADR-002](./docs/adr/002-domain-services.md) | [ADR-003](./docs/adr/003-dependency-injection.md)
