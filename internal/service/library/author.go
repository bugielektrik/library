package library

import (
	"library/internal/dto"
	"library/internal/entity"
)

func (s *Service) CreateAuthor(req dto.AuthorRequest) (res dto.AuthorResponse, err error) {
	author := entity.Author{
		FullName:  &req.FullName,
		Pseudonym: &req.Pseudonym,
		Specialty: &req.Specialty,
	}

	author.ID, err = s.authors.CreateRow(author)
	if err != nil {
		return
	}
	res = dto.ParseFromAuthor(author)

	return
}

func (s *Service) GetAuthor(id string) (res dto.AuthorResponse, err error) {
	author, err := s.authors.GetRowByID(id)
	if err != nil {
		return
	}
	res = dto.ParseFromAuthor(author)

	return
}

func (s *Service) GetAuthors() (res []dto.AuthorResponse, err error) {
	authors, err := s.authors.SelectRows()
	if err != nil {
		return
	}
	res = dto.ParseFromAuthors(authors)

	return
}

func (s *Service) UpdateAuthor(req dto.AuthorRequest) (err error) {
	author := entity.Author{
		FullName:  &req.FullName,
		Pseudonym: &req.Pseudonym,
		Specialty: &req.Specialty,
	}

	return s.authors.UpdateRow(author)
}

func (s *Service) DeleteAuthor(id string) (err error) {
	return s.authors.DeleteRow(id)
}
