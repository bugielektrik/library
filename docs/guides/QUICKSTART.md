# Quick Start Guide

Get the Library Management System running in 5 minutes.

## Prerequisites

- Go 1.25+
- Docker & Docker Compose
- Make (optional, for shortcuts)

## Quick Start

### 1. Clone and Setup (1 min)

```bash
git clone <repository-url>
cd library
go mod download
```

### 2. Start Services (2 min)

```bash
# Start PostgreSQL and Redis
make up
# or: cd deployments/docker && docker-compose up -d

# Run migrations
make migrate-up
# or: go run ./cmd/migrate up
```

### 3. Start API Server (1 min)

```bash
make run
# or: go run ./cmd/api
```

Server starts at `http://localhost:8080`

### 4. Test It Works (1 min)

```bash
# Create an author
curl -X POST http://localhost:8080/authors \
  -H "Content-Type: application/json" \
  -d '{"fullName": "J.K. Rowling", "specialty": "Fantasy"}'

# List authors
curl http://localhost:8080/authors
```

## API Documentation

Swagger UI: `http://localhost:8080/swagger/index.html`

## Development Commands

```bash
make help              # Show all available commands
make test              # Run tests
make build             # Build binaries
make lint              # Run linters
make dev               # Start dev environment (services + migrations + api)
```

## Project Structure

```
library/
├── cmd/               # Entry points (api, worker, migrate)
├── internal/
│   ├── domain/       # Business logic & entities
│   ├── usecase/      # Application use cases
│   └── adapters/     # External interfaces (HTTP, DB)
├── pkg/              # Shared utilities
└── docs/             # Documentation
```

## Next Steps

1. Read [Architecture Overview](../architecture.md)
2. Check [Development Guide](./DEVELOPMENT.md)
3. Review [Contributing Guidelines](./CONTRIBUTING.md)
4. Explore API endpoints at `/swagger/index.html`

## Common Issues

**Port already in use**: Change port in `.env` or stop conflicting service
**Database connection failed**: Ensure Docker services are running (`make up`)
**Tests failing**: Run `make migrate-up` to ensure test database schema is ready

## Quick Reference

| Command | Description |
|---------|-------------|
| `make dev` | Full dev environment (recommended) |
| `make test` | Run all tests |
| `make build` | Build all binaries |
| `make clean` | Clean artifacts |
| `make down` | Stop all services |
