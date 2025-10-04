package member

import (
	"errors"
	"net/http"
)

// Request represents the request payload for member operations.
type Request struct {
	ID       string   `json:"id"`
	FullName string   `json:"fullName"`
	Books    []string `json:"books"`
}

// Bind validates the request payload.
func (req *Request) Bind(r *http.Request) error {
	if req.FullName == "" {
		return errors.New("fullName: cannot be blank")
	}
	return nil
}

// Response represents the response payload for member operations.
type Response struct {
	ID       string   `json:"id"`
	FullName string   `json:"fullName"`
	Books    []string `json:"books"`
}

// ParseFromMember creates a new Response from a given Member.
func ParseFromMember(member Member) Response {
	return Response{
		ID:       member.ID,
		FullName: *member.FullName,
		Books:    member.Books,
	}
}

// ParseFromMembers creates a slice of Responses from a slice of Members.
func ParseFromMembers(members []Member) []Response {
	responses := make([]Response, len(members))
	for i, member := range members {
		responses[i] = ParseFromMember(member)
	}
	return responses
}
