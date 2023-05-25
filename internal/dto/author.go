package dto

import (
	"errors"
	"net/http"

	"library/internal/entity"
)

type AuthorRequest struct {
	FullName  string `json:"fullName"`
	Pseudonym string `json:"pseudonym"`
	Specialty string `json:"specialty"`
}

func (s *AuthorRequest) Bind(r *http.Request) error {
	if s.FullName == "" {
		return errors.New("phone: cannot be blank")
	}

	if s.Pseudonym == "" {
		return errors.New("pseudonym: cannot be blank")
	}

	if s.Specialty == "" {
		return errors.New("specialty: cannot be blank")
	}

	return nil
}

type AuthorResponse struct {
	ID        string `json:"id"`
	FullName  string `json:"fullName"`
	Pseudonym string `json:"pseudonym"`
	Specialty string `json:"specialty"`
}

func ParseFromAuthor(data entity.Author) (res AuthorResponse) {
	res = AuthorResponse{
		ID:        data.ID,
		FullName:  *data.FullName,
		Pseudonym: *data.Pseudonym,
		Specialty: *data.Specialty,
	}
	return
}

func ParseFromAuthors(data []entity.Author) (res []AuthorResponse) {
	for _, object := range data {
		res = append(res, ParseFromAuthor(object))
	}
	return
}
