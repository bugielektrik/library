package dto

import (
	"library/internal/entity"
)

type MemberRequest struct {
	ID       string   `json:"id"`
	FullName string   `json:"fullName" validate:"required"`
	Books    []string `json:"books" validate:"required"`
}

type MemberResponse struct {
	ID       string   `json:"id"`
	FullName string   `json:"fullName"`
	Books    []string `json:"books"`
}

func ParseFromMember(data entity.Member) (res MemberResponse) {
	res = MemberResponse{
		ID:       data.ID,
		FullName: *data.FullName,
		Books:    data.Books,
	}
	return
}

func ParseFromMembers(data []entity.Member) (res []MemberResponse) {
	for _, object := range data {
		res = append(res, ParseFromMember(object))
	}
	return
}
