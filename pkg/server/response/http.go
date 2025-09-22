package response

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/render"

	"library-service/pkg/store"
)

// Object is a generic response envelope.
type Object struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// HealthCheck contains basic service and DB health information.
type HealthCheck struct {
	Commit   string            `json:"commit"`
	Database map[string]string `json:"database"`
	Version  string            `json:"version"`
}

// checkDB attempts to open and immediately close a DB connection.
// Returns "up" or "down" and logs structured, lowercase messages.
func checkDB(name, dsn string) string {
	if strings.TrimSpace(dsn) == "" {
		return "down"
	}

	conn, err := store.NewSQL(dsn)
	if err != nil {
		return "down"
	}

	// Ensure connection is closed and log any close error.
	defer func() {
		if cerr := conn.Connection.Close(); cerr != nil {
			// Log the close error but don't return it to the caller.
			// We already returned "down" if we couldn't connect.
			log.Printf("health db=%s close_err=%v", name, cerr)
		}
	}()

	return "up"
}

// Health writes service health information including the status of configured DBs.
func Health(w http.ResponseWriter, r *http.Request) {
	// Build database health map using helper to keep logic concise.
	dbStatus := map[string]string{
		"postgres": checkDB("postgres", os.Getenv("POSTGRES_DSN")),
		"oracle":   checkDB("oracle", os.Getenv("ORACLE_DSN")),
	}

	health := HealthCheck{
		Commit:   os.Getenv("COMMIT_VERSION"),
		Database: dbStatus,
		Version:  "1.0.0",
	}

	OK(w, r, health)
}

// OK writes a 200 JSON response for successful operations.
func OK(w http.ResponseWriter, r *http.Request, data any) {
	render.Status(r, http.StatusOK)
	render.JSON(w, r, data)
}

// NoContent writes a 204 No Content response.
func NoContent(w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusNoContent)
	// No body for 204 responses.
}

// BadRequest writes a 400 response with a structured error body.
// It safely handles a nil error parameter.
func BadRequest(w http.ResponseWriter, r *http.Request, err error, data any) {
	msg := "bad request"
	if err != nil {
		msg = err.Error()
	}

	render.Status(r, http.StatusBadRequest)
	v := Object{
		Success: false,
		Data:    data,
		Message: msg,
	}
	render.JSON(w, r, v)
}

// NotFound writes a 404 response with a structured error body.
func NotFound(w http.ResponseWriter, r *http.Request, err error) {
	msg := "resource not found"
	if err != nil {
		msg = err.Error()
	}

	render.Status(r, http.StatusNotFound)
	v := Object{
		Success: false,
		Message: msg,
	}
	render.JSON(w, r, v)
}

// InternalServerError writes a 500 response and translates timeout errors
// into localized messages based on the request Language header.
// Uses errors.Is(context.DeadlineExceeded) when possible and falls back to string check.
func InternalServerError(w http.ResponseWriter, r *http.Request, err error, data any) {
	msg := "internal server error"
	if err != nil {
		msg = err.Error()
	}

	// Translate timeout messages if this was a deadline exceeded error.
	if err != nil && (errors.Is(err, context.DeadlineExceeded) || strings.Contains(err.Error(), "context deadline exceeded")) {
		switch r.Header.Get("Language") {
		case "RUS":
			msg = "Превышено время ожидания запроса"
		case "KAZ":
			msg = "Сұраудың күту уақыты асып кетті"
		default:
			msg = "Request timeout exceeded"
		}
		// Log that we translated a timeout error
	}

	render.Status(r, http.StatusInternalServerError)
	v := Object{
		Success: false,
		Data:    data,
		Message: msg,
	}
	render.JSON(w, r, v)
}
