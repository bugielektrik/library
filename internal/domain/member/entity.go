package member

type Entity struct {
	ID       string   `db:"id" bson:"_id"`
	FullName *string  `db:"full_name" bson:"full_name"`
	Books    []string `db:"books" bson:"books"`
}

func New(req Request) Entity {
	return Entity{
		FullName: &req.FullName,
		Books:    req.Books,
	}
}
