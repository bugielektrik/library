package dto

import "library/internal/entity"

type BookRequest struct {
	ID      string
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

func ParseFromBook(src entity.Book) (dst BookResponse) {
	dst = BookResponse{
		ID:      src.ID,
		Name:    *src.Name,
		Genre:   *src.Genre,
		ISBN:    *src.ISBN,
		Authors: src.Authors,
	}

	return
}

func ParseFromBooks(src []entity.Book) (dst []BookResponse) {
	for _, data := range src {
		dst = append(dst, ParseFromBook(data))
	}

	return
}
