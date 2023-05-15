package service

import (
	"library/internal/dto"
	"library/internal/entity"
	"library/internal/repository"
)

type BookService interface {
	Create(req dto.BookRequest) (res dto.BookResponse, err error)
	GetByID(id string) (res dto.BookResponse, err error)
	GetAll() (res []dto.BookResponse, err error)
	Update(id string, req dto.BookRequest) (err error)
	Delete(id string) (err error)
}

type bookService struct {
	bookRepository repository.BookRepository
}

func NewBookService(b repository.BookRepository) BookService {
	return &bookService{
		bookRepository: b,
	}
}

func (s *bookService) Create(req dto.BookRequest) (res dto.BookResponse, err error) {
	data := entity.Book{
		Name:    &req.Name,
		Genre:   &req.Genre,
		ISBN:    &req.ISBN,
		Authors: req.Authors,
	}

	data.ID, err = s.bookRepository.CreateRow(data)
	if err != nil {
		return
	}
	res = dto.ParseFromBook(data)

	return
}

func (s *bookService) GetByID(id string) (res dto.BookResponse, err error) {
	data, err := s.bookRepository.GetRowByID(id)
	if err != nil {
		return
	}
	res = dto.ParseFromBook(data)

	return
}

func (s *bookService) GetAll() (res []dto.BookResponse, err error) {
	data, err := s.bookRepository.SelectRows()
	if err != nil {
		return
	}
	res = dto.ParseFromBooks(data)

	return
}

func (s *bookService) Update(id string, req dto.BookRequest) (err error) {
	data := entity.Book{
		Name:    &req.Name,
		Genre:   &req.Genre,
		ISBN:    &req.ISBN,
		Authors: req.Authors,
	}
	return s.bookRepository.UpdateRow(id, data)
}

func (s *bookService) Delete(id string) (err error) {
	return s.bookRepository.DeleteRow(id)
}
