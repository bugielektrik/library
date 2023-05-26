package entity

import (
	"library/pkg/database/postgres"
)

type Book struct {
	ID      string         `db:"id"`
	Name    *string        `db:"name"`
	Genre   *string        `db:"genre"`
	ISBN    *string        `db:"isbn"`
	Authors postgres.Array `db:"authors"`
}
