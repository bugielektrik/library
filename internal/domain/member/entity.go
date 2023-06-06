package member

import (
	"library/pkg/store/postgres"
)

type Entity struct {
	ID       string         `db:"id"`
	FullName *string        `db:"full_name"`
	Books    postgres.Array `db:"books"`
}
