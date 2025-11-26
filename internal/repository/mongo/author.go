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
	return &AuthorRepository{db: db.Collection("authors")}
}

func (r *AuthorRepository) List(ctx context.Context) ([]author.Entity, error) {
	cur, err := r.db.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	var authors []author.Entity
	if err = cur.All(ctx, &authors); err != nil {
		return nil, err
	}
	return authors, nil
}

func (r *AuthorRepository) Add(ctx context.Context, data author.Entity) (string, error) {
	res, err := r.db.InsertOne(ctx, data)
	if err != nil {
		return "", err
	}
	return res.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (r *AuthorRepository) Get(ctx context.Context, id string) (author.Entity, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return author.Entity{}, err
	}
	var author author.Entity
	err = r.db.FindOne(ctx, bson.M{"_id": objID}).Decode(&author)
	if err != nil && errors.Is(err, mongo.ErrNoDocuments) {
		return author, store.ErrorNotFound
	}
	return author, err
}

func (r *AuthorRepository) Update(ctx context.Context, id string, data author.Entity) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	args := r.prepareArgs(data)
	if len(args) == 0 {
		return nil
	}
	res, err := r.db.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": args})
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return store.ErrorNotFound
	}
	return nil
}

func (r *AuthorRepository) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	res, err := r.db.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return store.ErrorNotFound
	}
	return nil
}

func (r *AuthorRepository) prepareArgs(data author.Entity) bson.M {
	args := bson.M{}
	if data.FullName != nil {
		args["full_name"] = data.FullName
	}
	if data.Pseudonym != nil {
		args["pseudonym"] = data.Pseudonym
	}
	if data.Specialty != nil {
		args["specialty"] = data.Specialty
	}
	return args
}
