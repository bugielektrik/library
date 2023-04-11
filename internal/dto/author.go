package dto

import "library/internal/entity"

type AuthorRequest struct {
	ID        string
	FullName  string `json:"fullName" validate:"required"`
	Pseudonym string `json:"pseudonym" validate:"required"`
	Specialty string `json:"specialty" validate:"required"`
}

type AuthorResponse struct {
	ID        string `json:"ID"`
	FullName  string `json:"fullName"`
	Pseudonym string `json:"pseudonym"`
	Specialty string `json:"specialty"`
}

func ParseFromAuthor(src entity.Author) (dst AuthorResponse) {
	dst = AuthorResponse{
		ID:        src.ID,
		FullName:  *src.FullName,
		Pseudonym: *src.Pseudonym,
		Specialty: *src.Specialty,
	}

	return
}

func ParseFromAuthors(src []entity.Author) (dst []AuthorResponse) {
	for _, data := range src {
		dst = append(dst, ParseFromAuthor(data))
	}

	return
}
