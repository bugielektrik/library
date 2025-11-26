package interfaces

import (
	"context"
	"library-service/internal/domain/book"
	"library-service/internal/domain/member"
)

type MemberService interface {
	ListMembers(ctx context.Context) ([]member.Response, error)
	CreateMember(ctx context.Context, req member.Request) (member.Response, error)
	GetMember(ctx context.Context, id string) (member.Response, error)
	UpdateMember(ctx context.Context, id string, req member.Request) error
	DeleteMember(ctx context.Context, id string) error
	ListMemberBooks(ctx context.Context, memberID string) ([]book.Response, error)
}
