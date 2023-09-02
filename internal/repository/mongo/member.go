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
	db *mongo.Collection
}

func NewMemberRepository(db *mongo.Database) *MemberRepository {
	return &MemberRepository{
		db: db.Collection("members"),
	}
}

func (r *MemberRepository) List(ctx context.Context) (dest []member.Entity, err error) {
	cur, err := r.db.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	if err = cur.All(ctx, &dest); err != nil {
		return nil, err
	}

	return
}

func (r *MemberRepository) Create(ctx context.Context, data member.Entity) (id string, err error) {
	res, err := r.db.InsertOne(ctx, data)
	if err != nil {
		return "", err
	}

	return res.InsertedID.(primitive.ObjectID).String(), nil
}

func (r *MemberRepository) Get(ctx context.Context, id string) (dest member.Entity, err error) {
	if err = r.db.FindOne(ctx, bson.M{"_id": id}).Decode(&dest); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			err = store.ErrorNotFound
		}
	}

	return
}

func (r *MemberRepository) Update(ctx context.Context, id string, data member.Entity) (err error) {
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

func (r *MemberRepository) prepareArgs(data member.Entity) (args bson.M) {
	if data.FullName != nil {
		args["full_name"] = data.FullName
	}

	if len(data.Books) > 0 {
		args["books"] = data.Books
	}

	return
}

func (r *MemberRepository) Delete(ctx context.Context, id string) (err error) {
	out, err := r.db.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}

	if out.DeletedCount == 0 {
		return store.ErrorNotFound
	}

	return
}
