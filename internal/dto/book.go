package dto

import (
	"errors"
	"net/http"

	"library/internal/entity"
)

type BookRequest struct {
	ID      string   `json:"id"`
	Name    string   `json:"name" validate:"required"`
	Genre   string   `json:"genre" validate:"required"`
	ISBN    string   `json:"isbn" validate:"required"`
	Authors []string `json:"authors" validate:"required"`
}

func (s *BookRequest) Bind(r *http.Request) error {
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

type BookResponse struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Genre   string   `json:"genre"`
	ISBN    string   `json:"isbn"`
	Authors []string `json:"authors"`
}

func ParseFromBook(data entity.Book) (res BookResponse) {
	res = BookResponse{
		ID:      data.ID,
		Name:    *data.Name,
		Genre:   *data.Genre,
		ISBN:    *data.ISBN,
		Authors: data.Authors,
	}
	return
}

func ParseFromBooks(data []entity.Book) (res []BookResponse) {
	for _, object := range data {
		res = append(res, ParseFromBook(object))
	}
	return
}
