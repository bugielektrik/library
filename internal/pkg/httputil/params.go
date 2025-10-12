package httputil

import (
	"context"
	"library-service/internal/pkg/logutil"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// ParamExtractor extracts parameters from the request
type ParamExtractor func(r *http.Request) (map[string]string, error)

// URLParamExtractor extracts URL parameters
func URLParamExtractor(params ...string) ParamExtractor {
	return func(r *http.Request) (map[string]string, error) {
		result := make(map[string]string)
		for _, param := range params {
			value := chi.URLParam(r, param)
			if value == "" {
				return nil, ErrMissingParameter{Name: param}
			}
			result[param] = value
		}
		return result, nil
	}
}

// QueryParamExtractor extracts query parameters
func QueryParamExtractor(params ...string) ParamExtractor {
	return func(r *http.Request) (map[string]string, error) {
		result := make(map[string]string)
		for _, param := range params {
			value := r.URL.Query().Get(param)
			if value != "" {
				result[param] = value
			}
		}
		return result, nil
	}
}

// ErrMissingParameter indicates a required parameter is missing
type ErrMissingParameter struct {
	Name string
}

func (e ErrMissingParameter) Error() string {
	return "missing required parameter: " + e.Name
}

// CreateParamHandler creates a handler that extracts parameters before execution
func CreateParamHandler[Req, Res any](
	buildRequest func(params map[string]string) (Req, error),
	executor SimpleExecutor[Req, Res],
	validate func(req Req) error,
	loggerName, operationName string,
	opts WrapperOptions,
	extractor ParamExtractor,
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

		// Extract parameters
		params, err := extractor(r)
		if err != nil {
			logger.Error("failed to extract parameters", zap.Error(err))
			RespondError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Build request from parameters
		req, err := buildRequest(params)
		if err != nil {
			logger.Error("failed to build request", zap.Error(err))
			RespondError(w, http.StatusBadRequest, err.Error())
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

// CreateGetHandler creates a handler for GET requests with no body
func CreateGetHandler[Res any](
	executor func(ctx context.Context) (Res, error),
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

		// Execute the function
		result, err := executor(ctx)
		if err != nil {
			handleError(w, logger, err)
			return
		}

		logger.Info("operation completed successfully")
		RespondJSON(w, http.StatusOK, result)
	}
}
