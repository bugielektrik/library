package entity

import (
	"library/pkg/database/postgres"
)

type Member struct {
	ID       string         `db:"id"`
	FullName *string        `db:"full_name"`
	Books    postgres.Array `db:"books"`
}
