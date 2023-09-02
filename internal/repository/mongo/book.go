package mongo

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"library-service/internal/domain/book"
	"library-service/pkg/store"
)

type BookRepository struct {
	db *mongo.Collection
}

func NewBookRepository(db *mongo.Database) *BookRepository {
	return &BookRepository{
		db: db.Collection("books"),
	}
}

func (r *BookRepository) List(ctx context.Context) (dest []book.Entity, err error) {
	cur, err := r.db.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	if err = cur.All(ctx, &dest); err != nil {
		return nil, err
	}

	return
}

func (r *BookRepository) Create(ctx context.Context, data book.Entity) (id string, err error) {
	res, err := r.db.InsertOne(ctx, data)
	if err != nil {
		return "", err
	}

	return res.InsertedID.(primitive.ObjectID).String(), nil
}

func (r *BookRepository) Get(ctx context.Context, id string) (dest book.Entity, err error) {
	if err = r.db.FindOne(ctx, bson.M{"_id": id}).Decode(&dest); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			err = store.ErrorNotFound
		}
	}

	return
}

func (r *BookRepository) Update(ctx context.Context, id string, data book.Entity) (err error) {
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

func (r *BookRepository) prepareArgs(data book.Entity) (args bson.M) {
	if data.Name != nil {
		args["name"] = data.Name
	}

	if data.Genre != nil {
		args["genre"] = data.Genre
	}

	if data.ISBN != nil {
		args["isbn"] = data.ISBN
	}

	if len(data.Authors) > 0 {
		args["authors"] = data.Authors
	}

	return
}

func (r *BookRepository) Delete(ctx context.Context, id string) (err error) {
	out, err := r.db.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}

	if out.DeletedCount == 0 {
		return store.ErrorNotFound
	}

	return
}
