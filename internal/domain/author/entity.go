package author

type Entity struct {
	ID        string  `db:"id"`
	FullName  *string `db:"full_name"`
	Pseudonym *string `db:"pseudonym"`
	Specialty *string `db:"specialty"`
}
