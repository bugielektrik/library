package httputil

import (
	"encoding/json"
	"library-service/internal/pkg/errors"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// DecodeJSON decodes JSON request body into target struct.
// Returns ErrInvalidInput if decoding fails.
func DecodeJSON(r *http.Request, target interface{}) error {
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		return errors.ErrInvalidInput.Wrap(err)
	}
	return nil
}

// GetURLParam extracts and validates URL parameter from the request.
// Returns ErrInvalidInput if parameter is missing or empty.
func GetURLParam(r *http.Request, name string) (string, error) {
	value := chi.URLParam(r, name)
	if value == "" {
		return "", errors.ErrInvalidInput.WithDetails("field", name)
	}
	return value, nil
}

// MustGetURLParam extracts URL parameter, panics if empty.
// Use this only in routes where the parameter is guaranteed by the router.
func MustGetURLParam(r *http.Request, name string) string {
	value := chi.URLParam(r, name)
	if value == "" {
		panic("missing required URL parameter: " + name)
	}
	return value
}
