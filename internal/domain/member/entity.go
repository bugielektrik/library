package member

type Entity struct {
	ID       string   `db:"id" db:"_id"`
	FullName *string  `db:"full_name" db:"full_name"`
	Books    []string `db:"books" db:"books"`
}
