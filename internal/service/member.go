package service

import (
	"library/internal/dto"
	"library/internal/entity"
	"library/internal/repository"
)

type MemberService interface {
	Create(req dto.MemberRequest) (res dto.MemberResponse, err error)
	GetByID(id string) (res dto.MemberResponse, err error)
	GetAll() (res []dto.MemberResponse, err error)
	Update(id string, req dto.MemberRequest) (err error)
	Delete(id string) (err error)
}

type memberService struct {
	memberRepository repository.MemberRepository
}

func NewMemberService(m repository.MemberRepository) MemberService {
	return &memberService{
		memberRepository: m,
	}
}

func (s *memberService) Create(req dto.MemberRequest) (res dto.MemberResponse, err error) {
	data := entity.Member{
		FullName: &req.FullName,
		Books:    req.Books,
	}

	data.ID, err = s.memberRepository.CreateRow(data)
	if err != nil {
		return
	}
	res = dto.ParseFromMember(data)

	return
}

func (s *memberService) GetByID(id string) (res dto.MemberResponse, err error) {
	data, err := s.memberRepository.GetRowByID(id)
	if err != nil {
		return
	}
	res = dto.ParseFromMember(data)

	return
}

func (s *memberService) GetAll() (res []dto.MemberResponse, err error) {
	data, err := s.memberRepository.SelectRows()
	if err != nil {
		return
	}
	res = dto.ParseFromMembers(data)

	return
}

func (s *memberService) Update(id string, req dto.MemberRequest) (err error) {
	data := entity.Member{
		FullName: &req.FullName,
		Books:    req.Books,
	}
	return s.memberRepository.UpdateRow(id, data)
}

func (s *memberService) Delete(id string) (err error) {
	return s.memberRepository.DeleteRow(id)
}
