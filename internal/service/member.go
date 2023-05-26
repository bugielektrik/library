package service

import (
	"context"

	"library/internal/dto"
	"library/internal/entity"
	"library/internal/repository"
)

type Member struct {
	memberRepository repository.Member
	bookService      Book
}

func NewMemberService(m repository.Member, b Book) Member {
	return Member{
		memberRepository: m,
		bookService:      b,
	}
}

func (s *Member) List(ctx context.Context) (res []dto.MemberResponse, err error) {
	data, err := s.memberRepository.SelectRows(ctx)
	if err != nil {
		return
	}
	res = dto.ParseFromMembers(data)

	return
}

func (s *Member) Add(ctx context.Context, req dto.MemberRequest) (res dto.MemberResponse, err error) {
	data := entity.Member{
		FullName: &req.FullName,
		Books:    req.Books,
	}

	data.ID, err = s.memberRepository.CreateRow(ctx, data)
	if err != nil {
		return
	}
	res = dto.ParseFromMember(data)

	return
}

func (s *Member) Get(ctx context.Context, id string) (res dto.MemberResponse, err error) {
	data, err := s.memberRepository.GetRow(ctx, id)
	if err != nil {
		return
	}
	res = dto.ParseFromMember(data)

	return
}

func (s *Member) Update(ctx context.Context, id string, req dto.MemberRequest) (err error) {
	data := entity.Member{
		FullName: &req.FullName,
		Books:    req.Books,
	}
	return s.memberRepository.UpdateRow(ctx, id, data)
}

func (s *Member) Delete(ctx context.Context, id string) (err error) {
	return s.memberRepository.DeleteRow(ctx, id)
}

func (s *Member) ListBook(ctx context.Context, id string) (res []dto.BookResponse, err error) {
	data, err := s.memberRepository.GetRow(ctx, id)
	if err != nil {
		return
	}
	res = make([]dto.BookResponse, len(data.Books))

	for i := 0; i < len(data.Books); i++ {
		res[i], err = s.bookService.Get(ctx, data.Books[i])
		if err != nil {
			return
		}
	}

	return
}
