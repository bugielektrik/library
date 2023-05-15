package dto

import (
	"library/internal/entity"
)

type BookRequest struct {
	ID      string   `json:"id"`
	Name    string   `json:"name" validate:"required"`
	Genre   string   `json:"genre" validate:"required"`
	ISBN    string   `json:"isbn" validate:"required"`
	Authors []string `json:"authors" validate:"required"`
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
	for _, data := range data {
		res = append(res, ParseFromBook(data))
	}
	return
}
