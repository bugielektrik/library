package dto

import (
	"library-service/internal/domain/book"
)

// CreateBookRequest represents the request to create a new book
type CreateBookRequest struct {
	Name    string   `json:"name" validate:"required,min=1,max=255"`
	Genre   string   `json:"genre" validate:"required,min=1,max=100"`
	ISBN    string   `json:"isbn" validate:"required,isbn"`
	Authors []string `json:"authors" validate:"required,min=1,dive,uuid4"`
}

// UpdateBookRequest represents the request to update an existing book
type UpdateBookRequest struct {
	Name    *string  `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Genre   *string  `json:"genre,omitempty" validate:"omitempty,min=1,max=100"`
	ISBN    *string  `json:"isbn,omitempty" validate:"omitempty,isbn"`
	Authors []string `json:"authors,omitempty" validate:"omitempty,min=1,dive,uuid4"`
}

// BookResponse represents the response for a book
type BookResponse struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Genre   string   `json:"genre"`
	ISBN    string   `json:"isbn"`
	Authors []string `json:"authors"`
}

// ToBookRequest converts CreateBookRequest to domain book.Request
func (r CreateBookRequest) ToBookRequest() book.Request {
	return book.Request{
		Name:    r.Name,
		Genre:   r.Genre,
		ISBN:    r.ISBN,
		Authors: r.Authors,
	}
}

// ToBookRequest converts UpdateBookRequest to domain book.Request
func (r UpdateBookRequest) ToBookRequest() book.Request {
	req := book.Request{}

	if r.Name != nil {
		req.Name = *r.Name
	}
	if r.Genre != nil {
		req.Genre = *r.Genre
	}
	if r.ISBN != nil {
		req.ISBN = *r.ISBN
	}
	if r.Authors != nil {
		req.Authors = r.Authors
	}

	return req
}

// FromBookEntity converts domain book.Book to BookResponse
func FromBookEntity(entity book.Book) BookResponse {
	return BookResponse{
		ID:      entity.ID,
		Name:    safeString(entity.Name),
		Genre:   safeString(entity.Genre),
		ISBN:    safeString(entity.ISBN),
		Authors: entity.Authors,
	}
}

// FromBookResponse converts domain book.Response to BookResponse
func FromBookResponse(resp book.Response) BookResponse {
	return BookResponse{
		ID:      resp.ID,
		Name:    resp.Name,
		Genre:   resp.Genre,
		ISBN:    resp.ISBN,
		Authors: resp.Authors,
	}
}

// FromBookResponses converts slice of domain book.Response to slice of BookResponse
func FromBookResponses(responses []book.Response) []BookResponse {
	result := make([]BookResponse, len(responses))
	for i, resp := range responses {
		result[i] = FromBookResponse(resp)
	}
	return result
}

// safeString returns the value of a string pointer or empty string if nil
func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
