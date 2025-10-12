package httputil

import (
	"context"
	"library-service/internal/pkg/errors"
	"library-service/internal/pkg/logutil"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

// WrapperOptions contains options for the simplified handler wrapper
type WrapperOptions struct {
	RequireAuth  bool
	RequireAdmin bool
}

// SimpleExecutor is a function type that matches our use case Execute method
type SimpleExecutor[Req, Res any] func(ctx context.Context, req Req) (Res, error)

// CreateHandler creates a simple HTTP handler from an executor function
func CreateHandler[Req, Res any](
	executor SimpleExecutor[Req, Res],
	validate func(req Req) error,
	loggerName, operationName string,
	opts WrapperOptions,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := logutil.HandlerLogger(ctx, loggerName, operationName)

		// Check authentication if required
		if opts.RequireAuth {
			if !checkAuth(w, r, logger) {
				return
			}
		}

		// Check admin role if required
		if opts.RequireAdmin {
			if !checkAdmin(w, r, logger) {
				return
			}
		}

		// Decode request if needed
		var req Req
		if r.Method != http.MethodGet && r.Method != http.MethodDelete {
			if err := DecodeJSON(r, &req); err != nil {
				logger.Error("failed to decode request", zap.Error(err))
				RespondError(w, http.StatusBadRequest, "Invalid request format")
				return
			}

			// Validate request if validator provided
			if validate != nil {
				if err := validate(req); err != nil {
					logger.Error("validation failed", zap.Error(err))
					RespondError(w, http.StatusBadRequest, err.Error())
					return
				}
			}
		}

		// Execute the function
		result, err := executor(ctx, req)
		if err != nil {
			handleError(w, logger, err)
			return
		}

		logger.Info("operation completed successfully")
		RespondJSON(w, http.StatusOK, result)
	}
}

// CreateHandlerWithStatus is like CreateHandler but allows custom success status code
func CreateHandlerWithStatus[Req, Res any](
	executor SimpleExecutor[Req, Res],
	validate func(req Req) error,
	loggerName, operationName string,
	opts WrapperOptions,
	successStatus int,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := logutil.HandlerLogger(ctx, loggerName, operationName)

		// Check authentication if required
		if opts.RequireAuth {
			if !checkAuth(w, r, logger) {
				return
			}
		}

		// Check admin role if required
		if opts.RequireAdmin {
			if !checkAdmin(w, r, logger) {
				return
			}
		}

		// Decode request if needed
		var req Req
		if r.Method != http.MethodGet && r.Method != http.MethodDelete {
			if err := DecodeJSON(r, &req); err != nil {
				logger.Error("failed to decode request", zap.Error(err))
				RespondError(w, http.StatusBadRequest, "Invalid request format")
				return
			}

			// Validate request if validator provided
			if validate != nil {
				if err := validate(req); err != nil {
					logger.Error("validation failed", zap.Error(err))
					RespondError(w, http.StatusBadRequest, err.Error())
					return
				}
			}
		}

		// Execute the function
		result, err := executor(ctx, req)
		if err != nil {
			handleError(w, logger, err)
			return
		}

		logger.Info("operation completed successfully")
		RespondJSON(w, successStatus, result)
	}
}

// CreateAuthHandler creates a handler that extracts auth token from header
func CreateAuthHandler[Res any](
	executor func(ctx context.Context, token string) (Res, error),
	loggerName, operationName string,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := logutil.HandlerLogger(ctx, loggerName, operationName)

		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			logger.Warn("missing authorization header")
			RespondError(w, http.StatusUnauthorized, "Missing authorization header")
			return
		}

		// Check for Bearer prefix
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			logger.Warn("invalid authorization header format")
			RespondError(w, http.StatusUnauthorized, "Invalid authorization header format")
			return
		}

		token := parts[1]

		// Execute the function with token
		result, err := executor(ctx, token)
		if err != nil {
			handleError(w, logger, err)
			return
		}

		logger.Info("operation completed successfully")
		RespondJSON(w, http.StatusOK, result)
	}
}

// checkAuth checks if the request has authentication
func checkAuth(w http.ResponseWriter, r *http.Request, logger *zap.Logger) bool {
	memberID := r.Context().Value("member_id")
	if memberID == nil {
		logger.Warn("unauthorized access attempt")
		RespondError(w, http.StatusUnauthorized, "Authentication required")
		return false
	}
	return true
}

// checkAdmin checks if the request has admin role
func checkAdmin(w http.ResponseWriter, r *http.Request, logger *zap.Logger) bool {
	role := r.Context().Value("role")
	if role != "admin" {
		logger.Warn("forbidden access attempt")
		RespondError(w, http.StatusForbidden, "Admin access required")
		return false
	}
	return true
}

// handleError maps domain errors to HTTP responses
func handleError(w http.ResponseWriter, logger *zap.Logger, err error) {
	logger.Error("operation failed", zap.Error(err))

	switch {
	case errors.Is(err, errors.ErrNotFound):
		RespondError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, errors.ErrAlreadyExists):
		RespondError(w, http.StatusConflict, err.Error())
	case errors.Is(err, errors.ErrValidation):
		RespondError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, errors.ErrUnauthorized):
		RespondError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, errors.ErrForbidden):
		RespondError(w, http.StatusForbidden, err.Error())
	case errors.Is(err, errors.ErrInvalidToken):
		RespondError(w, http.StatusUnauthorized, err.Error())
	case errors.Is(err, errors.ErrDatabase):
		RespondError(w, http.StatusInternalServerError, "Database error")
	default:
		RespondError(w, http.StatusInternalServerError, "Internal server error")
	}
}
