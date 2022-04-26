package main

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (s *Service) listDBUsers(ctx context.Context, query string) ([]User, error) {
	filter := bson.M{}
	if query != "" {
		filter = bson.M{"$text": bson.M{"$search": query}}
	}
	users := []User{}
	if err := s.mongo.Find(ctx, filter).All(&users); err != nil {
		return nil, err
	}
	return users, nil
}

func (s *Service) insertDBUser(ctx context.Context, u User) (string, error) {
	res, err := s.mongo.InsertOne(ctx, bson.M{
		"name":  u.Name,
		"email": u.Email,
		"phone": u.Phone,
	})
	if err != nil {
		return "", err
	}
	return res.InsertedID.(primitive.ObjectID).String(), err
}
