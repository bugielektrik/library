package entity

// Book represents a book entity
type Book struct {
	ID      string   `db:"id"`
	Name    *string  `db:"name"`
	Genre   *string  `db:"genre"`
	ISBN    *string  `db:"isbn"`
	Authors []string `db:"authors"`
}
