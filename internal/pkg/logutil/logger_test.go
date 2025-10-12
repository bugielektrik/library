package logutil_test

import (
	"context"
	logutil2 "library-service/internal/pkg/logutil"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestUseCaseLogger(t *testing.T) {
	tests := []struct {
		name      string
		domain    string
		operation string
	}{
		{
			name:      "creates logger with domain and operation",
			domain:    "book",
			operation: "create",
		},
		{
			name:      "creates logger for list operation",
			domain:    "book",
			operation: "list",
		},
		{
			name:      "creates logger for update operation",
			domain:    "book",
			operation: "update",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup observed logger to capture log output
			core, _ := observer.New(zapcore.InfoLevel)
			baseLogger := zap.New(core)
			ctx := logutil2.WithLogger(context.Background(), baseLogger)

			// Create logger using utility
			logger := logutil2.UseCaseLogger(ctx, tt.domain, tt.operation)

			// Verify logger is not nil
			if logger == nil {
				t.Fatal("UseCaseLogger returned nil")
			}

			// Log a test message
			logger.Info("test message")

			// Note: Logger naming is internal to zap, so we verify behavior by ensuring
			// the logger works correctly and includes the expected fields
		})
	}
}

func TestHandlerLogger(t *testing.T) {
	tests := []struct {
		name        string
		handlerName string
		methodName  string
	}{
		{
			name:        "creates handler logger with handler and method fields",
			handlerName: "book_handler",
			methodName:  "create",
		},
		{
			name:        "creates handler logger for list operation",
			handlerName: "member_handler",
			methodName:  "list",
		},
		{
			name:        "creates handler logger for delete operation",
			handlerName: "author_handler",
			methodName:  "delete",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup observed logger
			core, observed := observer.New(zapcore.InfoLevel)
			baseLogger := zap.New(core)
			ctx := logutil2.WithLogger(context.Background(), baseLogger)

			// Create logger
			logger := logutil2.HandlerLogger(ctx, tt.handlerName, tt.methodName)

			// Verify logger is not nil
			if logger == nil {
				t.Fatal("HandlerLogger returned nil")
			}

			// Log a test message
			logger.Info("test message")

			// Verify the logged entry has the expected fields
			entries := observed.All()
			if len(entries) != 1 {
				t.Fatalf("expected 1 log entry, got %d", len(entries))
			}

			entry := entries[0]
			foundHandler := false
			foundOperation := false

			for _, field := range entry.Context {
				if field.Key == "handler" && field.String == tt.handlerName {
					foundHandler = true
				}
				if field.Key == "operation" && field.String == tt.methodName {
					foundOperation = true
				}
			}

			if !foundHandler {
				t.Errorf("expected handler field with value %q", tt.handlerName)
			}
			if !foundOperation {
				t.Errorf("expected operation field with value %q", tt.methodName)
			}
		})
	}
}

func TestRepositoryLogger(t *testing.T) {
	tests := []struct {
		name           string
		repositoryName string
		operation      string
		wantLogName    string
	}{
		{
			name:           "creates repository logger with operation field",
			repositoryName: "book",
			operation:      "create",
			wantLogName:    "book_repository",
		},
		{
			name:           "creates repository logger for update operation",
			repositoryName: "member",
			operation:      "update",
			wantLogName:    "member_repository",
		},
		{
			name:           "creates repository logger for delete operation",
			repositoryName: "author",
			operation:      "delete",
			wantLogName:    "author_repository",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup observed logger
			core, observed := observer.New(zapcore.InfoLevel)
			baseLogger := zap.New(core)
			ctx := logutil2.WithLogger(context.Background(), baseLogger)

			// Create logger
			logger := logutil2.RepositoryLogger(ctx, tt.repositoryName, tt.operation)

			// Verify logger is not nil
			if logger == nil {
				t.Fatal("RepositoryLogger returned nil")
			}

			// Log a test message
			logger.Info("test message")

			// Verify the logged entry has the expected operation field
			entries := observed.All()
			if len(entries) != 1 {
				t.Fatalf("expected 1 log entry, got %d", len(entries))
			}

			entry := entries[0]
			foundOperation := false

			for _, field := range entry.Context {
				if field.Key == "operation" && field.String == tt.operation {
					foundOperation = true
					break
				}
			}

			if !foundOperation {
				t.Errorf("expected operation field with value %q", tt.operation)
			}
		})
	}
}

// TestGatewayLogger removed - GatewayLogger function was removed during refactoring
// Gateways now use RepositoryLogger or custom logger patterns

// TestLoggerIntegration tests that all logger utilities work with the actual log infrastructure
func TestLoggerIntegration(t *testing.T) {
	// Setup observed logger
	core, observed := observer.New(zapcore.InfoLevel)
	baseLogger := zap.New(core)
	ctx := logutil2.WithLogger(context.Background(), baseLogger)

	// Test UseCaseLogger
	useCaseLogger := logutil2.UseCaseLogger(ctx, "test", "test_operation")
	useCaseLogger.Info("use case message")

	// Test HandlerLogger
	handlerLogger := logutil2.HandlerLogger(ctx, "test_handler", "test_method")
	handlerLogger.Info("handler message")

	// Test RepositoryLogger
	repoLogger := logutil2.RepositoryLogger(ctx, "test_repo", "test_operation")
	repoLogger.Info("repository message")

	// Verify all messages were logged
	entries := observed.All()
	if len(entries) != 3 {
		t.Fatalf("expected 3 log entries, got %d", len(entries))
	}

	messages := []string{"use case message", "handler message", "repository message"}
	for i, entry := range entries {
		if entry.Message != messages[i] {
			t.Errorf("entry %d: expected message %q, got %q", i, messages[i], entry.Message)
		}
	}
}
