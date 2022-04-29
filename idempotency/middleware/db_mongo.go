package middleware

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
)

var _ db = &mongoClient2{}

type mongoClient2 struct {
	coll *mongo.Collection
}

func (m *mongoClient2) get(ctx context.Context, idempotencyKey string) (RequestRecording, bool, error) {
	filter := bson.D{{"idempotency_key", idempotencyKey}}

	var record RequestRecording

	err := m.coll.FindOne(ctx, filter).Decode(&record)
	switch err {
	case nil:
		return record, true, nil
	case mongo.ErrNoDocuments:
		return RequestRecording{}, false, nil
	default:
		return RequestRecording{}, false, err
	}
}

func (m *mongoClient2) update(ctx context.Context, idempotencyKey string, record RequestRecording) error {
	filter := bson.D{{"idempotency_key", idempotencyKey}}
	operation := bson.D{{"$set", record}}

	upsert := true
	opt := &options.UpdateOptions{Upsert: &upsert}
	_, err := m.coll.UpdateOne(ctx, filter, operation, opt)
	if err != nil {
		return err
	}

	return nil
}

func NewMongoDB(uri, dbName, collectionName string) *mongoClient2 {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalln("failed to connect", err)
	}
	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatalln("failed to ping", err)
	}
	collection := client.Database(dbName).Collection(collectionName)

	return &mongoClient2{
		coll: collection,
	}
}
