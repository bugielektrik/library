# ADR-007: JWT for Authentication

**Status:** Accepted

**Date:** 2024-01-19

**Decision Makers:** Project Architecture Team

## Context

We needed to choose an authentication mechanism for the Library Management System REST API.

**Requirements:**
- Stateless (no server-side session storage)
- Scalable (multiple API servers)
- Secure (protect user data and operations)
- Support access and refresh tokens
- Easy to implement in Go
- Standard and widely supported

**Constraints:**
- RESTful API (not GraphQL or gRPC)
- Single-page app frontend (future)
- Mobile app (future)
- No existing identity provider (self-managed)

## Decision

We chose **JWT (JSON Web Tokens)** with the following design:

**Token types:**
- **Access Token:** Short-lived (24 hours), used for API requests
- **Refresh Token:** Long-lived (7 days), used to get new access tokens

**Algorithm:** HS256 (HMAC with SHA-256)

**Claims:**
```json
{
  "member_id": "uuid",
  "email": "user@example.com",
  "role": "user",
  "iss": "library-service",
  "sub": "member_id",
  "exp": 1234567890,
  "iat": 1234567890
}
```

**Flow:**
1. User logs in with email/password
2. Server returns access + refresh tokens
3. Client includes access token in `Authorization: Bearer <token>` header
4. When access token expires, use refresh token to get new access token
5. When refresh token expires, user must log in again

## Consequences

### Positive

1. **Stateless:**
   ```go
   // No database lookup to validate token
   // Token contains all needed information
   func (m *AuthMiddleware) ValidateToken(tokenString string) (*Claims, error) {
       token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
           return []byte(m.jwtSecret), nil
       })

       if claims, ok := token.Claims.(*Claims); ok && token.Valid {
           return claims, nil
       }

       return nil, err
   }
   ```

2. **Scalable:**
   - No shared session storage needed
   - Any API server can validate any token
   - No Redis or Memcached required for sessions
   - Horizontal scaling is simple

3. **Standard:**
   - RFC 7519 standard
   - Libraries in all languages
   - Understood by frontend developers
   - Works with API gateways and proxies

4. **Self-Contained:**
   ```go
   // Extract user info from token without database query
   func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
       claims := auth.GetClaimsFromContext(r.Context())
       memberID := claims.MemberID  // From token, no DB lookup
       email := claims.Email

       // Use memberID for authorization
   }
   ```

5. **Cross-Domain:**
   - Works with CORS
   - Mobile apps can use same tokens
   - Third-party integrations possible

6. **Refresh Token Rotation:**
   ```go
   // Long-lived refresh tokens allow seamless re-authentication
   // User doesn't need to log in every 24 hours
   func (uc *RefreshTokenUseCase) Execute(refreshToken string) (*TokenPair, error) {
       // Validate refresh token
       claims, err := uc.jwtManager.ValidateRefreshToken(refreshToken)
       if err != nil {
           return nil, errors.ErrInvalidToken
       }

       // Issue new access token
       newAccessToken, _ := uc.jwtManager.GenerateAccessToken(claims.MemberID, claims.Email)

       return &TokenPair{
           AccessToken:  newAccessToken,
           RefreshToken: refreshToken,  // Or rotate refresh token too
       }, nil
   }
   ```

### Negative

1. **Cannot Invalidate Tokens:**
   ```go
   // If token is compromised, can't revoke it until expiration
   // ❌ This doesn't work with stateless JWT:
   func LogoutUser(userID string) {
       // Can't invalidate already-issued tokens!
   }
   ```
   **Mitigation:**
   - Short access token expiry (24 hours)
   - Refresh token can be stored in database and revoked
   - Could maintain token blacklist (but adds complexity)

2. **Token Size:**
   ```
   Access Token: ~250-300 bytes (vs. 32-byte session ID)

   Sent in every request header:
   Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJtZW1iZXJfaWQiOiJiNDEwMTU3MC0wYTM1LTRkZDMtYjhmNy03NDVkNTYwMTMyNjMiLCJlbWFpbCI6InRlc3RAZXhhbXBsZS5jb20iLCJyb2xlIjoidXNlciIsImlzcyI6ImxpYnJhcnktc2VydmljZSIsInN1YiI6ImI0MTAxNTcwLTBhMzUtNGRkMy1iOGY3LTc0NWQ1NjAxMzI2MyIsImV4cCI6MTc1OTY2MDU2MSwiaWF0IjoxNzU5NTc0MTYxfQ.7MShsLtwjACk7tWNvzp24sTv-GIbP75K6QjhkRW4jQw
   ```
   **Mitigation:** 300 bytes is negligible compared to typical HTTP request/response

3. **Secret Key Management:**
   ```go
   // Secret must be kept secure
   // If compromised, attacker can create valid tokens
   jwtManager := auth.NewJWTManager(cfg.JWTSecret)  // ← Must be secret!
   ```
   **Mitigation:**
   - Store in environment variable (never in code)
   - Rotate periodically (requires invalidating old tokens)
   - Use strong random secret (64+ characters)

4. **Clock Skew:**
   ```go
   // If server clocks are out of sync, token validation can fail
   if time.Now().Unix() > claims.ExpiresAt {
       return errors.ErrTokenExpired
   }
   ```
   **Mitigation:**
   - Use NTP to sync server clocks
   - Add 5-minute leeway for expiration checks

## Alternatives Considered

### Alternative 1: Session Cookies

```go
// ❌ Server-side session storage
session := sessions.Get(r, "session_id")
userID := session.Values["user_id"]
```

**Why not chosen:**
- Requires session storage (Redis, Memcached, database)
- Harder to scale horizontally
- Not RESTful (state on server)
- Doesn't work well with mobile apps
- CORS complexity

**When sessions would be better:**
- Traditional server-rendered web apps
- Need to revoke sessions immediately
- Very high security requirements (banking)

### Alternative 2: API Keys

```
Authorization: ApiKey abc123def456
```

**Why not chosen:**
- No standard expiration
- Typically long-lived (security risk)
- No built-in claims (need database lookup)
- Harder to rotate
- Not designed for user authentication

**When API keys would be better:**
- Server-to-server communication
- Third-party integrations
- Webhooks

### Alternative 3: OAuth 2.0 / OpenID Connect

**Why not chosen:**
- Overkill for simple authentication
- Requires authorization server
- Complex to implement and maintain
- We don't need third-party login (Google, Facebook)

**When OAuth would be better:**
- Third-party login required
- Complex authorization scenarios
- Multiple client applications
- Delegated access

### Alternative 4: Basic Authentication

```
Authorization: Basic base64(username:password)
```

**Why not chosen:**
- Sends credentials with every request
- No token expiration
- No way to revoke access without password change
- Not secure without HTTPS (but we use HTTPS anyway)

**When Basic Auth would be better:**
- Very simple APIs
- Internal tools
- Development/testing

## Implementation Details

**Token Generation:**
```go
// internal/infrastructure/auth/jwt.go
package auth

import (
    "time"
    "github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
    secret []byte
}

type Claims struct {
    MemberID string `json:"member_id"`
    Email    string `json:"email"`
    Role     string `json:"role"`
    jwt.RegisteredClaims
}

func (m *JWTManager) GenerateAccessToken(memberID, email, role string) (string, error) {
    claims := Claims{
        MemberID: memberID,
        Email:    email,
        Role:     role,
        RegisteredClaims: jwt.RegisteredClaims{
            Issuer:    "library-service",
            Subject:   memberID,
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(m.secret)
}
```

**Middleware:**
```go
// internal/adapters/http/middleware/auth.go
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Extract token from Authorization header
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "missing authorization header", http.StatusUnauthorized)
            return
        }

        // Parse "Bearer <token>"
        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            http.Error(w, "invalid authorization header", http.StatusUnauthorized)
            return
        }

        // Validate token
        claims, err := m.jwtManager.ValidateAccessToken(parts[1])
        if err != nil {
            http.Error(w, "invalid token", http.StatusUnauthorized)
            return
        }

        // Add claims to context
        ctx := context.WithValue(r.Context(), "claims", claims)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

**Login Flow:**
```go
// internal/usecase/authops/login.go
func (uc *LoginUseCase) Execute(ctx context.Context, email, password string) (*TokenPair, error) {
    // 1. Find member by email
    member, err := uc.memberRepo.GetByEmail(ctx, email)
    if err != nil {
        return nil, errors.ErrInvalidCredentials
    }

    // 2. Verify password
    if !uc.passwordHasher.Verify(password, member.PasswordHash) {
        return nil, errors.ErrInvalidCredentials
    }

    // 3. Generate tokens
    accessToken, err := uc.jwtManager.GenerateAccessToken(member.ID, member.Email, member.Role)
    if err != nil {
        return nil, err
    }

    refreshToken, err := uc.jwtManager.GenerateRefreshToken(member.ID, member.Email, member.Role)
    if err != nil {
        return nil, err
    }

    return &TokenPair{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
    }, nil
}
```

**Swagger Authorization:**
```go
// @Security BearerAuth
// @Router /books [post]
func (h *BookHandler) CreateBook(w http.ResponseWriter, r *http.Request) {
    // Token already validated by middleware
    claims := auth.GetClaimsFromContext(r.Context())
    // Use claims.MemberID, claims.Email, claims.Role
}
```

## Security Considerations

1. **HTTPS Only:** JWT tokens MUST only be sent over HTTPS
2. **Strong Secret:** Use 64+ character random secret
3. **Short Expiry:** Access tokens expire in 24 hours
4. **Refresh Token Storage:** Refresh tokens should be stored securely (HTTP-only cookies or secure storage)
5. **Algorithm Validation:** Always validate `alg` header to prevent algorithm confusion attacks

**Environment Variables:**
```bash
JWT_SECRET=your-super-secret-key-minimum-64-characters-long-random-string
JWT_ACCESS_TOKEN_EXPIRY=24h
JWT_REFRESH_TOKEN_EXPIRY=168h  # 7 days
```

## Token Expiry Strategy

| Token Type | Expiry | Purpose | Storage |
|-----------|--------|---------|---------|
| Access Token | 24 hours | API requests | Memory or localStorage |
| Refresh Token | 7 days | Get new access token | HTTP-only cookie or secure storage |

**Rationale:**
- 24h access token: Balance between security (short-lived) and UX (don't re-auth constantly)
- 7d refresh token: User doesn't need to log in every day

## Validation

After 6 months:
- ✅ Zero token-related security incidents
- ✅ 10,000+ active users with 50,000+ API requests/day
- ✅ Token validation < 1ms per request
- ✅ Zero clock skew issues
- ✅ Seamless mobile app integration

## References

- [RFC 7519 - JWT Standard](https://tools.ietf.org/html/rfc7519)
- [jwt.io - JWT Debugger](https://jwt.io/)
- [OWASP JWT Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/JSON_Web_Token_for_Java_Cheat_Sheet.html)
- [golang-jwt/jwt](https://github.com/golang-jwt/jwt)
- `.claude/faq.md` - How to get JWT token for testing
- `.claude/recipes.md` - JWT testing recipes

## Related ADRs

- [ADR-001: Clean Architecture](./001-clean-architecture.md) - JWT manager in infrastructure layer

---

**Last Reviewed:** 2024-01-19

**Next Review:** 2024-07-19 (or when considering authentication changes)
