package library

import (
	"context"

	"library/internal/domain/author"
)

func (s *Service) ListAuthors(ctx context.Context) (res []author.Response, err error) {
	data, err := s.authorRepository.SelectRows(ctx)
	if err != nil {
		return
	}
	res = author.ParseFromEntities(data)

	return
}

func (s *Service) AddAuthor(ctx context.Context, req author.Request) (res author.Response, err error) {
	data := author.Entity{
		FullName:  &req.FullName,
		Pseudonym: &req.Pseudonym,
		Specialty: &req.Specialty,
	}

	data.ID, err = s.authorRepository.CreateRow(ctx, data)
	if err != nil {
		return
	}
	res = author.ParseFromEntity(data)

	return
}

func (s *Service) GetAuthor(ctx context.Context, id string) (res author.Response, err error) {
	data, err := s.authorRepository.GetRow(ctx, id)
	if err != nil {
		return
	}
	res = author.ParseFromEntity(data)

	return
}

func (s *Service) UpdateAuthor(ctx context.Context, id string, req author.Request) (err error) {
	data := author.Entity{
		FullName:  &req.FullName,
		Pseudonym: &req.Pseudonym,
		Specialty: &req.Specialty,
	}
	return s.authorRepository.UpdateRow(ctx, id, data)
}

func (s *Service) DeleteAuthor(ctx context.Context, id string) (err error) {
	return s.authorRepository.DeleteRow(ctx, id)
}
