package dto

import (
	"library/internal/entity"
)

type AuthorRequest struct {
	ID        string `json:"id"`
	FullName  string `json:"fullName" validate:"required"`
	Pseudonym string `json:"pseudonym" validate:"required"`
	Specialty string `json:"specialty" validate:"required"`
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
	for _, data := range data {
		res = append(res, ParseFromAuthor(data))
	}
	return
}
