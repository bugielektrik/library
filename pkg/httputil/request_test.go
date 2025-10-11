package httputil

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"library-service/pkg/errors"
)

func TestDecodeJSON(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{
			name:    "valid JSON",
			body:    `{"name":"test","value":123}`,
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			body:    `{"name":"test"`,
			wantErr: true,
		},
		{
			name:    "empty body",
			body:    ``,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(tt.body))
			var target map[string]interface{}

			err := DecodeJSON(req, &target)

			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeJSON() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && !errors.Is(err, errors.ErrInvalidInput) {
				t.Errorf("DecodeJSON() should return ErrInvalidInput, got %v", err)
			}
		})
	}
}

func TestGetURLParam(t *testing.T) {
	tests := []struct {
		name      string
		paramName string
		paramVal  string
		wantErr   bool
		wantValue string
	}{
		{
			name:      "valid parameter",
			paramName: "id",
			paramVal:  "123",
			wantErr:   false,
			wantValue: "123",
		},
		{
			name:      "empty parameter",
			paramName: "id",
			paramVal:  "",
			wantErr:   true,
			wantValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create chi context with URL param
			r := chi.NewRouter()
			r.Get("/{id}", func(w http.ResponseWriter, req *http.Request) {
				value, err := GetURLParam(req, tt.paramName)

				if (err != nil) != tt.wantErr {
					t.Errorf("GetURLParam() error = %v, wantErr %v", err, tt.wantErr)
				}

				if value != tt.wantValue {
					t.Errorf("GetURLParam() = %v, want %v", value, tt.wantValue)
				}

				if err != nil && !errors.Is(err, errors.ErrInvalidInput) {
					t.Errorf("GetURLParam() should return ErrInvalidInput, got %v", err)
				}
			})

			// Execute request through router
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/"+tt.paramVal, nil)
			r.ServeHTTP(w, req)
		})
	}
}

func TestMustGetURLParam(t *testing.T) {
	t.Run("valid parameter", func(t *testing.T) {
		r := chi.NewRouter()
		r.Get("/{id}", func(w http.ResponseWriter, req *http.Request) {
			value := MustGetURLParam(req, "id")
			if value != "test-id" {
				t.Errorf("MustGetURLParam() = %v, want test-id", value)
			}
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/test-id", nil)
		r.ServeHTTP(w, req)
	})

	t.Run("empty parameter panics", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("MustGetURLParam() should panic on empty parameter")
			}
		}()

		// Create a request without chi context (simulates missing param)
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		_ = MustGetURLParam(req, "nonexistent")
	})
}
