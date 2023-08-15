package book

type Entity struct {
	ID      string   `db:"id"`
	Name    *string  `db:"name"`
	Genre   *string  `db:"genre"`
	ISBN    *string  `db:"isbn"`
	Authors []string `db:"authors"`
}
