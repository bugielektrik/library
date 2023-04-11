package library

import (
	"library/internal/dto"
	"library/internal/entity"
)

func (s *Service) CreateMember(req dto.MemberRequest) (res dto.MemberResponse, err error) {
	data := entity.Member{
		FullName: &req.FullName,
		Books:    &req.Books,
	}

	data.ID, err = s.members.CreateRow(data)
	if err != nil {
		return
	}
	res = dto.ParseFromMember(data)

	return
}

func (s *Service) GetMember(id string) (res dto.MemberResponse, err error) {
	data, err := s.members.GetRowByID(id)
	if err != nil {
		return
	}
	res = dto.ParseFromMember(data)

	return
}

func (s *Service) GetMembers() (res []dto.MemberResponse, err error) {
	data, err := s.members.SelectRows()
	if err != nil {
		return
	}
	res = dto.ParseFromMembers(data)

	return
}

func (s *Service) UpdateMember(req dto.MemberRequest) (err error) {
	data := entity.Member{
		FullName: &req.FullName,
		Books:    &req.Books,
	}

	return s.members.UpdateRow(data)
}

func (s *Service) DeleteMember(id string) (err error) {
	return s.members.DeleteRow(id)
}
