package mongo

import (
	"context"
	"errors"
	"library-service/internal/domain/member"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"library-service/pkg/store"
)

type MemberRepository struct {
	collection *mongo.Collection
}

func NewMemberRepository(db *mongo.Database) *MemberRepository {
	return &MemberRepository{
		collection: db.Collection("members"),
	}
}

func (r *MemberRepository) List(ctx context.Context) ([]member.Entity, error) {
	cur, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	var members []member.Entity
	if err = cur.All(ctx, &members); err != nil {
		return nil, err
	}

	return members, nil
}

func (r *MemberRepository) Add(ctx context.Context, data member.Entity) (string, error) {
	res, err := r.collection.InsertOne(ctx, data)
	if err != nil {
		return "", err
	}

	id := res.InsertedID.(primitive.ObjectID).Hex()
	return id, nil
}

func (r *MemberRepository) Get(ctx context.Context, id string) (member.Entity, error) {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return member.Entity{}, err
	}

	var member member.Entity
	if err = r.collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&member); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return member, store.ErrorNotFound
		}
		return member, err
	}

	return member, nil
}

func (r *MemberRepository) Update(ctx context.Context, id string, data member.Entity) error {
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

func (r *MemberRepository) prepareUpdateData(data member.Entity) bson.M {
	updateData := bson.M{}

	if data.FullName != nil {
		updateData["full_name"] = data.FullName
	}

	if len(data.Books) > 0 {
		updateData["books"] = data.Books
	}

	return updateData
}

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
