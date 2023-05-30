package status

import (
	"net/http"
)

type Response struct {
	Status  int    `json:"-"`
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

func (e *Response) Render(w http.ResponseWriter, r *http.Request) error {
	w.WriteHeader(e.Status)
	return nil
}

func OK(data any) Response {
	return Response{
		Status:  http.StatusOK,
		Success: true,
		Data:    data,
	}
}

func BadRequest(err error, data any) Response {
	return Response{
		Status:  http.StatusBadRequest,
		Success: false,
		Message: err.Error(),
		Data:    data,
	}
}

func InternalServerError(err error) Response {
	return Response{
		Status:  http.StatusInternalServerError,
		Success: false,
		Message: err.Error(),
	}
}
