package member

import (
	"errors"
	"net/http"
)

type Request struct {
	ID       string   `json:"id"`
	FullName string   `json:"fullName"`
	Books    []string `json:"books"`
}

func (req *Request) Bind(r *http.Request) error {
	if req.FullName == "" {
		return errors.New("fullName: cannot be blank")
	}
	return nil
}

type Response struct {
	ID       string   `json:"id"`
	FullName string   `json:"fullName"`
	Books    []string `json:"books"`
}

func ParseFromEntity(entity Entity) Response {
	return Response{
		ID:       entity.ID,
		FullName: *entity.FullName,
		Books:    entity.Books,
	}
}

func ParseFromEntities(entities []Entity) []Response {
	responses := make([]Response, len(entities))
	for i, entity := range entities {
		responses[i] = ParseFromEntity(entity)
	}
	return responses
}
