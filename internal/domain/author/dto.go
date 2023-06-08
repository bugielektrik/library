package author

import (
	"errors"
	"net/http"
)

type Request struct {
	FullName  string `json:"fullName"`
	Pseudonym string `json:"pseudonym"`
	Specialty string `json:"specialty"`
}

func (s *Request) Bind(r *http.Request) error {
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

type Response struct {
	ID        string `json:"id"`
	FullName  string `json:"fullName"`
	Pseudonym string `json:"pseudonym"`
	Specialty string `json:"specialty"`
}

func ParseFromEntity(data Entity) (res Response) {
	res = Response{
		ID:        data.ID,
		FullName:  *data.FullName,
		Pseudonym: *data.Pseudonym,
		Specialty: *data.Specialty,
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
