package response

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/render"

	"library-service/pkg/store"
)

type Object struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

type HealthCheck struct {
	Commit   string            `json:"commit"`
	Database map[string]string `json:"database"`
	Version  string            `json:"version"`
}

func checkDB(name, dsn string) string {
	if strings.TrimSpace(dsn) == "" {
		return "down"
	}

	conn, err := store.NewSQL(dsn)
	if err != nil {
		return "down"
	}

	defer conn.Connection.Close()

	return "up"
}

func Health(w http.ResponseWriter, r *http.Request) {
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

func OK(w http.ResponseWriter, r *http.Request, data any) {
	render.Status(r, http.StatusOK)
	render.JSON(w, r, data)
}

func NoContent(w http.ResponseWriter, r *http.Request) {
	render.Status(r, http.StatusNoContent)
}

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

func Unauthorized(w http.ResponseWriter, r *http.Request, err error) {
	msg := "unauthorized"
	if err != nil {
		msg = err.Error()
	}

	render.Status(r, http.StatusUnauthorized)
	v := Object{
		Success: false,
		Message: msg,
	}
	render.JSON(w, r, v)
}

func Conflict(w http.ResponseWriter, r *http.Request, err error) {
	msg := "resource conflict"
	if err != nil {
		msg = err.Error()
	}

	render.Status(r, http.StatusConflict)
	v := Object{
		Success: false,
		Message: msg,
	}
	render.JSON(w, r, v)
}

func InternalServerError(w http.ResponseWriter, r *http.Request, err error, data any) {
	msg := "internal server error"
	if err != nil {
		msg = err.Error()
	}

	if err != nil && (errors.Is(err, context.DeadlineExceeded) || strings.Contains(err.Error(), "context deadline exceeded")) {
		switch r.Header.Get("Language") {
		case "RUS":
			msg = "Превышено время ожидания запроса"
		case "KAZ":
			msg = "Сұраудың күту уақыты асып кетті"
		default:
			msg = "Request timeout exceeded"
		}
	}

	render.Status(r, http.StatusInternalServerError)
	v := Object{
		Success: false,
		Data:    data,
		Message: msg,
	}
	render.JSON(w, r, v)
}
