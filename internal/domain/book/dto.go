package book

import (
	"errors"
	"net/http"
)

// Request represents the request payload for book operations.
type Request struct {
	Name    string   `json:"name"`
	Genre   string   `json:"genre"`
	ISBN    string   `json:"isbn"`
	Authors []string `json:"authors"`
}

// Bind validates the request payload.
func (s *Request) Bind(r *http.Request) error {
	if s.Name == "" {
		return errors.New("name: cannot be blank")
	}

	if s.Genre == "" {
		return errors.New("genre: cannot be blank")
	}

	if s.ISBN == "" {
		return errors.New("isbn: cannot be blank")
	}

	return nil
}

// Response represents the response payload for book operations.
type Response struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Genre   string   `json:"genre"`
	ISBN    string   `json:"isbn"`
	Authors []string `json:"authors"`
}

// ParseFromEntity converts a book entity to a response payload.
func ParseFromEntity(data Entity) Response {
	return Response{
		ID:      data.ID,
		Name:    *data.Name,
		Genre:   *data.Genre,
		ISBN:    *data.ISBN,
		Authors: data.Authors,
	}
}

// ParseFromEntities converts a list of book entities to a list of response payloads.
func ParseFromEntities(data []Entity) []Response {
	res := make([]Response, len(data))
	for i, entity := range data {
		res[i] = ParseFromEntity(entity)
	}
	return res
}
