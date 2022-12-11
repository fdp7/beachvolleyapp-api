package mongodb

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/fdp7/beachvolleyapp-api/store"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type mongoStore struct {
	client *mongo.Client
}

func New(ctx context.Context, connectionURI string) (store.Store, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionURI))
	if err != nil {
		return nil, fmt.Errorf("failed to create mongoDB client: %w", err)
	}

	return &mongoStore{client: client}, nil
}

func (s *mongoStore) AddMatch(ctx context.Context, m *store.Match) error {
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
