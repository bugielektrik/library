package http

import (
	"library-service/internal/members/domain"
	"library-service/internal/pkg/strutil"
)

// ============================================================================
// Member DTOs
// ============================================================================

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
	Email    string   `json:"email,omitempty"`
	FullName string   `json:"full_name"`
	Role     string   `json:"role,omitempty"`
	Books    []string `json:"books"`
}

// FromMemberEntity converts domain domain.Member to MemberResponse
func FromMemberEntity(entity domain.Member) MemberResponse {
	return MemberResponse{
		ID:       entity.ID,
		Email:    entity.Email,
		FullName: strutil.SafeString(entity.FullName),
		Role:     string(entity.Role),
		Books:    entity.Books,
	}
}

// FromMemberEntities converts slice of domain domain.Member to slice of MemberResponse
func FromMemberEntities(entities []domain.Member) []MemberResponse {
	result := make([]MemberResponse, len(entities))
	for i, entity := range entities {
		result[i] = FromMemberEntity(entity)
	}
	return result
}

// ============================================================================
// Domain Conversion Helpers (Legacy - Consider removing if unused)
// ============================================================================

// ToMemberRequest converts CreateMemberRequest to domain domain.Request
func (r CreateMemberRequest) ToMemberRequest() domain.Request {
	return domain.Request{
		FullName: r.FullName,
		Books:    r.Books,
	}
}

// ToMemberRequest converts UpdateMemberRequest to domain domain.Request
func (r UpdateMemberRequest) ToMemberRequest() domain.Request {
	req := domain.Request{}

	if r.FullName != nil {
		req.FullName = *r.FullName
	}
	if r.Books != nil {
		req.Books = r.Books
	}

	return req
}
