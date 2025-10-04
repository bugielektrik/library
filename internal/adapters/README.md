# Adapters Layer

**External interfaces and infrastructure implementations - the boundary of the application.**

## Purpose

This layer contains:
- **HTTP Handlers**: REST API endpoints (Gin framework)
- **Repository Implementations**: Database access (PostgreSQL)
- **Cache Implementations**: Redis caching
- **External Service Adapters**: Third-party integrations
- **Mappers**: DTO ↔ Domain entity conversions

## Dependency Rule

Adapters depend on **domain interfaces** (inward dependency via Dependency Inversion).

```
Adapters (this layer)
  ↓ depends on
Domain Interfaces (contracts)

  ↓ implements
Repository, Cache, etc.
```

## Directory Structure

```
adapters/
├── http/                    # HTTP handlers
│   ├── book/               # Book endpoints
│   │   ├── handler.go      # HTTP handlers
│   │   ├── dto.go          # HTTP request/response DTOs
│   │   └── mapper.go       # DTO ↔ UseCase mapping
│   ├── member/             # Member endpoints
│   ├── author/             # Author endpoints
│   └── middleware/         # HTTP middlewares
│
├── repository/             # Database implementations
│   ├── book_postgres.go    # Book repository (PostgreSQL)
│   ├── member_postgres.go  # Member repository
│   └── author_postgres.go  # Author repository
│
├── cache/                  # Cache implementations
│   ├── book_redis.go       # Book cache (Redis)
│   └── author_redis.go     # Author cache
│
└── storage/                # File/object storage
    └── s3_storage.go       # S3 storage adapter
```

## HTTP Layer

### Handler Pattern

```go
type BookHandler struct {
    createBookUC *usecase.CreateBookUseCase
    updateBookUC *usecase.UpdateBookUseCase
    listBooksUC  *usecase.ListBooksUseCase
    // ... other use cases
}

func NewBookHandler(
    createUC *usecase.CreateBookUseCase,
    updateUC *usecase.UpdateBookUseCase,
    listUC *usecase.ListBooksUseCase,
) *BookHandler {
    return &BookHandler{
        createBookUC: createUC,
        updateBookUC: updateUC,
        listBooksUC:  listUC,
    }
}

// Handler method
func (h *BookHandler) CreateBook(c *gin.Context) {
    // 1. Parse request
    var req CreateBookRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    // 2. Map to use case input
    input := req.ToUseCaseInput()

    // 3. Execute use case
    book, err := h.createBookUC.Execute(c.Request.Context(), input)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    // 4. Map to response
    response := MapToBookResponse(book)
    c.JSON(200, response)
}
```

### HTTP DTOs

```go
// Request DTO
type CreateBookRequest struct {
    Name    string   `json:"name" binding:"required"`
    ISBN    string   `json:"isbn" binding:"required"`
    Authors []string `json:"authors" binding:"required"`
    Genre   string   `json:"genre"`
}

func (req CreateBookRequest) ToUseCaseInput() usecase.CreateBookInput {
    return usecase.CreateBookInput{
        Name:    req.Name,
        ISBN:    req.ISBN,
        Authors: req.Authors,
        Genre:   req.Genre,
    }
}

// Response DTO
type BookResponse struct {
    ID      string   `json:"id"`
    Name    string   `json:"name"`
    ISBN    string   `json:"isbn"`
    Authors []string `json:"authors"`
    Genre   string   `json:"genre"`
}

func MapToBookResponse(entity *book.Entity) BookResponse {
    return BookResponse{
        ID:      entity.ID,
        Name:    entity.Name,
        ISBN:    entity.ISBN,
        Authors: entity.Authors,
        Genre:   entity.Genre,
    }
}
```

### Route Registration

```go
// cmd/api/main.go or routes.go
func setupRoutes(router *gin.Engine, handler *http.BookHandler) {
    v1 := router.Group("/api/v1")
    {
        books := v1.Group("/books")
        {
            books.POST("", handler.CreateBook)
            books.GET("", handler.ListBooks)
            books.GET("/:id", handler.GetBook)
            books.PUT("/:id", handler.UpdateBook)
            books.DELETE("/:id", handler.DeleteBook)
        }
    }
}
```

## Repository Layer

### Repository Implementation

```go
type PostgresBookRepository struct {
    db *sqlx.DB
}

func NewPostgresBookRepository(db *sqlx.DB) book.Repository {
    return &PostgresBookRepository{db: db}
}

// Implements domain.Repository interface
func (r *PostgresBookRepository) Create(ctx context.Context, book book.Entity) error {
    query := `
        INSERT INTO books (id, name, isbn, authors, genre, created_at)
        VALUES ($1, $2, $3, $4, $5, $6)
    `
    _, err := r.db.ExecContext(ctx, query,
        book.ID,
        book.Name,
        book.ISBN,
        pq.Array(book.Authors),
        book.Genre,
        time.Now(),
    )
    if err != nil {
        return fmt.Errorf("failed to insert book: %w", err)
    }
    return nil
}

func (r *PostgresBookRepository) GetByID(ctx context.Context, id string) (*book.Entity, error) {
    var entity book.Entity
    query := `SELECT id, name, isbn, authors, genre FROM books WHERE id = $1`

    err := r.db.GetContext(ctx, &entity, query, id)
    if err == sql.ErrNoRows {
        return nil, errors.New("book not found")
    }
    if err != nil {
        return nil, fmt.Errorf("failed to get book: %w", err)
    }

    return &entity, nil
}
```

### Database Scanning

```go
// For complex types, use custom scanning
type dbBook struct {
    ID      string         `db:"id"`
    Name    string         `db:"name"`
    ISBN    string         `db:"isbn"`
    Authors pq.StringArray `db:"authors"`  // PostgreSQL array
    Genre   string         `db:"genre"`
}

func (db *dbBook) ToEntity() book.Entity {
    return book.Entity{
        ID:      db.ID,
        Name:    db.Name,
        ISBN:    db.ISBN,
        Authors: []string(db.Authors),
        Genre:   db.Genre,
    }
}
```

## Cache Layer

### Cache Implementation

```go
type RedisBookCache struct {
    client *redis.Client
    ttl    time.Duration
}

func NewRedisBookCache(client *redis.Client) book.Cache {
    return &RedisBookCache{
        client: client,
        ttl:    1 * time.Hour,
    }
}

// Implements domain.Cache interface
func (c *RedisBookCache) Set(ctx context.Context, book book.Entity) error {
    key := fmt.Sprintf("book:%s", book.ID)
    data, err := json.Marshal(book)
    if err != nil {
        return err
    }

    return c.client.Set(ctx, key, data, c.ttl).Err()
}

func (c *RedisBookCache) Get(ctx context.Context, id string) (*book.Entity, error) {
    key := fmt.Sprintf("book:%s", id)
    data, err := c.client.Get(ctx, key).Bytes()
    if err == redis.Nil {
        return nil, errors.New("cache miss")
    }
    if err != nil {
        return nil, err
    }

    var entity book.Entity
    if err := json.Unmarshal(data, &entity); err != nil {
        return nil, err
    }

    return &entity, nil
}

func (c *RedisBookCache) Delete(ctx context.Context, id string) error {
    key := fmt.Sprintf("book:%s", id)
    return c.client.Del(ctx, key).Err()
}
```

## Middleware

### Common Middlewares

```go
// Error handling middleware
func ErrorHandler() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()

        if len(c.Errors) > 0 {
            err := c.Errors.Last()
            c.JSON(500, gin.H{"error": err.Error()})
        }
    }
}

// Logging middleware
func RequestLogger() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        path := c.Request.URL.Path

        c.Next()

        latency := time.Since(start)
        log.Printf("[%s] %s - %v", c.Request.Method, path, latency)
    }
}

// CORS middleware
func CORS() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
        c.Next()
    }
}
```

## Testing

### HTTP Handler Testing

```go
func TestBookHandler_CreateBook(t *testing.T) {
    // Setup
    mockUseCase := &MockCreateBookUseCase{}
    handler := NewBookHandler(mockUseCase, nil, nil)

    router := gin.Default()
    router.POST("/books", handler.CreateBook)

    // Prepare request
    reqBody := `{"name":"Test Book","isbn":"978-0-306-40615-7","authors":["Author"]}`
    req := httptest.NewRequest("POST", "/books", strings.NewReader(reqBody))
    req.Header.Set("Content-Type", "application/json")

    w := httptest.NewRecorder()

    // Mock use case
    mockUseCase.On("Execute", mock.Anything, mock.Anything).Return(&book.Entity{
        ID:   "123",
        Name: "Test Book",
    }, nil)

    // Execute
    router.ServeHTTP(w, req)

    // Assert
    assert.Equal(t, 200, w.Code)
    mockUseCase.AssertExpectations(t)
}
```

## Best Practices

1. **Thin Handlers**: Minimal logic, delegate to use cases
2. **DTO Mapping**: Never expose domain entities directly
3. **Error Handling**: Proper HTTP status codes
4. **Validation**: Use Gin binding tags (`binding:"required"`)
5. **Context Propagation**: Pass `context.Context` to use cases
6. **Database Transactions**: Use repository pattern for transaction boundaries
7. **Cache Invalidation**: Invalidate on write operations
8. **Logging**: Log at adapter boundaries

## Error Handling

### HTTP Status Code Mapping

```go
func mapErrorToHTTPStatus(err error) int {
    switch {
    case errors.Is(err, domain.ErrNotFound):
        return http.StatusNotFound
    case errors.Is(err, domain.ErrInvalidInput):
        return http.StatusBadRequest
    case errors.Is(err, domain.ErrUnauthorized):
        return http.StatusUnauthorized
    default:
        return http.StatusInternalServerError
    }
}
```

## Performance

### Caching Strategy

- **Read-Through**: Check cache → DB → Update cache
- **Write-Through**: Write to DB → Update cache
- **Cache-Aside**: Application manages cache explicitly

### Database Optimization

- Use prepared statements
- Index frequently queried columns
- Batch operations where possible
- Connection pooling (configured in main.go)

## References

- [Domain Layer](../domain/README.md)
- [Use Case Layer](../usecase/README.md)
- [API Documentation](../../api/openapi/swagger.yaml)
