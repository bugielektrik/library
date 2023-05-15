package service

import (
	"library/internal/dto"
	"library/internal/entity"
	"library/internal/repository"
)

// BookService represents the book bookService
type BookService interface {
	Create(req dto.BookRequest) (res dto.BookResponse, err error)
	GetByID(id string) (res dto.BookResponse, err error)
	GetAll() (res []dto.BookResponse, err error)
	Update(id string, req dto.BookRequest) (err error)
	Delete(id string) (err error)
}

// bookService is an implementation of the BookService interface
type bookService struct {
	bookRepository repository.BookRepository
}

// NewBookService creates a new instance of the bookService struct
func NewBookService(b repository.BookRepository) BookService {
	return &bookService{
		bookRepository: b,
	}
}

// Create creates a new book
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

// GetByID retrieves an book by ID
func (s *bookService) GetByID(id string) (res dto.BookResponse, err error) {
	data, err := s.bookRepository.GetRowByID(id)
	if err != nil {
		return
	}
	res = dto.ParseFromBook(data)

	return
}

// GetAll retrieves all books
func (s *bookService) GetAll() (res []dto.BookResponse, err error) {
	data, err := s.bookRepository.SelectRows()
	if err != nil {
		return
	}
	res = dto.ParseFromBooks(data)

	return
}

// Update updates an existing book
func (s *bookService) Update(id string, req dto.BookRequest) (err error) {
	data := entity.Book{
		Name:    &req.Name,
		Genre:   &req.Genre,
		ISBN:    &req.ISBN,
		Authors: req.Authors,
	}
	return s.bookRepository.UpdateRow(id, data)
}

// Delete deletes a book
func (s *bookService) Delete(id string) (err error) {
	return s.bookRepository.DeleteRow(id)
}
