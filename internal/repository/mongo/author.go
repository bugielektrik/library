package mongo

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"library-service/internal/domain/author"
	"library-service/pkg/store"
)

type AuthorRepository struct {
	db *mongo.Collection
}

func NewAuthorRepository(db *mongo.Database) *AuthorRepository {
	return &AuthorRepository{
		db: db.Collection("authors"),
	}
}

func (r *AuthorRepository) List(ctx context.Context) (dest []author.Entity, err error) {
	cur, err := r.db.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	if err = cur.All(ctx, &dest); err != nil {
		return nil, err
	}

	return
}

func (r *AuthorRepository) Add(ctx context.Context, data author.Entity) (id string, err error) {
	res, err := r.db.InsertOne(ctx, data)
	if err != nil {
		return "", err
	}

	return res.InsertedID.(primitive.ObjectID).String(), nil
}

func (r *AuthorRepository) Get(ctx context.Context, id string) (dest author.Entity, err error) {
	if err = r.db.FindOne(ctx, bson.M{"_id": id}).Decode(&dest); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			err = store.ErrorNotFound
		}
	}

	return
}

func (r *AuthorRepository) Update(ctx context.Context, id string, data author.Entity) (err error) {
	args := r.prepareArgs(data)
	if len(args) > 0 {

		out, err := r.db.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": args})
		if err != nil {
			return err
		}

		if out.MatchedCount == 0 {
			return store.ErrorNotFound
		}
	}

	return
}

func (r *AuthorRepository) prepareArgs(data author.Entity) (args bson.M) {
	if data.FullName != nil {
		args["full_name"] = data.FullName
	}

	if data.Pseudonym != nil {
		args["pseudonym"] = data.Pseudonym
	}

	if data.Specialty != nil {
		args["specialty"] = data.Specialty
	}

	return
}

func (r *AuthorRepository) Delete(ctx context.Context, id string) (err error) {
	out, err := r.db.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}

	if out.DeletedCount == 0 {
		return store.ErrorNotFound
	}

	return
}
