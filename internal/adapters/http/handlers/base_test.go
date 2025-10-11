package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	pkgerrors "library-service/pkg/errors"
	"library-service/pkg/httputil"
)

func TestBaseHandler_RespondJSON(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		data       interface{}
		wantStatus int
		wantHeader string
	}{
		{
			name:       "success response",
			status:     http.StatusOK,
			data:       map[string]string{"message": "success"},
			wantStatus: http.StatusOK,
			wantHeader: httputil.ContentTypeJSON,
		},
		{
			name:       "created response",
			status:     http.StatusCreated,
			data:       map[string]string{"id": "123"},
			wantStatus: http.StatusCreated,
			wantHeader: httputil.ContentTypeJSON,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := BaseHandler{}
			w := httptest.NewRecorder()

			handler.RespondJSON(w, tt.status, tt.data)

			if w.Code != tt.wantStatus {
				t.Errorf("RespondJSON() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if got := w.Header().Get("Content-Type"); got != tt.wantHeader {
				t.Errorf("RespondJSON() Content-Type = %v, want %v", got, tt.wantHeader)
			}

			var response map[string]string
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}
		})
	}
}

func TestBaseHandler_RespondError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
	}{
		{
			name:       "not found error",
			err:        pkgerrors.ErrNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "validation error",
			err:        pkgerrors.ErrValidation,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "internal error",
			err:        pkgerrors.ErrDatabase,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "generic error",
			err:        errors.New("some error"),
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := BaseHandler{}
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/", nil)

			handler.RespondError(w, r, tt.err)

			if w.Code != tt.wantStatus {
				t.Errorf("RespondError() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if got := w.Header().Get("Content-Type"); got != httputil.ContentTypeJSON {
				t.Errorf("RespondError() Content-Type = %v, want %v", got, httputil.ContentTypeJSON)
			}
		})
	}
}
