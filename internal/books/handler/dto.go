package http

import (
	"library-service/internal/books/domain/author"
	"library-service/internal/books/domain/book"
	"library-service/internal/books/service"
	"library-service/internal/pkg/strutil"
)

// ============================================================================
// Book DTOs
// ============================================================================

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

// ToBookResponseFromGet converts use case GetBookResponse to DTO BookResponse
func ToBookResponseFromGet(resp service.GetBookResponse) BookResponse {
	return BookResponse{
		ID:      resp.ID,
		Name:    resp.Name,
		Genre:   resp.Genre,
		ISBN:    resp.ISBN,
		Authors: resp.Authors,
	}
}

// ToBookResponseFromCreate converts use case CreateBookResponse to DTO BookResponse
func ToBookResponseFromCreate(resp service.CreateBookResponse) BookResponse {
	return BookResponse{
		ID:      resp.ID,
		Name:    resp.Name,
		Genre:   resp.Genre,
		ISBN:    resp.ISBN,
		Authors: resp.Authors,
	}
}

// ToBookResponses converts slice of use case GetBookResponse to slice of DTO BookResponse
func ToBookResponses(books []service.GetBookResponse) []BookResponse {
	responses := make([]BookResponse, len(books))
	for i, b := range books {
		responses[i] = ToBookResponseFromGet(b)
	}
	return responses
}

// ============================================================================
// Author DTOs
// ============================================================================

// CreateAuthorRequest represents the request to create a new author
type CreateAuthorRequest struct {
	FullName  string `json:"full_name" validate:"required,min=1,max=255"`
	Pseudonym string `json:"pseudonym" validate:"required,min=1,max=100"`
	Specialty string `json:"specialty" validate:"required,min=1,max=100"`
}

// UpdateAuthorRequest represents the request to update an existing author
type UpdateAuthorRequest struct {
	FullName  *string `json:"full_name,omitempty" validate:"omitempty,min=1,max=255"`
	Pseudonym *string `json:"pseudonym,omitempty" validate:"omitempty,min=1,max=100"`
	Specialty *string `json:"specialty,omitempty" validate:"omitempty,min=1,max=100"`
}

// AuthorResponse represents the response for an author
type AuthorResponse struct {
	ID        string `json:"id"`
	FullName  string `json:"full_name"`
	Pseudonym string `json:"pseudonym"`
	Specialty string `json:"specialty"`
}

// ToAuthorResponse converts use case service.AuthorResponse to DTO AuthorResponse
func ToAuthorResponse(resp service.AuthorResponse) AuthorResponse {
	return AuthorResponse{
		ID:        resp.ID,
		FullName:  resp.FullName,
		Pseudonym: resp.Pseudonym,
		Specialty: resp.Specialty,
	}
}

// ToAuthorResponses converts slice of use case service.AuthorResponse to slice of DTO AuthorResponse
func ToAuthorResponses(authors []service.AuthorResponse) []AuthorResponse {
	responses := make([]AuthorResponse, len(authors))
	for i, a := range authors {
		responses[i] = ToAuthorResponse(a)
	}
	return responses
}

// FromAuthorEntity converts domain author.Author to AuthorResponse
func FromAuthorEntity(entity author.Author) AuthorResponse {
	return AuthorResponse{
		ID:        entity.ID,
		FullName:  strutil.SafeString(entity.FullName),
		Pseudonym: strutil.SafeString(entity.Pseudonym),
		Specialty: strutil.SafeString(entity.Specialty),
	}
}

// FromAuthorEntities converts slice of domain author.Author to slice of AuthorResponse
func FromAuthorEntities(entities []author.Author) []AuthorResponse {
	result := make([]AuthorResponse, len(entities))
	for i, entity := range entities {
		result[i] = FromAuthorEntity(entity)
	}
	return result
}

// ============================================================================
// Domain Conversion Helpers (Legacy - Consider removing if unused)
// ============================================================================

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

// ToAuthorRequest converts CreateAuthorRequest to domain author.Request
func (r CreateAuthorRequest) ToAuthorRequest() author.Request {
	return author.Request{
		FullName:  r.FullName,
		Pseudonym: r.Pseudonym,
		Specialty: r.Specialty,
	}
}

// ToAuthorRequest converts UpdateAuthorRequest to domain author.Request
func (r UpdateAuthorRequest) ToAuthorRequest() author.Request {
	req := author.Request{}

	if r.FullName != nil {
		req.FullName = *r.FullName
	}
	if r.Pseudonym != nil {
		req.Pseudonym = *r.Pseudonym
	}
	if r.Specialty != nil {
		req.Specialty = *r.Specialty
	}

	return req
}
