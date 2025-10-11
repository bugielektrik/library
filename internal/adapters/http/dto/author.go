package dto

import (
	"library-service/internal/domain/author"
	"library-service/internal/usecase/bookops"
	"library-service/pkg/strutil"
)

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

// FromAuthorEntity converts domain author.Author to AuthorResponse
func FromAuthorEntity(entity author.Author) AuthorResponse {
	return AuthorResponse{
		ID:        entity.ID,
		FullName:  strutil.SafeString(entity.FullName),
		Pseudonym: strutil.SafeString(entity.Pseudonym),
		Specialty: strutil.SafeString(entity.Specialty),
	}
}

// FromAuthorResponse converts domain author.Response to AuthorResponse
func FromAuthorResponse(resp author.Response) AuthorResponse {
	return AuthorResponse{
		ID:        resp.ID,
		FullName:  resp.FullName,
		Pseudonym: resp.Pseudonym,
		Specialty: resp.Specialty,
	}
}

// FromAuthorResponses converts slice of domain author.Response to slice of AuthorResponse
func FromAuthorResponses(responses []author.Response) []AuthorResponse {
	result := make([]AuthorResponse, len(responses))
	for i, resp := range responses {
		result[i] = FromAuthorResponse(resp)
	}
	return result
}

// ToAuthorResponse converts use case bookops.AuthorResponse to DTO AuthorResponse
func ToAuthorResponse(resp bookops.AuthorResponse) AuthorResponse {
	return AuthorResponse{
		ID:        resp.ID,
		FullName:  resp.FullName,
		Pseudonym: resp.Pseudonym,
		Specialty: resp.Specialty,
	}
}

// ToAuthorResponses converts slice of use case bookops.AuthorResponse to slice of DTO AuthorResponse
func ToAuthorResponses(authors []bookops.AuthorResponse) []AuthorResponse {
	responses := make([]AuthorResponse, len(authors))
	for i, a := range authors {
		responses[i] = ToAuthorResponse(a)
	}
	return responses
}

// FromAuthorEntities converts slice of domain author.Author to slice of AuthorResponse
func FromAuthorEntities(entities []author.Author) []AuthorResponse {
	result := make([]AuthorResponse, len(entities))
	for i, entity := range entities {
		result[i] = FromAuthorEntity(entity)
	}
	return result
}
