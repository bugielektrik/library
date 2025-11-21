package book

type Entity struct {
	ID      string   `db:"id" bson:"_id"`
	Name    *string  `db:"name" bson:"name"`
	Genre   *string  `db:"genre" bson:"genre"`
	ISBN    *string  `db:"isbn" bson:"isbn"`
	Authors []string `db:"authors" bson:"authors"`
}

func New(req Request) Entity {
	return Entity{
		Name:    &req.Name,
		Genre:   &req.Genre,
		ISBN:    &req.ISBN,
		Authors: req.Authors,
	}
}
