package service

import (
	"library/internal/dto"
	"library/internal/entity"
	"library/internal/repository"
)

// AuthorService represents the author service
type AuthorService interface {
	Create(req dto.AuthorRequest) (res dto.AuthorResponse, err error)
	GetByID(id string) (res dto.AuthorResponse, err error)
	GetAll() (res []dto.AuthorResponse, err error)
	Update(id string, req dto.AuthorRequest) (err error)
	Delete(id string) (err error)
}

// authorService is an implementation of the AuthorService interface
type authorService struct {
	authorRepository repository.AuthorRepository
}

// NewAuthorService creates a new instance of the authorService struct
func NewAuthorService(a repository.AuthorRepository) AuthorService {
	return &authorService{
		authorRepository: a,
	}
}

// Create creates a new author
func (s *authorService) Create(req dto.AuthorRequest) (res dto.AuthorResponse, err error) {
	data := entity.Author{
		FullName:  &req.FullName,
		Pseudonym: &req.Pseudonym,
		Specialty: &req.Specialty,
	}

	data.ID, err = s.authorRepository.CreateRow(data)
	if err != nil {
		return
	}
	res = dto.ParseFromAuthor(data)

	return
}

// GetByID retrieves an author by ID
func (s *authorService) GetByID(id string) (res dto.AuthorResponse, err error) {
	data, err := s.authorRepository.GetRowByID(id)
	if err != nil {
		return
	}
	res = dto.ParseFromAuthor(data)

	return
}

// GetAll retrieves all authors
func (s *authorService) GetAll() (res []dto.AuthorResponse, err error) {
	data, err := s.authorRepository.SelectRows()
	if err != nil {
		return
	}
	res = dto.ParseFromAuthors(data)

	return
}

// Update updates an existing author
func (s *authorService) Update(id string, req dto.AuthorRequest) (err error) {
	data := entity.Author{
		FullName:  &req.FullName,
		Pseudonym: &req.Pseudonym,
		Specialty: &req.Specialty,
	}
	return s.authorRepository.UpdateRow(id, data)
}

// Delete deletes a author
func (s *authorService) Delete(id string) (err error) {
	return s.authorRepository.DeleteRow(id)
}
