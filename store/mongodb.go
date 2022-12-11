package store

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type mongoStore struct {
	client *mongo.Client
}

func NewMongoDBStore(ctx context.Context, connectionURI string) (Store, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionURI))
	if err != nil {
		return nil, fmt.Errorf("failed to create mongoDB client: %w", err)
	}

	return &mongoStore{client: client}, nil
}

func (s *mongoStore) AddMatch(ctx context.Context, m *Match) error {
	collection := s.client.Database("beachvolley").Collection("match")

	_, err := collection.InsertOne(ctx, bson.M{
		"hello": "world",
		"pippo": "world",
	})
	if err != nil {
		return fmt.Errorf("failed to add a new match: %w", err)
	}

	return nil
}
