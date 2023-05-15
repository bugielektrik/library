package entity

// Member represents a domain entity
type Member struct {
	ID       string   `db:"id"`
	FullName *string  `db:"full_name"`
	Books    []string `db:"books"`
}
