package store

import (
	"context"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// mongodb://username:password@localhost:27017/?retryWrites=true&w=majority&tls=false

const timeout = 10 * time.Second

type Mongo struct {
	Client *mongo.Client
}

func NewMongo(uri string) (store Mongo, err error) {
	store.Client, err = mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err = store.Client.Connect(ctx); err != nil {
		return
	}

	if err = store.Client.Ping(context.Background(), nil); err != nil {
		return
	}

	return
}
