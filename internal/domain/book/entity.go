package book

type Entity struct {
	ID       string  `db:"id" bson:"_id"`
	Name     *string `db:"name" bson:"name"`
	Genre    *string `db:"genre" bson:"genre"`
	ISBN     *string `db:"isbn" bson:"isbn"`
	AuthorId *string `db:"authorId" bson:"author-id"`
}

func New(req Request) Entity {
	return Entity{
		Name:     &req.Name,
		Genre:    &req.Genre,
		ISBN:     &req.ISBN,
		AuthorId: &req.AuthorID,
	}
}
