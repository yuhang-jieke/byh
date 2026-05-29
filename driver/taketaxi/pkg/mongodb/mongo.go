package mongodb

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func NewMongoDB(uri, dbName string) (*mongo.Database, func()) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, func() {}
	}
	err = client.Ping(context.Background(), nil)
	if err != nil {
		client.Disconnect(context.Background())
		return nil, func() {}
	}
	db := client.Database(dbName)
	return db, func() {
		client.Disconnect(context.Background())
	}
}
