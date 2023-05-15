package service

import (
	"context"
	"library/internal/dto"
	"library/internal/entity"
	"library/internal/repository"
)

type MemberService interface {
	Create(ctx context.Context, req dto.MemberRequest) (res dto.MemberResponse, err error)
	GetByID(ctx context.Context, id string) (res dto.MemberResponse, err error)
	GetAll(ctx context.Context) (res []dto.MemberResponse, err error)
	Update(ctx context.Context, id string, req dto.MemberRequest) (err error)
	Delete(ctx context.Context, id string) (err error)
}

type memberService struct {
	memberRepository repository.MemberRepository
}

func NewMemberService(m repository.MemberRepository) MemberService {
	return &memberService{
		memberRepository: m,
	}
}

func (s *memberService) Create(ctx context.Context, req dto.MemberRequest) (res dto.MemberResponse, err error) {
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

func (s *memberService) GetByID(ctx context.Context, id string) (res dto.MemberResponse, err error) {
	data, err := s.memberRepository.GetRowByID(ctx, id)
	if err != nil {
		return
	}
	res = dto.ParseFromMember(data)

	return
}

func (s *memberService) GetAll(ctx context.Context) (res []dto.MemberResponse, err error) {
	data, err := s.memberRepository.SelectRows(ctx)
	if err != nil {
		return
	}
	res = dto.ParseFromMembers(data)

	return
}

func (s *memberService) Update(ctx context.Context, id string, req dto.MemberRequest) (err error) {
	data := entity.Member{
		FullName: &req.FullName,
		Books:    req.Books,
	}
	return s.memberRepository.UpdateRow(ctx, id, data)
}

func (s *memberService) Delete(ctx context.Context, id string) (err error) {
	return s.memberRepository.DeleteRow(ctx, id)
}
