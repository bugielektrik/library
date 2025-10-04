package dto

import (
	"library-service/internal/domain/member"
)

// CreateMemberRequest represents the request to create a new member
type CreateMemberRequest struct {
	FullName string   `json:"full_name" validate:"required,min=1,max=200"`
	Books    []string `json:"books,omitempty"`
}

// UpdateMemberRequest represents the request to update an existing member
type UpdateMemberRequest struct {
	FullName *string  `json:"full_name,omitempty" validate:"omitempty,min=1,max=200"`
	Books    []string `json:"books,omitempty"`
}

// MemberResponse represents the response for a member
type MemberResponse struct {
	ID       string   `json:"id"`
	FullName string   `json:"full_name"`
	Books    []string `json:"books"`
}

// ToMemberRequest converts CreateMemberRequest to domain member.Request
func (r CreateMemberRequest) ToMemberRequest() member.Request {
	return member.Request{
		FullName: r.FullName,
		Books:    r.Books,
	}
}

// ToMemberRequest converts UpdateMemberRequest to domain member.Request
func (r UpdateMemberRequest) ToMemberRequest() member.Request {
	req := member.Request{}

	if r.FullName != nil {
		req.FullName = *r.FullName
	}
	if r.Books != nil {
		req.Books = r.Books
	}

	return req
}

// FromMemberEntity converts domain member.Entity to MemberResponse
func FromMemberEntity(entity member.Entity) MemberResponse {
	return MemberResponse{
		ID:       entity.ID,
		FullName: safeString(entity.FullName),
		Books:    entity.Books,
	}
}

// FromMemberResponse converts domain member.Response to MemberResponse
func FromMemberResponse(resp member.Response) MemberResponse {
	return MemberResponse{
		ID:       resp.ID,
		FullName: resp.FullName,
		Books:    resp.Books,
	}
}

// FromMemberResponses converts slice of domain member.Response to slice of MemberResponse
func FromMemberResponses(responses []member.Response) []MemberResponse {
	result := make([]MemberResponse, len(responses))
	for i, resp := range responses {
		result[i] = FromMemberResponse(resp)
	}
	return result
}
