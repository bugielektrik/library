package strutil

import "testing"

func TestSafeString(t *testing.T) {
	tests := []struct {
		name     string
		input    *string
		expected string
	}{
		{
			name:     "nil pointer returns empty string",
			input:    nil,
			expected: "",
		},
		{
			name:     "valid pointer returns value",
			input:    stringPtr("hello"),
			expected: "hello",
		},
		{
			name:     "empty string pointer returns empty string",
			input:    stringPtr(""),
			expected: "",
		},
		{
			name:     "string with spaces preserved",
			input:    stringPtr("  hello world  "),
			expected: "  hello world  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SafeString(tt.input)
			if result != tt.expected {
				t.Errorf("SafeString() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSafeStringPtr(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "non-empty string",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "string with special characters",
			input:    "hello@#$%^&*()",
			expected: "hello@#$%^&*()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SafeStringPtr(tt.input)
			if result == nil {
				t.Errorf("SafeStringPtr() returned nil")
				return
			}
			if *result != tt.expected {
				t.Errorf("SafeStringPtr() = %q, want %q", *result, tt.expected)
			}
		})
	}
}

// Helper function to create string pointers for tests
func stringPtr(s string) *string {
	return &s
}
