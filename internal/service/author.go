package service

import (
	"context"

	"library/internal/dto"
	"library/internal/entity"
	"library/internal/repository"
)

type Author struct {
	authorRepository repository.Author
}

func NewAuthorService(a repository.Author) Author {
	return Author{
		authorRepository: a,
	}
}

func (s *Author) List(ctx context.Context) (res []dto.AuthorResponse, err error) {
	data, err := s.authorRepository.SelectRows(ctx)
	if err != nil {
		return
	}
	res = dto.ParseFromAuthors(data)

	return
}

func (s *Author) Add(ctx context.Context, req dto.AuthorRequest) (res dto.AuthorResponse, err error) {
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

func (s *Author) Get(ctx context.Context, id string) (res dto.AuthorResponse, err error) {
	data, err := s.authorRepository.GetRow(ctx, id)
	if err != nil {
		return
	}
	res = dto.ParseFromAuthor(data)

	return
}

func (s *Author) Update(ctx context.Context, id string, req dto.AuthorRequest) (err error) {
	data := entity.Author{
		FullName:  &req.FullName,
		Pseudonym: &req.Pseudonym,
		Specialty: &req.Specialty,
	}
	return s.authorRepository.UpdateRow(ctx, id, data)
}

func (s *Author) Delete(ctx context.Context, id string) (err error) {
	return s.authorRepository.DeleteRow(ctx, id)
}
