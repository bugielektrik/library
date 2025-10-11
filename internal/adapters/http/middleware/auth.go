package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"library-service/internal/adapters/http/dto"
	"library-service/internal/domain/member"
	"library-service/internal/infrastructure/auth"
	"library-service/pkg/errors"
	"library-service/pkg/httputil"
)

// ContextKey type for context values
type ContextKey string

const (
	// ContextKeyMemberID stores the authenticated member's ID
	ContextKeyMemberID ContextKey = "member_id"
	// ContextKeyMemberEmail stores the authenticated member's email
	ContextKeyMemberEmail ContextKey = "member_email"
	// ContextKeyMemberRole stores the authenticated member's role
	ContextKeyMemberRole ContextKey = "member_role"
	// ContextKeyClaims stores the JWT claims
	ContextKeyClaims ContextKey = "jwt_claims"
)

// AuthMiddleware handles JWT authentication for protected routes
type AuthMiddleware struct {
	jwtService *auth.JWTService
}

// NewAuthMiddleware creates a new auth middleware instance
func NewAuthMiddleware(jwtService *auth.JWTService) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
	}
}

// Authenticate is a middleware that validates JWT tokens
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := m.validateAndExtractClaims(w, r)
		if claims == nil {
			return
		}

		ctx := addClaimsToContext(r.Context(), claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole is a middleware that checks if the user has the required role
func (m *AuthMiddleware) RequireRole(roles ...member.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := m.validateAndExtractClaims(w, r)
			if claims == nil {
				return
			}

			// Check if user has required role
			hasRole := false
			for _, role := range roles {
				if claims.Role == string(role) {
					hasRole = true
					break
				}
			}

			if !hasRole {
				m.respondError(w, errors.ErrForbidden.WithDetails("required_roles", roles))
				return
			}

			ctx := addClaimsToContext(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAdmin is a convenience middleware that requires admin role
func (m *AuthMiddleware) RequireAdmin(next http.Handler) http.Handler {
	return m.RequireRole(member.RoleAdmin)(next)
}

// validateAndExtractClaims validates the JWT token and returns claims.
// If validation fails, it writes an error response and returns nil.
func (m *AuthMiddleware) validateAndExtractClaims(w http.ResponseWriter, r *http.Request) *auth.Claims {
	token := m.extractToken(r)
	if token == "" {
		m.respondError(w, errors.ErrUnauthorized.WithDetails("reason", "missing or invalid authorization header"))
		return nil
	}

	claims, err := m.jwtService.ValidateToken(token)
	if err != nil {
		m.respondError(w, errors.ErrUnauthorized.WithDetails("reason", err.Error()))
		return nil
	}

	return claims
}

// addClaimsToContext adds JWT claims to the request context
func addClaimsToContext(ctx context.Context, claims *auth.Claims) context.Context {
	ctx = context.WithValue(ctx, ContextKeyMemberID, claims.MemberID)
	ctx = context.WithValue(ctx, ContextKeyMemberEmail, claims.Email)
	ctx = context.WithValue(ctx, ContextKeyMemberRole, claims.Role)
	ctx = context.WithValue(ctx, ContextKeyClaims, claims)
	return ctx
}

// extractToken extracts the JWT token from the Authorization header
func (m *AuthMiddleware) extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return ""
	}

	// Check for Bearer prefix
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return ""
	}

	return parts[1]
}

// respondError sends an error response
func (m *AuthMiddleware) respondError(w http.ResponseWriter, err error) {
	w.Header().Set(httputil.HeaderContentType, httputil.ContentTypeJSON)

	status := errors.GetHTTPStatus(err)
	w.WriteHeader(status)

	response := dto.FromError(err)

	// Write response (ignore error as header is already sent)
	_ = json.NewEncoder(w).Encode(response)
}

// GetMemberIDFromContext extracts member ID from context
func GetMemberIDFromContext(ctx context.Context) (string, bool) {
	memberID, ok := ctx.Value(ContextKeyMemberID).(string)
	return memberID, ok
}

// GetMemberEmailFromContext extracts member email from context
func GetMemberEmailFromContext(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(ContextKeyMemberEmail).(string)
	return email, ok
}

// GetMemberRoleFromContext extracts member role from context
func GetMemberRoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(ContextKeyMemberRole).(string)
	return role, ok
}

// GetClaimsFromContext extracts JWT claims from context
func GetClaimsFromContext(ctx context.Context) (*auth.Claims, bool) {
	claims, ok := ctx.Value(ContextKeyClaims).(*auth.Claims)
	return claims, ok
}
