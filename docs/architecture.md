# Library Management System - Architecture Documentation

## Overview

This document describes the architecture of the Library Management System, a Go-based REST API service following Clean Architecture principles optimized for vibecoding with Claude Code.

## Architecture Pattern: Clean Architecture (Onion Architecture)

The system follows Uncle Bob's Clean Architecture with clear separation of concerns across four layers:

```
┌─────────────────────────────────────────────────────────────┐
│                     External Systems                        │
│  (HTTP Clients, Browsers, gRPC Clients, CLI Tools)         │
└─────────────────────────────────────────────────────────────┘
                              ↕
┌─────────────────────────────────────────────────────────────┐
│                    Adapters Layer                           │
│  (HTTP Handlers, gRPC Services, CLI Commands)              │
│  (Repository Implementations, External Service Adapters)     │
└─────────────────────────────────────────────────────────────┘
                              ↕
┌─────────────────────────────────────────────────────────────┐
│                    Use Case Layer                           │
│  (Application Business Rules, Orchestration)                │
│  (CreateBook, SubscribeMember, etc.)                        │
└─────────────────────────────────────────────────────────────┘
                              ↕
┌─────────────────────────────────────────────────────────────┐
│                    Domain Layer                             │
│  (Entities, Value Objects, Domain Services)                 │
│  (Book, Member, Author, Business Rules)                     │
└─────────────────────────────────────────────────────────────┘
```

## Layer Responsibilities

### 1. Domain Layer (`internal/domain/`)
**Purpose:** Core business logic with zero external dependencies

- **Entities:** Core business objects (Book, Member, Author)
- **Value Objects:** Immutable domain concepts (ISBN, SubscriptionType)
- **Domain Services:** Business rules that don't belong to a single entity
- **Repository Interfaces:** Contracts for data persistence
- **Domain Events:** Business events (future implementation)

**Example - Book Domain Service:**
```go
// ValidateISBN validates both ISBN-10 and ISBN-13 formats
func (s *Service) ValidateISBN(isbn string) error
// Business rule: Books cannot be deleted if they have active loans
func (s *Service) CanBookBeDeleted(book Entity) error
```

### 2. Use Case Layer (`internal/usecase/`)
**Purpose:** Application-specific business rules and orchestration

- **Use Cases:** Single business operations (CreateBook, SubscribeMember)
- **Orchestration:** Coordinates domain entities and services
- **Transaction Boundaries:** Defines unit of work
- **External Service Interfaces:** Contracts for external services

**Example - Create Book Use Case:**
```go
1. Validate input using domain service
2. Check for duplicate ISBN
3. Create book entity
4. Persist to repository
5. Update cache
6. Return response
```

### 3. Adapters Layer (`internal/adapters/`)
**Purpose:** Translates between external world and use cases

#### Inbound Adapters (Driving):
- **HTTP Handlers:** REST API endpoints
- **gRPC Services:** RPC endpoints
- **CLI Commands:** Command-line interface

#### Outbound Adapters (Driven):
- **Repository Implementations:** PostgreSQL, MongoDB, Memory
- **External Services:** Email (SMTP), Payment (Stripe/PayPal), Storage (S3)
- **Cache Implementations:** Redis, In-memory

### 4. Infrastructure Layer (`internal/infrastructure/`)
**Purpose:** Technical concerns and cross-cutting functionality

- **Configuration:** Environment and application config
- **Logging:** Structured logging with Zap
- **Database:** Connection pooling and management
- **Authentication:** JWT token management
- **Server:** HTTP server configuration
- **Monitoring:** Metrics and health checks

## Dependency Rules

The fundamental rule: **Dependencies only point inward**

```
Domain ← Use Cases ← Adapters ← Infrastructure
```

- Domain layer has NO dependencies on outer layers
- Use Cases depend only on Domain
- Adapters depend on Use Cases and Domain
- Infrastructure can depend on any layer

## Data Flow Example: Creating a Book

```
1. HTTP Request → BookHandler (Adapter)
   POST /api/v1/books
   {
     "name": "Clean Architecture",
     "isbn": "978-0134494166",
     "authors": ["author-id-1"]
   }

2. BookHandler → CreateBookUseCase (Use Case)
   - Parses request into CreateBookRequest
   - Calls use case Execute method

3. CreateBookUseCase → BookService (Domain)
   - Validates book using domain service
   - Checks business rules

4. CreateBookUseCase → BookRepository (Adapter)
   - Persists book to database
   - Updates cache

5. Response Flow (reverse)
   Repository → UseCase → Handler → HTTP Response
```

## Key Design Decisions

### 1. Domain Services vs Anemic Domain Model
**Decision:** Use domain services for complex business logic
**Rationale:** Keeps entities focused while encapsulating business rules

### 2. Repository Pattern
**Decision:** Repository interfaces in domain, implementations in adapters
**Rationale:** Domain remains independent of persistence technology

### 3. Use Case per Operation
**Decision:** One use case class per business operation
**Rationale:** Single Responsibility Principle, easier testing

### 4. DTO Pattern
**Decision:** Separate DTOs for each layer
**Rationale:** Prevents coupling between layers

### 5. Dependency Injection
**Decision:** Constructor injection everywhere
**Rationale:** Explicit dependencies, easier testing

## Testing Strategy

### Unit Tests
- **Domain Services:** Pure logic tests, no mocks needed
- **Use Cases:** Mock repositories and external services
- **Handlers:** Mock use cases

### Integration Tests
- **Repository Tests:** Test against real database
- **API Tests:** Full HTTP request/response cycle

### End-to-End Tests
- **User Journeys:** Complete workflows through the system

## Performance Considerations

### Caching Strategy
- **Read-Through Cache:** Check cache, fallback to database
- **Write-Through Cache:** Update cache on writes
- **TTL:** 5 minutes for frequently accessed data

### Database Optimization
- **Connection Pooling:** Max 25 connections
- **Prepared Statements:** Reuse query plans
- **Indexes:** On frequently queried fields

### Concurrent Processing
- **Worker Pool:** Background job processing
- **Rate Limiting:** Per-user and per-IP limits
- **Circuit Breaker:** For external service calls

## Security Architecture

### Authentication & Authorization
- **JWT Tokens:** Stateless authentication
- **RBAC:** Role-based access control
- **Token Refresh:** Automatic token renewal

### Input Validation
- **Domain Validation:** Business rule validation
- **Input Sanitization:** XSS and SQL injection prevention
- **Request Size Limits:** Prevent DoS attacks

### Secrets Management
- **Environment Variables:** For sensitive configuration
- **Never in Code:** No hardcoded credentials
- **Rotation:** Regular credential rotation

## Deployment Architecture

### Container Strategy
```
┌─────────────┐ ┌─────────────┐ ┌─────────────┐
│  API Server │ │   Worker    │ │  Migration  │
│  (Port 8080)│ │  (Async)    │ │   Tool      │
└─────────────┘ └─────────────┘ └─────────────┘
       ↕              ↕                ↕
┌─────────────────────────────────────────────┐
│            PostgreSQL Database              │
└─────────────────────────────────────────────┘
       ↕              ↕                ↕
┌─────────────────────────────────────────────┐
│             Redis Cache                     │
└─────────────────────────────────────────────┘
```

### Scaling Strategy
- **Horizontal Scaling:** API servers behind load balancer
- **Database Replication:** Read replicas for queries
- **Cache Distribution:** Redis cluster for high availability

## Monitoring & Observability

### Metrics
- **Application Metrics:** Request rate, error rate, duration
- **Business Metrics:** Books created, members subscribed
- **Infrastructure Metrics:** CPU, memory, disk usage

### Logging
- **Structured Logging:** JSON format with Zap
- **Log Levels:** Debug, Info, Warn, Error
- **Correlation IDs:** Trace requests across services

### Health Checks
- **Liveness:** Is the service running?
- **Readiness:** Is the service ready to accept traffic?
- **Dependencies:** Are external services accessible?

## Development Workflow

### Local Development
```bash
# Start dependencies
docker-compose up -d postgres redis

# Run migrations
go run cmd/migrate/main.go up

# Start API server
go run cmd/api/main.go

# Run tests
go test ./...
```

### Code Organization
- **Feature-based:** Group by domain concept
- **Screaming Architecture:** Folders reveal business intent
- **Consistent Naming:** Predictable file locations

### Continuous Integration
1. Lint with golangci-lint
2. Run unit tests
3. Run integration tests
4. Build Docker images
5. Deploy to staging

## Future Enhancements

### Planned Features
- [ ] Event Sourcing for audit trail
- [ ] CQRS for read/write separation
- [ ] GraphQL API alongside REST
- [ ] WebSocket support for real-time updates
- [ ] Multi-tenancy support

### Technical Improvements
- [ ] OpenTelemetry for distributed tracing
- [ ] Feature flags for gradual rollouts
- [ ] API versioning strategy
- [ ] Database sharding for scale
- [ ] Service mesh integration

## Conclusion

This architecture provides:
- **Maintainability:** Clear separation of concerns
- **Testability:** Each layer independently testable
- **Flexibility:** Easy to change implementations
- **Scalability:** Horizontal scaling ready
- **Developer Experience:** Intuitive structure for vibecoding

The clean architecture ensures the business logic remains independent of technical details, making the system resilient to change and optimal for development with Claude Code.