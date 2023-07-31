package member

type Entity struct {
	ID       string   `db:"id"`
	FullName *string  `db:"full_name"`
	Books    []string `db:"books"`
}
