package author

type Entity struct {
	ID        string  `db:"id" bson:"_id"`
	FullName  *string `db:"full_name" bson:"full_name"`
	Pseudonym *string `db:"pseudonym" bson:"pseudonym"`
	Specialty *string `db:"specialty" bson:"specialty"`
}

func New(req Request) Entity {
	return Entity{
		FullName:  &req.FullName,
		Pseudonym: &req.Pseudonym,
		Specialty: &req.Specialty,
	}
}
