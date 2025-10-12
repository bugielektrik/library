package httputil

import "testing"

func TestIsServerError(t *testing.T) {
	tests := []struct {
		name string
		code int
		want bool
	}{
		{"500 Internal Server Error", 500, true},
		{"502 Bad Gateway", 502, true},
		{"503 Service Unavailable", 503, true},
		{"404 Not Found", 404, false},
		{"200 OK", 200, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsServerError(tt.code); got != tt.want {
				t.Errorf("IsServerError(%d) = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}

func TestIsClientError(t *testing.T) {
	tests := []struct {
		name string
		code int
		want bool
	}{
		{"400 Bad Request", 400, true},
		{"404 Not Found", 404, true},
		{"422 Unprocessable Entity", 422, true},
		{"500 Internal Server Error", 500, false},
		{"200 OK", 200, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsClientError(tt.code); got != tt.want {
				t.Errorf("IsClientError(%d) = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}

func TestIsSuccess(t *testing.T) {
	tests := []struct {
		name string
		code int
		want bool
	}{
		{"200 OK", 200, true},
		{"201 Created", 201, true},
		{"204 No Content", 204, true},
		{"301 Moved Permanently", 301, false},
		{"404 Not Found", 404, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsSuccess(tt.code); got != tt.want {
				t.Errorf("IsSuccess(%d) = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}

func TestIsRedirect(t *testing.T) {
	tests := []struct {
		name string
		code int
		want bool
	}{
		{"301 Moved Permanently", 301, true},
		{"302 Found", 302, true},
		{"304 Not Modified", 304, true},
		{"200 OK", 200, false},
		{"404 Not Found", 404, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRedirect(tt.code); got != tt.want {
				t.Errorf("IsRedirect(%d) = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}

func TestIsError(t *testing.T) {
	tests := []struct {
		name string
		code int
		want bool
	}{
		{"400 Bad Request", 400, true},
		{"404 Not Found", 404, true},
		{"500 Internal Server Error", 500, true},
		{"503 Service Unavailable", 503, true},
		{"200 OK", 200, false},
		{"301 Moved Permanently", 301, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsError(tt.code); got != tt.want {
				t.Errorf("IsError(%d) = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}
