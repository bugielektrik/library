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

// AuthorRepository handles CRUD operations for authors in a MongoDB database.
type AuthorRepository struct {
	db *mongo.Collection
}

// NewAuthorRepository creates a new AuthorRepository.
func NewAuthorRepository(db *mongo.Database) *AuthorRepository {
	return &AuthorRepository{db: db.Collection("authors")}
}

// List retrieves all authors from the database.
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

// Add inserts a new author into the database.
func (r *AuthorRepository) Add(ctx context.Context, data author.Entity) (string, error) {
	res, err := r.db.InsertOne(ctx, data)
	if err != nil {
		return "", err
	}
	return res.InsertedID.(primitive.ObjectID).Hex(), nil
}

// Get retrieves an author by ID from the database.
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

// Update modifies an existing author in the database.
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

// Delete removes an author by ID from the database.
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

// prepareArgs prepares the update arguments for the MongoDB query.
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
