package repository

import (
	"library/internal/entity"
)

type AuthorRepository interface {
	CreateRow(data entity.Author) (dest string, err error)
	GetRowByID(id string) (dest entity.Author, err error)
	SelectRows() (dest []entity.Author, err error)
	UpdateRow(data entity.Author) (err error)
	DeleteRow(id string) (err error)
}

type BookRepository interface {
	CreateRow(data entity.Book) (dest string, err error)
	GetRowByID(id string) (dest entity.Book, err error)
	SelectRows() (dest []entity.Book, err error)
	UpdateRow(data entity.Book) (err error)
	DeleteRow(id string) (err error)
}

type MemberRepository interface {
	CreateRow(data entity.Member) (dest string, err error)
	GetRowByID(id string) (dest entity.Member, err error)
	SelectRows() (dest []entity.Member, err error)
	UpdateRow(data entity.Member) (err error)
	DeleteRow(id string) (err error)
}
