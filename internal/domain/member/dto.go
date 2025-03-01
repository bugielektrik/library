package member

import (
	"errors"
	"net/http"
)

// Request represents the request payload for member operations.
type Request struct {
	ID       string   `json:"id"`
	FullName string   `json:"fullName"`
	Books    []string `json:"books"`
}

// Bind validates the request payload.
func (req *Request) Bind(r *http.Request) error {
	if req.FullName == "" {
		return errors.New("fullName: cannot be blank")
	}
	return nil
}

// Response represents the response payload for member operations.
type Response struct {
	ID       string   `json:"id"`
	FullName string   `json:"fullName"`
	Books    []string `json:"books"`
}

// ParseFromEntity creates a new Response from a given Entity.
func ParseFromEntity(entity Entity) Response {
	return Response{
		ID:       entity.ID,
		FullName: *entity.FullName,
		Books:    entity.Books,
	}
}

// ParseFromEntities creates a slice of Responses from a slice of Entities.
func ParseFromEntities(entities []Entity) []Response {
	responses := make([]Response, len(entities))
	for i, entity := range entities {
		responses[i] = ParseFromEntity(entity)
	}
	return responses
}
