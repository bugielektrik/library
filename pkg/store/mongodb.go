package store

import (
	"context"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const timeout = 10 * time.Second

type Mongo struct {
	Connection *mongo.Client
}

func NewMongo(uri string) (store *Mongo, err error) {
	store.Connection, err = mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err = store.Connection.Connect(ctx); err != nil {
		return
	}

	if err = store.Connection.Ping(context.Background(), nil); err != nil {
		return
	}

	return
}
