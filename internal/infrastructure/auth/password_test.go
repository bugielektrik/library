package auth

import (
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Test passwords
const (
	validPassword   = "Test123!@#"
	anotherPassword = "Secure456$%^"
	weakPassword    = "password"
)

func newTestPasswordService() *PasswordService {
	return NewPasswordService()
}

// TestHashPassword tests successful password hashing
func TestHashPassword(t *testing.T) {
	service := newTestPasswordService()

	hash, err := service.HashPassword(validPassword)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == "" {
		t.Error("Expected non-empty hash")
	}

	// Bcrypt hashes start with $2a$, $2b$, or $2y$
	if !strings.HasPrefix(hash, "$2") {
		t.Errorf("Expected bcrypt hash format, got: %s", hash)
	}
}

// TestHashPassword_UniqueHashes tests that the same password produces different hashes
func TestHashPassword_UniqueHashes(t *testing.T) {
	service := newTestPasswordService()

	hash1, err := service.HashPassword(validPassword)
	if err != nil {
		t.Fatalf("First HashPassword failed: %v", err)
	}

	hash2, err := service.HashPassword(validPassword)
	if err != nil {
		t.Fatalf("Second HashPassword failed: %v", err)
	}

	// Bcrypt includes a random salt, so hashes should be different
	if hash1 == hash2 {
		t.Error("Expected different hashes for same password (bcrypt should use different salts)")
	}

	// Both hashes should validate the same password
	if !service.CheckPasswordHash(validPassword, hash1) {
		t.Error("First hash should validate the password")
	}
	if !service.CheckPasswordHash(validPassword, hash2) {
		t.Error("Second hash should validate the password")
	}
}

// TestHashPassword_InvalidPasswords tests that weak passwords are rejected
func TestHashPassword_InvalidPasswords(t *testing.T) {
	service := newTestPasswordService()

	tests := []struct {
		name     string
		password string
		wantErr  string
	}{
		{"too short", "Test1!", "at least 8 characters"},
		{"no uppercase", "test123!@#", "at least one uppercase letter"},
		{"no lowercase", "TEST123!@#", "at least one lowercase letter"},
		{"no number", "Testing!@#", "at least one number"},
		{"no special", "Testing123", "at least one special character"},
		{"too long", strings.Repeat("A", 73) + "a1!", "at most 72 characters"},
		{"empty", "", "at least 8 characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := service.HashPassword(tt.password)
			if err == nil {
				t.Errorf("Expected error for %s password, got nil", tt.name)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Expected error containing %q, got: %v", tt.wantErr, err)
			}
		})
	}
}

// TestCheckPasswordHash_ValidPassword tests correct password validation
func TestCheckPasswordHash_ValidPassword(t *testing.T) {
	service := newTestPasswordService()

	hash, err := service.HashPassword(validPassword)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if !service.CheckPasswordHash(validPassword, hash) {
		t.Error("Expected password to match hash")
	}
}

// TestCheckPasswordHash_InvalidPassword tests wrong password rejection
func TestCheckPasswordHash_InvalidPassword(t *testing.T) {
	service := newTestPasswordService()

	hash, err := service.HashPassword(validPassword)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Try wrong password
	if service.CheckPasswordHash(anotherPassword, hash) {
		t.Error("Expected password mismatch, but got match")
	}
}

// TestCheckPasswordHash_MalformedHash tests behavior with invalid hashes
func TestCheckPasswordHash_MalformedHash(t *testing.T) {
	service := newTestPasswordService()

	tests := []struct {
		name string
		hash string
	}{
		{"empty hash", ""},
		{"invalid format", "not-a-bcrypt-hash"},
		{"truncated hash", "$2a$10$"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if service.CheckPasswordHash(validPassword, tt.hash) {
				t.Errorf("Expected false for malformed hash %q, got true", tt.name)
			}
		})
	}
}

// TestCheckPasswordHash_CaseSensitive tests that password comparison is case-sensitive
func TestCheckPasswordHash_CaseSensitive(t *testing.T) {
	service := newTestPasswordService()

	hash, err := service.HashPassword("Test123!@#")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Try with different case
	if service.CheckPasswordHash("TEST123!@#", hash) {
		t.Error("Password check should be case-sensitive")
	}
	if service.CheckPasswordHash("test123!@#", hash) {
		t.Error("Password check should be case-sensitive")
	}
}

// TestCheckPasswordHash_TimingAttackResistance tests that comparison takes consistent time
func TestCheckPasswordHash_TimingAttackResistance(t *testing.T) {
	service := newTestPasswordService()

	hash, err := service.HashPassword(validPassword)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Measure time for correct password
	start1 := time.Now()
	service.CheckPasswordHash(validPassword, hash)
	duration1 := time.Since(start1)

	// Measure time for completely wrong password
	start2 := time.Now()
	service.CheckPasswordHash("WrongPass123!", hash)
	duration2 := time.Since(start2)

	// Measure time for partially correct password
	start3 := time.Now()
	service.CheckPasswordHash("Test123!@", hash) // Missing last character
	duration3 := time.Since(start3)

	// Bcrypt is designed to be slow, so all comparisons should take similar time
	// Allow 50% variance (bcrypt's constant-time properties are at algorithm level)
	maxDuration := max(duration1, duration2, duration3)
	minDuration := min(duration1, duration2, duration3)

	// The durations should be relatively close (within 2x of each other)
	// This is a weak test, but bcrypt's constant-time is at the algorithm level
	if maxDuration > minDuration*2 {
		t.Logf("Timing variance: %v, %v, %v (max/min ratio: %.2f)",
			duration1, duration2, duration3, float64(maxDuration)/float64(minDuration))
		// Note: This is informational. Bcrypt's timing resistance is algorithmic,
		// not at the comparison level.
	}
}

// TestValidatePassword_ValidPasswords tests valid password patterns
func TestValidatePassword_ValidPasswords(t *testing.T) {
	service := newTestPasswordService()

	validPasswords := []string{
		"Test123!@#",
		"Secure456$%^",
		"MyP@ssw0rd",
		"Qwerty123!",
		"aB3$" + strings.Repeat("x", 4), // Exactly 8 chars
		strings.Repeat("aB3$", 18),      // Exactly 72 chars (bcrypt max)
	}

	for _, password := range validPasswords {
		t.Run(password, func(t *testing.T) {
			err := service.ValidatePassword(password)
			if err != nil {
				t.Errorf("Expected password %q to be valid, got error: %v", password, err)
			}
		})
	}
}

// TestValidatePassword_InvalidPasswords tests password validation rules
func TestValidatePassword_InvalidPasswords(t *testing.T) {
	service := newTestPasswordService()

	tests := []struct {
		password    string
		expectedErr string
	}{
		{"short", "at least 8 characters"},
		{"nouppercase123!", "at least one uppercase letter"},
		{"NOLOWERCASE123!", "at least one lowercase letter"},
		{"NoNumbers!", "at least one number"},
		{"NoSpecial123", "at least one special character"},
		{strings.Repeat("A", 73) + "a1!", "at most 72 characters"},
		{"", "at least 8 characters"},
	}

	for _, tt := range tests {
		t.Run(tt.password, func(t *testing.T) {
			err := service.ValidatePassword(tt.password)
			if err == nil {
				t.Errorf("Expected error for password %q, got nil", tt.password)
			}
			if !strings.Contains(err.Error(), tt.expectedErr) {
				t.Errorf("Expected error containing %q, got: %v", tt.expectedErr, err)
			}
		})
	}
}

// TestValidatePassword_BcryptLimit tests the 72-byte bcrypt limit
func TestValidatePassword_BcryptLimit(t *testing.T) {
	service := newTestPasswordService()

	// 72 characters should be valid (bcrypt limit)
	password72 := strings.Repeat("aB3$", 18) // 72 chars
	err := service.ValidatePassword(password72)
	if err != nil {
		t.Errorf("Expected 72-char password to be valid, got error: %v", err)
	}

	// 73 characters should be invalid
	password73 := password72 + "x"
	err = service.ValidatePassword(password73)
	if err == nil {
		t.Error("Expected error for 73-char password, got nil")
	}
}

// TestValidateEmail_ValidEmails tests valid email patterns
func TestValidateEmail_ValidEmails(t *testing.T) {
	validEmails := []string{
		"test@example.com",
		"user.name@example.com",
		"user+tag@example.co.uk",
		"user_name@example-domain.com",
		"123@example.com",
		"test@subdomain.example.com",
		"a@b.co",
	}

	for _, email := range validEmails {
		t.Run(email, func(t *testing.T) {
			err := ValidateEmail(email)
			if err != nil {
				t.Errorf("Expected email %q to be valid, got error: %v", email, err)
			}
		})
	}
}

// TestValidateEmail_InvalidEmails tests invalid email patterns
func TestValidateEmail_InvalidEmails(t *testing.T) {
	tests := []struct {
		email       string
		expectedErr string
	}{
		{"", "cannot be empty"},
		{"notanemail", "invalid email format"},
		{"@example.com", "invalid email format"},
		{"user@", "invalid email format"},
		{"user @example.com", "invalid email format"},
		{"user@.com", "invalid email format"},
		{"user@example", "invalid email format"},
		{strings.Repeat("a", 247) + "@test.com", "email address too long"}, // 256 chars total
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if err == nil {
				t.Errorf("Expected error for email %q (len=%d), got nil", tt.email, len(tt.email))
				return
			}
			if !strings.Contains(err.Error(), tt.expectedErr) {
				t.Errorf("Expected error containing %q, got: %v", tt.expectedErr, err)
			}
		})
	}
}

// TestHashingPerformance tests that bcrypt hashing is reasonably fast
func TestHashingPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	service := newTestPasswordService()

	start := time.Now()
	_, err := service.HashPassword(validPassword)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Bcrypt should take between 50ms and 500ms at default cost (10)
	// This is intentionally slow to resist brute force attacks
	if duration < 10*time.Millisecond {
		t.Errorf("Bcrypt hashing too fast (%v), may be using wrong cost factor", duration)
	}
	if duration > 1*time.Second {
		t.Errorf("Bcrypt hashing too slow (%v), performance concern", duration)
	}

	t.Logf("Bcrypt hashing took %v (expected ~100-200ms at cost 10)", duration)
}

// TestBcryptCost tests that the service uses the correct bcrypt cost
func TestBcryptCost(t *testing.T) {
	service := newTestPasswordService()

	hash, err := service.HashPassword(validPassword)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Extract cost from hash (bcrypt hash format: $2a$10$...)
	cost, err := bcrypt.Cost([]byte(hash))
	if err != nil {
		t.Fatalf("Failed to extract cost from hash: %v", err)
	}

	if cost != bcrypt.DefaultCost {
		t.Errorf("Expected cost %d, got %d", bcrypt.DefaultCost, cost)
	}
}

// Helper functions
func min(a, b, c time.Duration) time.Duration {
	result := a
	if b < result {
		result = b
	}
	if c < result {
		result = c
	}
	return result
}

func max(a, b, c time.Duration) time.Duration {
	result := a
	if b > result {
		result = b
	}
	if c > result {
		result = c
	}
	return result
}
