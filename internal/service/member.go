package service

import (
	"library/internal/dto"
	"library/internal/entity"
	"library/internal/repository"
)

// MemberService represents the member memberService
type MemberService interface {
	Create(req dto.MemberRequest) (res dto.MemberResponse, err error)
	GetByID(id string) (res dto.MemberResponse, err error)
	GetAll() (res []dto.MemberResponse, err error)
	Update(id string, req dto.MemberRequest) (err error)
	Delete(id string) (err error)
}

// memberService is an implementation of the MemberService interface
type memberService struct {
	memberRepository repository.MemberRepository
}

// NewMemberService creates a new instance of the memberService struct
func NewMemberService(m repository.MemberRepository) MemberService {
	return &memberService{
		memberRepository: m,
	}
}

// Create creates a new member
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

// GetByID retrieves an member by ID
func (s *memberService) GetByID(id string) (res dto.MemberResponse, err error) {
	data, err := s.memberRepository.GetRowByID(id)
	if err != nil {
		return
	}
	res = dto.ParseFromMember(data)

	return
}

// GetAll retrieves all members
func (s *memberService) GetAll() (res []dto.MemberResponse, err error) {
	data, err := s.memberRepository.SelectRows()
	if err != nil {
		return
	}
	res = dto.ParseFromMembers(data)

	return
}

// Update updates an existing member
func (s *memberService) Update(id string, req dto.MemberRequest) (err error) {
	data := entity.Member{
		FullName: &req.FullName,
		Books:    req.Books,
	}
	return s.memberRepository.UpdateRow(id, data)
}

// Delete deletes a member
func (s *memberService) Delete(id string) (err error) {
	return s.memberRepository.DeleteRow(id)
}
