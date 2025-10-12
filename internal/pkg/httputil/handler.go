package httputil

import (
	"context"
	"encoding/json"
	"library-service/internal/pkg/errors"
	"library-service/internal/pkg/logutil"
	"net/http"

	"go.uber.org/zap"
)

// HandlerFunc is a generic handler function type
type HandlerFunc[Req, Res any] func(ctx context.Context, req Req) (Res, error)

// UseCase interface for generic use case execution
type UseCase[Req, Res any] interface {
	Execute(ctx context.Context, req Req) (Res, error)
}

// Validator interface for request validation
type Validator interface {
	ValidateStruct(v interface{}) error
}

// HandlerOptions contains options for the handler wrapper
type HandlerOptions struct {
	RequireAuth   bool
	RequireAdmin  bool
	LoggerName    string
	OperationName string
}

// WrapHandler creates a generic HTTP handler from a use case
func WrapHandler[Req, Res any](
	useCase UseCase[Req, Res],
	validator Validator,
	opts HandlerOptions,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Setup logger
		logger := logutil.HandlerLogger(ctx, opts.LoggerName, opts.OperationName)

		// Check authentication if required
		if opts.RequireAuth {
			memberID := r.Context().Value("member_id")
			if memberID == nil {
				logger.Warn("unauthorized access attempt")
				RespondError(w, http.StatusUnauthorized, "Authentication required")
				return
			}
		}

		// Check admin role if required
		if opts.RequireAdmin {
			role := r.Context().Value("role")
			if role != "admin" {
				logger.Warn("forbidden access attempt")
				RespondError(w, http.StatusForbidden, "Admin access required")
				return
			}
		}

		// Decode request
		var req Req
		if r.Method != http.MethodGet && r.Method != http.MethodDelete {
			if err := DecodeJSON(r, &req); err != nil {
				logger.Error("failed to decode request", zap.Error(err))
				RespondError(w, http.StatusBadRequest, "Invalid request format")
				return
			}

			// Validate request
			if validator != nil {
				if err := validator.ValidateStruct(req); err != nil {
					logger.Error("validation failed", zap.Error(err))
					RespondError(w, http.StatusBadRequest, err.Error())
					return
				}
			}
		}

		// Execute use case
		result, err := useCase.Execute(ctx, req)
		if err != nil {
			handleUseCaseError(w, logger, err)
			return
		}

		// Log success
		logger.Info("operation completed successfully")

		// Respond with result
		RespondJSON(w, http.StatusOK, result)
	}
}

// handleUseCaseError maps domain errors to HTTP responses
func handleUseCaseError(w http.ResponseWriter, logger *zap.Logger, err error) {
	logger.Error("use case execution failed", zap.Error(err))

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
	default:
		RespondError(w, http.StatusInternalServerError, "Internal server error")
	}
}

// RespondError sends an error response
func RespondError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": message,
		"code":  statusCode,
	})
}

// RespondJSON sends a JSON response
func RespondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}
