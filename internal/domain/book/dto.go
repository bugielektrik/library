package book

import (
	"errors"
	"net/http"
)

type Request struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Genre   string   `json:"genre"`
	ISBN    string   `json:"isbn"`
	Authors []string `json:"authors"`
}

func (s *Request) Bind(r *http.Request) error {
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

type Response struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Genre   string   `json:"genre"`
	ISBN    string   `json:"isbn"`
	Authors []string `json:"authors"`
}

func ParseFromEntity(data Entity) (res Response) {
	res = Response{
		ID:      data.ID,
		Name:    *data.Name,
		Genre:   *data.Genre,
		ISBN:    *data.ISBN,
		Authors: data.Authors,
	}
	return
}

func ParseFromEntities(data []Entity) (res []Response) {
	res = make([]Response, 0)
	for _, object := range data {
		res = append(res, ParseFromEntity(object))
	}
	return
}
