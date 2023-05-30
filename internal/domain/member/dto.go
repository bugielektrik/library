package member

import (
	"errors"
	"net/http"
)

type Request struct {
	ID       string   `json:"id"`
	FullName string   `json:"fullName" validate:"required"`
	Books    []string `json:"books" validate:"required"`
}

func (s *Request) Bind(r *http.Request) error {
	if s.FullName == "" {
		return errors.New("fullName: cannot be blank")
	}

	return nil
}

type Response struct {
	ID       string   `json:"id"`
	FullName string   `json:"fullName"`
	Books    []string `json:"books"`
}

func ParseFromEntity(data Entity) (res Response) {
	res = Response{
		ID:       data.ID,
		FullName: *data.FullName,
		Books:    data.Books,
	}
	return
}

func ParseFromEntities(data []Entity) (res []Response) {
	for _, object := range data {
		res = append(res, ParseFromEntity(object))
	}
	return
}
