# Security Guide

> **Security best practices and common vulnerabilities to avoid**

## Purpose

Security checklist for AI-assisted development. Claude Code should check these before suggesting code changes.

**Principle:** Security must be baked in, not bolted on.

---

## 🔐 Security Checklist (Quick Reference)

Before committing code, verify:

- [ ] No hardcoded secrets (passwords, API keys, JWT secrets)
- [ ] All SQL queries use parameterized queries ($1, $2, not string concatenation)
- [ ] User input is validated before use
- [ ] Authentication required on protected endpoints
- [ ] Authorization checked (user can access this resource?)
- [ ] Errors don't leak sensitive information
- [ ] HTTPS enforced in production
- [ ] Dependencies are up to date (no known vulnerabilities)

---

## 🚨 Critical: Never Do These

### 1. Never Hardcode Secrets

**❌ WRONG:**
```go
// DANGER: Secret in code!
jwtManager := auth.NewJWTManager("my-super-secret-key-123")

const dbPassword = "postgres123"

// DANGER: API key in code!
apiKey := "sk-1234567890abcdef"
```

**✅ CORRECT:**
```go
// Load from environment variable
jwtSecret := os.Getenv("JWT_SECRET")
if jwtSecret == "" {
    log.Fatal("JWT_SECRET environment variable not set")
}
jwtManager := auth.NewJWTManager(jwtSecret)
```

**How to check:**
```bash
# Search for potential secrets
grep -r "password.*=.*\"" --include="*.go" internal/
grep -r "secret.*=.*\"" --include="*.go" internal/
grep -r "key.*=.*\"" --include="*.go" internal/

# Should find ZERO results in production code
# (Test code with mock secrets is OK)
```

**Automated check in review.sh:**
```bash
if grep -ri "password.*=.*\"" --include="*.go" ./internal | grep -v "_test.go"; then
    echo "❌ Possible hardcoded password found!"
    exit 1
fi
```

---

### 2. Never Use String Concatenation in SQL

**❌ WRONG (SQL Injection Vulnerability):**
```go
// DANGER: SQL injection!
query := "SELECT * FROM books WHERE id = '" + bookID + "'"
row := db.QueryRow(query)

// Attacker can inject: bookID = "'; DROP TABLE books; --"
// Resulting query: SELECT * FROM books WHERE id = ''; DROP TABLE books; --'
```

**✅ CORRECT:**
```go
// Safe: Parameterized query
query := "SELECT * FROM books WHERE id = $1"
row := db.QueryRow(query, bookID)

// bookID is properly escaped, SQL injection impossible
```

**How to check:**
```bash
# Find dangerous patterns
grep -r "SELECT.*+.*\"" --include="*.go" internal/infrastructure/pkg/repository/
grep -r "INSERT.*+.*\"" --include="*.go" internal/infrastructure/pkg/repository/
grep -r "UPDATE.*+.*\"" --include="*.go" internal/infrastructure/pkg/repository/
grep -r "DELETE.*+.*\"" --include="*.go" internal/infrastructure/pkg/repository/

# Should find ZERO results
```

**SQL injection examples to prevent:**
```go
// ❌ ALL DANGEROUS:
query := "SELECT * FROM users WHERE email = '" + email + "'"
query := fmt.Sprintf("SELECT * FROM users WHERE id = %s", userID)
query := "INSERT INTO books VALUES ('" + id + "', '" + title + "')"

// ✅ ALL SAFE:
query := "SELECT * FROM users WHERE email = $1"
query := "SELECT * FROM users WHERE id = $1"
query := "INSERT INTO books (id, title) VALUES ($1, $2)"
```

---

### 3. Never Trust User Input

**❌ WRONG:**
```go
// DANGER: No validation!
func (h *Handler) CreateBook(w http.ResponseWriter, r *http.Request) {
    var req dto.CreateBookRequest
    json.NewDecoder(r.Body).Decode(&req)

    // Direct use without validation
    book := book.NewEntity(req.Title, req.ISBN)
    h.useCase.Execute(ctx, book)  // ← What if req.Title is empty? Or 10MB string?
}
```

**✅ CORRECT:**
```go
func (h *Handler) CreateBook(w http.ResponseWriter, r *http.Request) {
    var req dto.CreateBookRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        respondError(w, http.StatusBadRequest, "invalid JSON")
        return
    }

    // Validate input
    if err := validate.Struct(req); err != nil {
        respondError(w, http.StatusBadRequest, err.Error())
        return
    }

    // Now safe to use
    book := book.NewEntity(req.Title, req.ISBN)
    h.useCase.Execute(ctx, book)
}
```

**Validation rules:**
```go
type CreateBookRequest struct {
    Title string `json:"title" validate:"required,min=1,max=255"`
    ISBN  string `json:"isbn" validate:"required,len=13,numeric"`
}
```

---

### 4. Never Expose Internal Errors to Users

**❌ WRONG:**
```go
// DANGER: Leaks database structure!
book, err := repo.GetByID(ctx, id)
if err != nil {
    // Returns: "pq: relation 'books' does not exist"
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
}
```

**✅ CORRECT:**
```go
book, err := repo.GetByID(ctx, id)
if err != nil {
    // Log detailed error
    log.Printf("Error fetching book %s: %v", id, err)

    // Return generic error to user
    if errors.Is(err, errors.ErrNotFound) {
        http.Error(w, "book not found", http.StatusNotFound)
    } else {
        http.Error(w, "internal server error", http.StatusInternalServerError)
    }
    return
}
```

**Error messages should:**
- ✅ Be generic to users ("internal server error", "not found")
- ✅ Be detailed in logs (full stack trace, context)
- ❌ Never reveal database structure, file paths, or implementation details

---

## 🔑 Authentication & Authorization

### Authentication: "Who are you?"

**Protect endpoints with JWT middleware:**
```go
// internal/infrastructure/server/routes/router.go
r.Route("/api/v1", func(r chi.Router) {
    // Public endpoints (no auth)
    r.Post("/auth/login", handlers.Auth.Login)
    r.Post("/auth/register", handlers.Auth.Register)

    // Protected endpoints (auth required)
    r.Group(func(r chi.Router) {
        r.Use(authMiddleware)  // ← Enforces authentication

        r.Route("/books", func(r chi.Router) {
            r.Post("/", handlers.Book.CreateBook)
            r.Get("/{id}", handlers.Book.GetBook)
        })
    })
})
```

**Swagger annotation:**
```go
// @Security BearerAuth  ← REQUIRED for protected endpoints
// @Router /books [post]
func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request)
```

---

### Authorization: "Can you do this?"

**Check ownership/permissions:**
```go
func (h *BookHandler) UpdateBook(w http.ResponseWriter, r *http.Request) {
    bookID := chi.URLParam(r, "id")

    // Get current user from JWT claims
    claims := auth.GetClaimsFromContext(r.Context())
    memberID := claims.MemberID

    // Check if user can update this book
    // (e.g., only librarians can update books, or only book owner)
    if claims.Role != "librarian" {
        http.Error(w, "forbidden", http.StatusForbidden)
        return
    }

    // Proceed with update
}
```

**Authorization patterns:**
```go
// ✅ Check role
if claims.Role != "admin" {
    return errors.New("admin access required")
}

// ✅ Check ownership
loan, _ := loanRepo.GetByID(ctx, loanID)
if loan.MemberID != claims.MemberID {
    return errors.New("you don't own this loan")
}

// ✅ Check business rule
member, _ := memberRepo.GetByID(ctx, memberID)
if member.TotalLateFees > 10.0 {
    return errors.New("outstanding late fees must be paid")
}
```

---

## 🛡️ Input Validation

### Validate at the Edge (HTTP Layer)

```go
type CreateBookRequest struct {
    Title       string   `json:"title" validate:"required,min=1,max=255"`
    ISBN        string   `json:"isbn" validate:"required,len=13,numeric"`
    PublishYear int      `json:"publish_year" validate:"required,min=1000,max=2100"`
    AuthorIDs   []string `json:"author_ids" validate:"required,min=1,dive,uuid4"`
}
```

**Validation tags:**
- `required` - Field must not be empty
- `min=N,max=N` - Length/value constraints
- `email` - Must be valid email
- `uuid4` - Must be valid UUID
- `len=N` - Exact length
- `numeric` - Only numbers
- `alpha` - Only letters
- `dive` - Validate each element in slice

### Sanitize Input

```go
// Trim whitespace
title := strings.TrimSpace(req.Title)

// Convert to lowercase for comparison
email := strings.ToLower(req.Email)

// Remove dangerous characters (if storing HTML)
import "html"
safeDescription := html.EscapeString(req.Description)
```

---

## 🔒 Password Security

### Never Store Plain Text Passwords

**❌ WRONG:**
```go
// DANGER: Plain text password!
member := member.Entity{
    Email:    req.Email,
    Password: req.Password,  // ← Stored in plain text!
}
repo.Create(ctx, member)
```

**✅ CORRECT:**
```go
// Hash password with bcrypt
import "golang.org/x/crypto/bcrypt"

hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
if err != nil {
    return err
}

member := member.Entity{
    Email:        req.Email,
    PasswordHash: string(hashedPassword),  // ← Hashed, safe to store
}
repo.Create(ctx, member)
```

### Verify Passwords Correctly

```go
// Verify password
func (uc *LoginUseCase) Execute(ctx context.Context, email, password string) error {
    member, err := uc.memberRepo.GetByEmail(ctx, email)
    if err != nil {
        return errors.New("invalid credentials")  // Generic error
    }

    // Compare hashed password
    err = bcrypt.CompareHashAndPassword([]byte(member.PasswordHash), []byte(password))
    if err != nil {
        return errors.New("invalid credentials")  // Same generic error
    }

    return nil
}
```

**Security notes:**
- ✅ Return same error whether user not found or password wrong (prevents user enumeration)
- ✅ Use bcrypt (designed for passwords, intentionally slow)
- ❌ Don't use MD5, SHA1, or plain SHA256 (too fast, vulnerable to brute force)

---

## 🌐 HTTPS / TLS

### Always Use HTTPS in Production

**In production config:**
```go
server := &http.Server{
    Addr:    ":443",
    Handler: router,
    TLSConfig: &tls.Config{
        MinVersion: tls.VersionTLS13,  // TLS 1.3 only
    },
}

log.Fatal(server.ListenAndServeTLS("cert.pem", "key.pem"))
```

**Never in production:**
```go
// ❌ ONLY for local development!
server.ListenAndServe()  // HTTP without TLS
```

---

## 🕒 Rate Limiting

**Prevent brute force attacks:**
```go
// Example using go-chi middleware
import "github.com/didip/tollbooth"

// Rate limit: 20 requests per minute
limiter := tollbooth.NewLimiter(20, nil)

r.Use(tollbooth_chi.LimitHandler(limiter))
```

**Rate limit login endpoint:**
```go
// More strict for auth endpoints
r.Post("/auth/login", tollbooth_chi.LimitHandler(
    tollbooth.NewLimiter(5, nil),  // 5 attempts per minute
    handlers.Auth.Login,
))
```

---

## 📦 Dependency Security

### Keep Dependencies Updated

```bash
# Check for known vulnerabilities
go list -json -m all | nancy sleuth

# Or use govulncheck
govulncheck ./...
```

**Update dependencies regularly:**
```bash
# Update all dependencies
go get -u ./...

# Run tests after update
make test
```

### Use Go Modules

```go
// go.mod should specify exact versions
require (
    github.com/golang-jwt/jwt/v5 v5.0.0  // Exact version
    // NOT: github.com/golang-jwt/jwt/v5 latest
)
```

---

## 🔍 Security Audit Checklist

### Pre-Deployment Checklist

- [ ] **Secrets:** No hardcoded secrets (grep for "password", "secret", "key")
- [ ] **SQL Injection:** All queries use $1, $2 (no string concatenation)
- [ ] **Authentication:** Protected endpoints use auth middleware
- [ ] **Authorization:** Users can only access their own resources
- [ ] **Input Validation:** All user input validated
- [ ] **Password Hashing:** Passwords hashed with bcrypt
- [ ] **HTTPS:** TLS enabled in production
- [ ] **Error Messages:** Generic errors to users, detailed in logs
- [ ] **Dependencies:** No known vulnerabilities (run govulncheck)
- [ ] **Rate Limiting:** Auth endpoints rate limited

---

## 🔧 Automated Security Checks

**Add to `.claude/scripts/review.sh`:**
```bash
#!/bin/bash

echo "🔒 Running security checks..."

# Check for hardcoded secrets
if grep -ri "password.*=.*\"" --include="*.go" ./internal | grep -v "_test.go" 2>/dev/null; then
    echo "❌ Possible hardcoded password found!"
    exit 1
fi

# Check for SQL injection
if grep -r "SELECT.*+.*\"" --include="*.go" ./internal/infrastructure/pkg/repository/ 2>/dev/null; then
    echo "❌ Possible SQL injection vulnerability (string concatenation in SQL)!"
    exit 1
fi

# Check for vulnerabilities
govulncheck ./... || echo "⚠️  govulncheck not installed or found issues"

echo "✅ Security checks passed!"
```

---

## 🎓 Security Principles

1. **Defense in Depth:** Multiple layers of security (auth + authorization + validation)
2. **Least Privilege:** Users only get minimum permissions needed
3. **Fail Secure:** If something fails, fail closed (deny access)
4. **Never Trust Input:** Validate everything from users
5. **Keep Secrets Secret:** Never in code, never in logs
6. **Update Regularly:** Keep dependencies updated

---

## 📚 Resources

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Go Security Checklist](https://github.com/Checkmarx/Go-SCP)
- [JWT Best Practices](https://cheatsheetseries.owasp.org/cheatsheets/JSON_Web_Token_for_Java_Cheat_Sheet.html)
- [SQL Injection Prevention](https://cheatsheetseries.owasp.org/cheatsheets/SQL_Injection_Prevention_Cheat_Sheet.html)

---

**Last Updated:** 2025-01-19
**Review Frequency:** Before every production deployment
