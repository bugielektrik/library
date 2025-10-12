package handlers

import (
	"encoding/json"
	errors2 "library-service/internal/pkg/errors"
	httputil2 "library-service/internal/pkg/httputil"
	"library-service/internal/pkg/middleware"
	"net/http"

	"go.uber.org/zap"

	"library-service/internal/infrastructure/log"
)

// BaseHandler provides common handler functionality that can be embedded in all handler.
//
// This eliminates duplication of respondError and respondJSON methods across handler.
// Handlers can embed BaseHandler to inherit these common response methods.
//
// Example usage:
//
//	type BookHandler struct {
//	    BaseHandler  // Embed base handler
//	    createBookUC *bookops.CreateBookUseCase
//	}
//
//	func (h *BookHandler) create(w http.ResponseWriter, r *http.Request) {
//	    // ...
//	    if err != nil {
//	        h.RespondError(w, r, err)
//	        return
//	    }
//	    h.RespondJSON(w, http.StatusCreated, response)
//	}
type BaseHandler struct{}

// RespondError sends an error response with appropriate logging based on status code.
//
// Server errors (5xx) are logged at ERROR level with full error details.
// Client errors (4xx) are logged at WARN level.
// The error is converted to a DTO error response before sending.
func (b *BaseHandler) RespondError(w http.ResponseWriter, r *http.Request, err error) {
	logger := log.FromContext(r.Context())
	status := errors2.GetHTTPStatus(err)

	if httputil2.IsServerError(status) {
		logger.Error("internal error", zap.Error(err))
	} else {
		logger.Warn("request error", zap.Error(err))
	}

	response := errors2.FromError(err)
	b.RespondJSON(w, status, response)
}

// RespondJSON sends a JSON response with the specified status code.
//
// Sets Content-Type to application/json and encodes the data as JSON.
// Logs encoding errors but doesn't fail the response (status code already sent).
func (b *BaseHandler) RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set(httputil2.HeaderContentType, httputil2.ContentTypeJSON)
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		// Status already sent, can't change response, just log
		// Note: Using zap.L() for global logger since we don't have context here
		zap.L().Error("failed to encode JSON response", zap.Error(err))
	}
}

// GetMemberID extracts the member ID from the request context.
//
// Returns the member ID and true if found, or empty string and false if not found.
// If not found, automatically sends an unauthorized error response.
// Callers should check the bool return value and return early if false.
//
// Example usage:
//
//	memberID, ok := h.GetMemberID(w, r)
//	if !ok {
//	    return // Error already sent
//	}
//	// Use memberID...
func (b *BaseHandler) GetMemberID(w http.ResponseWriter, r *http.Request) (string, bool) {
	memberID, ok := middleware.GetMemberIDFromContext(r.Context())
	if !ok {
		b.RespondError(w, r, errors2.ErrUnauthorized)
		return "", false
	}
	return memberID, true
}

// GetURLParam extracts a URL parameter from the request.
//
// Returns the parameter value and true if extraction succeeds, or empty string and false if it fails.
// If extraction fails, automatically sends an appropriate error response.
// Callers should check the bool return value and return early if false.
//
// Example usage:
//
//	id, ok := h.GetURLParam(w, r, "id")
//	if !ok {
//	    return // Error already sent
//	}
//	// Use id...
func (b *BaseHandler) GetURLParam(w http.ResponseWriter, r *http.Request, paramName string) (string, bool) {
	value, err := httputil2.GetURLParam(r, paramName)
	if err != nil {
		b.RespondError(w, r, err)
		return "", false
	}
	return value, true
}
