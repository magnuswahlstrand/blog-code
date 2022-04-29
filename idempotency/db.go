package main

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var _ db = &inMemoryDB{}

//func (s *Service) listDBUsers(ctx context.Context, query string) ([]Order, error) {
//	filter := bson.M{}
//	if query != "" {
//		filter = bson.M{"$text": bson.M{"$search": query}}
//	}
//	users := []Order{}
//	if err := s.mongo.Find(ctx, filter).All(&users); err != nil {
//		return nil, err
//	}
//	return users, nil
//}

func (s *IdempotenceChecker) insertDBUser(ctx context.Context, u Order) (string, error) {
	res, err := s.mongo.InsertOne(ctx, bson.M{
		"product_type": u.ProductType,
	})
	if err != nil {
		return "", err
	}
	return res.InsertedID.(primitive.ObjectID).String(), err
}
