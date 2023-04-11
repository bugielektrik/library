package dto

import "library/internal/entity"

type MemberRequest struct {
	ID       string
	FullName string `json:"fullName" validate:"required"`
	Books    string `json:"books" validate:"required"`
}

type MemberResponse struct {
	ID       string `json:"id"`
	FullName string `json:"fullName"`
	Books    string `json:"books"`
}

func ParseFromMember(src entity.Member) (dst MemberResponse) {
	dst = MemberResponse{
		ID:       src.ID,
		FullName: *src.FullName,
		Books:    *src.Books,
	}

	return
}

func ParseFromMembers(src []entity.Member) (dst []MemberResponse) {
	for _, data := range src {
		dst = append(dst, ParseFromMember(data))
	}

	return
}
