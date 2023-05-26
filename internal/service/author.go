package service

import (
	"context"

	"library/internal/dto"
	"library/internal/entity"
	"library/internal/repository"
)

type AuthorService interface {
	List(ctx context.Context) (res []dto.AuthorResponse, err error)
	Add(ctx context.Context, req dto.AuthorRequest) (res dto.AuthorResponse, err error)
	Get(ctx context.Context, id string) (res dto.AuthorResponse, err error)
	Update(ctx context.Context, id string, req dto.AuthorRequest) (err error)
	Delete(ctx context.Context, id string) (err error)
}

type authorService struct {
	authorRepository repository.AuthorRepository
}

func NewAuthorService(a repository.AuthorRepository) AuthorService {
	return &authorService{
		authorRepository: a,
	}
}

func (s *authorService) List(ctx context.Context) (res []dto.AuthorResponse, err error) {
	data, err := s.authorRepository.SelectRows(ctx)
	if err != nil {
		return
	}
	res = dto.ParseFromAuthors(data)

	return
}

func (s *authorService) Add(ctx context.Context, req dto.AuthorRequest) (res dto.AuthorResponse, err error) {
	data := entity.Author{
		FullName:  &req.FullName,
		Pseudonym: &req.Pseudonym,
		Specialty: &req.Specialty,
	}

	data.ID, err = s.authorRepository.CreateRow(ctx, data)
	if err != nil {
		return
	}
	res = dto.ParseFromAuthor(data)

	return
}

func (s *authorService) Get(ctx context.Context, id string) (res dto.AuthorResponse, err error) {
	data, err := s.authorRepository.GetRow(ctx, id)
	if err != nil {
		return
	}
	res = dto.ParseFromAuthor(data)

	return
}

func (s *authorService) Update(ctx context.Context, id string, req dto.AuthorRequest) (err error) {
	data := entity.Author{
		FullName:  &req.FullName,
		Pseudonym: &req.Pseudonym,
		Specialty: &req.Specialty,
	}
	return s.authorRepository.UpdateRow(ctx, id, data)
}

func (s *authorService) Delete(ctx context.Context, id string) (err error) {
	return s.authorRepository.DeleteRow(ctx, id)
}
