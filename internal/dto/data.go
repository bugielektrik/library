package dto

import (
	"github.com/go-chi/render"
	"net/http"
)

type Data struct {
	Error  error `json:"-"`
	Status int   `json:"-"`

	Success bool   `json:"success,omitempty" json:"success,omitempty"`
	Message string `json:"message,omitempty" json:"message,omitempty"`
	Data    any    `json:"data,omitempty" json:"data,omitempty" json:"data,omitempty"`
}

func (s Data) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, s.Status)
	return nil
}
