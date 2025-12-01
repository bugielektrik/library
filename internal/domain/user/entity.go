package user

import "time"

type Entity struct {
	ID           string     `db:"id" bson:"_id"`
	Email        string     `db:"email" bson:"email"`
	PasswordHash string     `db:"password_hash" bson:"password_hash"`
	FullName     *string    `db:"full_name" bson:"full_name"`
	CreatedAt    *time.Time `db:"created_at" bson:"created_at"`
	UpdatedAt    *time.Time `db:"updated_at" bson:"updated_at"`
}

func New(req SignUpRequest, passwordHash string) Entity {
	now := time.Now()
	return Entity{
		Email:        req.Email,
		PasswordHash: passwordHash,
		FullName:     &req.FullName,
		CreatedAt:    &now,
		UpdatedAt:    &now,
	}
}
