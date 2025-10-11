# Shared Packages

**Reusable utilities and helpers used across the application.**

## Purpose

This directory contains:
- **Common Utilities**: Shared helper functions
- **Error Handling**: Custom error types and utilities
- **Validation**: Input validation helpers
- **Constants**: Application-wide constants
- **Types**: Shared type definitions

## Dependency Rule

Package code should be:
- **Standalone**: No dependencies on internal packages
- **Reusable**: Can be used in any layer
- **Well-tested**: High test coverage
- **Documented**: Clear godoc comments

```
pkg/ (this directory)
  ✗ NO dependency on internal/
  ✓ Can depend on external libraries
  ✓ Used by all layers
```

## Directory Structure

```
pkg/
├── crypto/           # Password hashing and cryptographic utilities
│   └── crypto.go
│
├── errors/           # Error handling
│   ├── errors.go     # Custom error types
│   └── domain.go     # Domain-specific errors
│
├── httputil/         # HTTP utilities
│   ├── status.go     # HTTP status code constants and helpers
│   ├── status_test.go
│   └── doc.go
│
├── logutil/          # Logger utilities
│   ├── logger.go     # Logger initialization helpers
│   ├── logger_test.go
│   └── doc.go
│
├── pagination/       # Pagination utilities
│   └── pagination.go
│
├── strutil/          # String utilities
│   ├── string.go     # Safe string pointer helpers
│   ├── string_test.go
│   └── doc.go
│
├── timeutil/         # Time-related utilities
│   └── time.go
│
└── validator/        # Validation utilities
    └── validator.go
```

## Error Package

### Custom Error Types

```go
// pkg/errors/errors.go
type Error struct {
    Code    string
    Message string
    Err     error  // Wrapped error
}

func (e *Error) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("%s: %v", e.Message, e.Err)
    }
    return e.Message
}

func (e *Error) Unwrap() error {
    return e.Err
}

func (e *Error) Is(target error) bool {
    t, ok := target.(*Error)
    if !ok {
        return false
    }
    return e.Code == t.Code
}

// Constructor
func New(code, message string) *Error {
    return &Error{
        Code:    code,
        Message: message,
    }
}

func Wrap(err error, code, message string) *Error {
    return &Error{
        Code:    code,
        Message: message,
        Err:     err,
    }
}
```

### Domain Errors

```go
// pkg/errors/domain.go
var (
    ErrNotFound      = New("NOT_FOUND", "Resource not found")
    ErrInvalidInput  = New("INVALID_INPUT", "Invalid input data")
    ErrUnauthorized  = New("UNAUTHORIZED", "Unauthorized access")
    ErrAlreadyExists = New("ALREADY_EXISTS", "Resource already exists")
    ErrInternal      = New("INTERNAL_ERROR", "Internal server error")
)

// Usage in domain
if book == nil {
    return nil, errors.ErrNotFound
}

// Check error type
if errors.Is(err, errors.ErrNotFound) {
    // Handle not found
}
```

## Validator Package

### Common Validators

```go
// pkg/validator/validator.go
type Validator struct{}

func (v *Validator) ValidateEmail(email string) error {
    pattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
    if !regexp.MustCompile(pattern).MatchString(email) {
        return errors.New("invalid email format")
    }
    return nil
}

func (v *Validator) ValidateUUID(id string) error {
    if _, err := uuid.Parse(id); err != nil {
        return errors.New("invalid UUID format")
    }
    return nil
}

func (v *Validator) ValidateRequired(value string, fieldName string) error {
    if strings.TrimSpace(value) == "" {
        return fmt.Errorf("%s is required", fieldName)
    }
    return nil
}

func (v *Validator) ValidateLength(value string, min, max int) error {
    length := len(value)
    if length < min || length > max {
        return fmt.Errorf("length must be between %d and %d", min, max)
    }
    return nil
}
```

### ISBN Validator

```go
// pkg/validator/isbn.go
type ISBNValidator struct{}

func (v *ISBNValidator) ValidateISBN10(isbn string) bool {
    // Remove hyphens
    isbn = strings.ReplaceAll(isbn, "-", "")

    if len(isbn) != 10 {
        return false
    }

    // Checksum validation
    sum := 0
    for i := 0; i < 9; i++ {
        digit := int(isbn[i] - '0')
        sum += digit * (10 - i)
    }

    lastChar := isbn[9]
    if lastChar == 'X' {
        sum += 10
    } else {
        sum += int(lastChar - '0')
    }

    return sum%11 == 0
}
```

## Logger Package

### Structured Logging

```go
// pkg/log/log.go
type Logger struct {
    level  Level
    output io.Writer
}

type Level int

const (
    DEBUG Level = iota
    INFO
    WARN
    ERROR
)

func New(level Level) *Logger {
    return &Logger{
        level:  level,
        output: os.Stdout,
    }
}

func (l *Logger) Info(message string, fields ...Field) {
    if l.level <= INFO {
        l.log(INFO, message, fields...)
    }
}

func (l *Logger) Error(message string, err error, fields ...Field) {
    if l.level <= ERROR {
        fields = append(fields, Field{Key: "error", Value: err.Error()})
        l.log(ERROR, message, fields...)
    }
}

type Field struct {
    Key   string
    Value interface{}
}

// Usage
logger := logger.New(logger.INFO)
logger.Info("Book created",
    logger.Field{Key: "book_id", Value: "123"},
    logger.Field{Key: "title", Value: "Clean Code"},
)
```

## Logger Utilities Package

### Simplified Logger Initialization

The `logutil` package provides helper functions to reduce boilerplate when initializing loggers across different architectural layers:

```go
// pkg/logutil/logger.go

// Use Case Logger - for orchestration layer
func UseCaseLogger(ctx context.Context, useCaseName string, fields ...zap.Field) *zap.Logger

// Handler Logger - for HTTP handlers
func HandlerLogger(ctx context.Context, handlerName, methodName string) *zap.Logger

// Repository Logger - for data access layer
func RepositoryLogger(ctx context.Context, repositoryName, operation string) *zap.Logger

// Gateway Logger - for external service integrations
func GatewayLogger(ctx context.Context, gatewayName, operation string) *zap.Logger
```

### Usage Examples

**Before (3 lines):**
```go
logger := log.FromContext(ctx).Named("create_book_usecase").With(
    zap.String("isbn", req.ISBN),
)
```

**After (1 line):**
```go
logger := logutil.UseCaseLogger(ctx, "create_book", zap.String("isbn", req.ISBN))
```

**Use Case Layer:**
```go
import "library-service/pkg/logutil"

func (uc *CreateBookUseCase) Execute(ctx context.Context, req Request) (Response, error) {
    logger := logutil.UseCaseLogger(ctx, "create_book",
        zap.String("isbn", req.ISBN),
        zap.String("title", req.Title),
    )
    logger.Info("creating book")
    // ...
}
```

**HTTP Handler Layer:**
```go
import "library-service/pkg/logutil"

func (h *BookHandler) create(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    logger := logutil.HandlerLogger(ctx, "book_handler", "create")
    logger.Info("handling create book request")
    // ...
}
```

**Repository Layer:**
```go
import "library-service/pkg/logutil"

func (r *BookRepository) Create(ctx context.Context, book Book) error {
    logger := logutil.RepositoryLogger(ctx, "book", "create")
    logger.Info("creating book in database", zap.String("id", book.ID))
    // ...
}
```

**Gateway Layer:**
```go
import "library-service/pkg/logutil"

func (g *PaymentGateway) InitiatePayment(ctx context.Context, req PaymentRequest) error {
    logger := logutil.GatewayLogger(ctx, "epayment", "initiate_payment")
    logger.Info("initiating payment", zap.Int64("amount", req.Amount))
    // ...
}
```

### Benefits

- **Reduces Boilerplate**: 3-line logger initialization reduced to 1 line
- **Consistent Naming**: Automatic naming convention (e.g., "create_book_usecase")
- **Context Propagation**: Automatically extracts logger from context
- **Layer Clarity**: Different helpers for different architectural layers
- **Easy Refactoring**: Change logging strategy in one place

## Config Package

### Configuration Loading

```go
// pkg/config/config.go
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Redis    RedisConfig
}

type ServerConfig struct {
    Port         int
    Host         string
    ReadTimeout  time.Duration
    WriteTimeout time.Duration
}

type DatabaseConfig struct {
    Host     string
    Port     int
    Name     string
    User     string
    Password string
    SSLMode  string
}

func Load() (*Config, error) {
    // Load from environment variables
    cfg := &Config{
        Server: ServerConfig{
            Port:         getEnvInt("SERVER_PORT", 8080),
            Host:         getEnv("SERVER_HOST", "0.0.0.0"),
            ReadTimeout:  getEnvDuration("READ_TIMEOUT", 30*time.Second),
            WriteTimeout: getEnvDuration("WRITE_TIMEOUT", 30*time.Second),
        },
        Database: DatabaseConfig{
            Host:     getEnv("DB_HOST", "localhost"),
            Port:     getEnvInt("DB_PORT", 5432),
            Name:     getEnv("DB_NAME", "library"),
            User:     getEnv("DB_USER", "library"),
            Password: getEnv("DB_PASSWORD", ""),
            SSLMode:  getEnv("DB_SSL_MODE", "disable"),
        },
    }

    return cfg, cfg.Validate()
}

func (c *Config) Validate() error {
    if c.Server.Port < 1 || c.Server.Port > 65535 {
        return errors.New("invalid server port")
    }
    if c.Database.Name == "" {
        return errors.New("store name is required")
    }
    return nil
}
```

## Types Package

### Pagination

```go
// pkg/types/pagination.go
type PaginationParams struct {
    Page     int `json:"page" form:"page"`
    PageSize int `json:"page_size" form:"page_size"`
}

func (p *PaginationParams) Validate() error {
    if p.Page < 1 {
        p.Page = 1
    }
    if p.PageSize < 1 {
        p.PageSize = 10
    }
    if p.PageSize > 100 {
        p.PageSize = 100
    }
    return nil
}

func (p *PaginationParams) Offset() int {
    return (p.Page - 1) * p.PageSize
}

func (p *PaginationParams) Limit() int {
    return p.PageSize
}

type PaginatedResponse struct {
    Data       interface{} `json:"data"`
    Page       int         `json:"page"`
    PageSize   int         `json:"page_size"`
    TotalItems int64       `json:"total_items"`
    TotalPages int         `json:"total_pages"`
}

func NewPaginatedResponse(data interface{}, page, pageSize int, totalItems int64) *PaginatedResponse {
    totalPages := int(math.Ceil(float64(totalItems) / float64(pageSize)))
    return &PaginatedResponse{
        Data:       data,
        Page:       page,
        PageSize:   pageSize,
        TotalItems: totalItems,
        TotalPages: totalPages,
    }
}
```

## Testing

All pkg code should be thoroughly tested:

```bash
# Test all packages
go test ./pkg/... -cover

# Test specific package
go test ./pkg/errors -v
go test ./pkg/validator -v
```

## Best Practices

1. **No Internal Dependencies**: pkg/ should never import from internal/
2. **Backward Compatibility**: Changes should be backward compatible
3. **Documentation**: All exported functions must have godoc comments
4. **Test Coverage**: Aim for 90%+ coverage on pkg code
5. **Small Packages**: Keep packages focused and cohesive
6. **Error Handling**: Return errors, don't panic
7. **Immutability**: Prefer immutable types where possible

## Adding New Package

```bash
# 1. Create directory
mkdir pkg/newpackage

# 2. Create main file
touch pkg/newpackage/newpackage.go

// Package newpackage provides utilities for X.
package newpackage

# 3. Write tests
touch pkg/newpackage/newpackage_test.go

# 4. Update this README
```

## Usage Examples

### Error Handling

```go
import "library-service/pkg/errors"

// Create error
err := errors.New("INVALID_ISBN", "ISBN format is invalid")

// Wrap error
if err := repo.Create(ctx, book); err != nil {
    return errors.Wrap(err, "DB_ERROR", "failed to create book")
}

// Check error type
if errors.Is(err, errors.ErrNotFound) {
    return http.StatusNotFound
}
```

### Validation

```go
import "library-service/pkg/validator"

v := validator.Validator{}

// Validate email
if err := v.ValidateEmail(user.Email); err != nil {
    return err
}

// Validate required
if err := v.ValidateRequired(book.Name, "book name"); err != nil {
    return err
}
```

### Logging

```go
import "library-service/pkg/log"

log := logger.New(logger.INFO)

log.Info("Processing request",
    logger.Field{Key: "user_id", Value: userID},
    logger.Field{Key: "action", Value: "create_book"},
)

if err != nil {
    log.Error("Failed to create book", err,
        logger.Field{Key: "book_id", Value: bookID},
    )
}
```

## References

- [Internal Packages](../internal/README.md)
- [Development Guide](../docs/guides/DEVELOPMENT.md)
- [Go Package Best Practices](https://golang.org/doc/effective_go#package-names)
