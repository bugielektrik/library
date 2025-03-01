package author

import (
	"errors"
	"net/http"
)

// Request represents the request payload for author operations.
type Request struct {
	FullName  string `json:"fullName"`
	Pseudonym string `json:"pseudonym"`
	Specialty string `json:"specialty"`
}

// Bind validates the request payload.
func (s *Request) Bind(r *http.Request) error {
	if s.FullName == "" {
		return errors.New("fullname: cannot be blank")
	}

	if s.Pseudonym == "" {
		return errors.New("pseudonym: cannot be blank")
	}

	if s.Specialty == "" {
		return errors.New("specialty: cannot be blank")
	}

	return nil
}

// Response represents the response payload for author operations.
type Response struct {
	ID        string `json:"id"`
	FullName  string `json:"fullName"`
	Pseudonym string `json:"pseudonym"`
	Specialty string `json:"specialty"`
}

// ParseFromEntity converts an author entity to a response payload.
func ParseFromEntity(data Entity) Response {
	return Response{
		ID:        data.ID,
		FullName:  *data.FullName,
		Pseudonym: *data.Pseudonym,
		Specialty: *data.Specialty,
	}
}

// ParseFromEntities converts a list of author entities to a list of response payloads.
func ParseFromEntities(data []Entity) []Response {
	res := make([]Response, len(data))
	for i, entity := range data {
		res[i] = ParseFromEntity(entity)
	}
	return res
}
