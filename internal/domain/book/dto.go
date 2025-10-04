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

// ParseFromBook converts a book to a response payload.
func ParseFromBook(data Book) Response {
	return Response{
		ID:      data.ID,
		Name:    *data.Name,
		Genre:   *data.Genre,
		ISBN:    *data.ISBN,
		Authors: data.Authors,
	}
}

// ParseFromBooks converts a list of books to a list of response payloads.
func ParseFromBooks(data []Book) []Response {
	res := make([]Response, len(data))
	for i, book := range data {
		res[i] = ParseFromBook(book)
	}
	return res
}
