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
	return &BookRepository{db: db.Collection("books")}
}

func (r *BookRepository) List(ctx context.Context) ([]book.Entity, error) {
	cur, err := r.db.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	var books []book.Entity
	if err = cur.All(ctx, &books); err != nil {
		return nil, err
	}
	return books, nil
}

func (r *BookRepository) Add(ctx context.Context, data book.Entity) (string, error) {
	res, err := r.db.InsertOne(ctx, data)
	if err != nil {
		return "", err
	}
	return res.InsertedID.(primitive.ObjectID).Hex(), nil
}

func (r *BookRepository) Get(ctx context.Context, id string) (book.Entity, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return book.Entity{}, err
	}
	var book book.Entity
	err = r.db.FindOne(ctx, bson.M{"_id": objID}).Decode(&book)
	if err != nil && errors.Is(err, mongo.ErrNoDocuments) {
		return book, store.ErrorNotFound
	}
	return book, err
}

func (r *BookRepository) Update(ctx context.Context, id string, data book.Entity) error {
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

func (r *BookRepository) Delete(ctx context.Context, id string) error {
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

func (r *BookRepository) prepareArgs(data book.Entity) bson.M {
	args := bson.M{}
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
	return args
}
