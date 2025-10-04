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

// ParseFromAuthor converts an author to a response payload.
func ParseFromAuthor(data Author) Response {
	return Response{
		ID:        data.ID,
		FullName:  *data.FullName,
		Pseudonym: *data.Pseudonym,
		Specialty: *data.Specialty,
	}
}

// ParseFromAuthors converts a list of authors to a list of response payloads.
func ParseFromAuthors(data []Author) []Response {
	res := make([]Response, len(data))
	for i, author := range data {
		res[i] = ParseFromAuthor(author)
	}
	return res
}
