package helpers

import (
	"context"
	"testing"
	"time"

	"library-service/internal/infrastructure/auth"
)

// TestContext creates a context with timeout for tests
func TestContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// TestContextWithTimeout creates a context with custom timeout
func TestContextWithTimeout(t *testing.T, timeout time.Duration) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	t.Cleanup(cancel)
	return ctx
}

// TestContextWithCancel creates a cancellable context
func TestContextWithCancel(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	return ctx, cancel
}

// TestContextWithAuth creates a context with authentication claims
func TestContextWithAuth(t *testing.T, memberID, email, role string) context.Context {
	t.Helper()
	ctx := TestContext(t)

	// Add auth claims to context
	authClaims := &auth.Claims{
		MemberID: memberID,
		Email:    email,
		Role:     role,
	}

	return context.WithValue(ctx, "claims", authClaims)
}

// TestContextWithUserAuth creates a context with regular user authentication
func TestContextWithUserAuth(t *testing.T, memberID string) context.Context {
	t.Helper()
	return TestContextWithAuth(t, memberID, "test@example.com", "user")
}

// TestContextWithAdminAuth creates a context with admin authentication
func TestContextWithAdminAuth(t *testing.T, memberID string) context.Context {
	t.Helper()
	return TestContextWithAuth(t, memberID, "admin@example.com", "admin")
}

// TestContextWithValue adds a value to test context
func TestContextWithValue(t *testing.T, key, value interface{}) context.Context {
	t.Helper()
	ctx := TestContext(t)
	return context.WithValue(ctx, key, value)
}

// TestContextWithRequestID adds a request ID to context
func TestContextWithRequestID(t *testing.T, requestID string) context.Context {
	t.Helper()
	return TestContextWithValue(t, "request-id", requestID)
}

// ExtractClaimsFromContext extracts auth claims from context
func ExtractClaimsFromContext(ctx context.Context) (*auth.Claims, bool) {
	claims, ok := ctx.Value("claims").(*auth.Claims)
	return claims, ok
}
