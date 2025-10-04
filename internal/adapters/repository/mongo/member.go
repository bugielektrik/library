package mongo

import (
	"context"
	"errors"
	"time"
	"library-service/internal/domain/member"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"library-service/internal/infrastructure/store"
)

// MemberRepository handles CRUD operations for members in MongoDB.
type MemberRepository struct {
	collection *mongo.Collection
}

// NewMemberRepository creates a new instance of MemberRepository.
func NewMemberRepository(db *mongo.Database) *MemberRepository {
	return &MemberRepository{
		collection: db.Collection("members"),
	}
}

// List retrieves all members from the MongoDB collection.
func (r *MemberRepository) List(ctx context.Context) ([]member.Member, error) {
	cur, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var members []member.Member
	if err = cur.All(ctx, &members); err != nil {
		return nil, err
	}

	return members, nil
}

// Add inserts a new member into the MongoDB collection.
func (r *MemberRepository) Add(ctx context.Context, data member.Member) (string, error) {
	res, err := r.collection.InsertOne(ctx, data)
	if err != nil {
		return "", err
	}

	id := res.InsertedID.(primitive.ObjectID).Hex()
	return id, nil
}

// Get retrieves a member by ID from the MongoDB collection.
func (r *MemberRepository) Get(ctx context.Context, id string) (member.Member, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return member.Member{}, err
	}

	var member member.Member
	if err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&member); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return member, store.ErrorNotFound
		}
		return member, err
	}

	return member, nil
}

// Update modifies an existing member in the MongoDB collection.
func (r *MemberRepository) Update(ctx context.Context, id string, data member.Member) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	updateData := r.prepareUpdateData(data)
	if len(updateData) == 0 {
		return nil
	}

	res, err := r.collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": updateData})
	if err != nil {
		return err
	}

	if res.MatchedCount == 0 {
		return store.ErrorNotFound
	}

	return nil
}

// prepareUpdateData prepares the data for the update query.
func (r *MemberRepository) prepareUpdateData(data member.Member) bson.M {
	updateData := bson.M{}

	if data.FullName != nil {
		updateData["full_name"] = data.FullName
	}

	if len(data.Books) > 0 {
		updateData["books"] = data.Books
	}

	return updateData
}

// Delete removes a member by ID from the MongoDB collection.
func (r *MemberRepository) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	res, err := r.collection.DeleteOne(ctx, bson.M{"_id": objID})
	if err != nil {
		return err
	}

	if res.DeletedCount == 0 {
		return store.ErrorNotFound
	}

	return nil
}

// GetByEmail retrieves a member by email (stub implementation)
func (r *MemberRepository) GetByEmail(ctx context.Context, email string) (member.Member, error) {
	return member.Member{}, errors.New("not implemented")
}

// UpdateLastLogin updates last login time (stub implementation)
func (r *MemberRepository) UpdateLastLogin(ctx context.Context, id string, loginTime time.Time) error {
	return errors.New("not implemented")
}

// EmailExists checks if email exists (stub implementation)
func (r *MemberRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	return false, errors.New("not implemented")
}
