package store

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoStore struct {
	client           *mongo.Client
	dbName           string
	matchCollection  string
	playerCollection string
}

func NewMongoDBStore(ctx context.Context, connectionURI string) (Store, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionURI))
	if err != nil {
		return nil, fmt.Errorf("failed to create mongoDB client: %w", err)
	}
	ms := mongoStore{
		client:           client,
		dbName:           viper.Get("DATABASE_NAME").(string),
		matchCollection:  viper.Get("MATCH_COLLECTION_NAME").(string),
		playerCollection: viper.Get("PLAYER_COLLECTION_NAME").(string),
	}
	return &ms, nil
}

func (s *mongoStore) AddMatch(ctx context.Context, m *Match) error {
	collection := s.client.Database(s.dbName).Collection(s.matchCollection)

	_, err := collection.InsertOne(ctx, bson.M{
		"team_a":  m.TeamA,
		"team_b":  m.TeamB,
		"score_a": m.ScoreA,
		"score_b": m.ScoreB,
		"date":    m.Date,
	})
	if err != nil {
		return fmt.Errorf("failed to add a new match: %w", err)
	}

	return nil
}

func (s *mongoStore) GetMatches(ctx context.Context, player string) ([]byte, error) {
	collection := s.client.Database(s.dbName).Collection(s.matchCollection)

	// get all, ordered by descending date, limit number of samples, with query filters
	filter := bson.M{}
	if player != "" {
		filterTeamA := bson.M{"team_a": player}
		filterTeamB := bson.M{"team_b": player}
		filter = bson.M{"$or": []bson.M{filterTeamA, filterTeamB}}

	}

	orderDate := bson.D{{"date", -1}}
	options := options.Find().SetSort(orderDate).SetLimit(10)

	results, err := collection.Find(ctx, filter, options)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve matches: %w", err)
	}

	var matches []Match

	for results.Next(ctx) {
		match := Match{}
		if err := results.Decode(&match); err != nil {
			return nil, fmt.Errorf("failed to retrieve matches: %w", err)
		}
		matches = append(matches, match)
	}

	if len(matches) == 0 {
		return nil, ErrNoMatchFound
	}

	return json.Marshal(matches)
}

func (s *mongoStore) AddPlayer(ctx context.Context, p *Player) error {
	collection := s.client.Database(s.dbName).Collection(s.playerCollection)

	_, err := collection.InsertOne(ctx, bson.M{
		"_id":         p.Name,
		"name":        p.Name,
		"match_count": p.MatchCount,
		"win_count":   p.WinCount,
	})
	if mongo.IsDuplicateKeyError(err) {
		return ErrPlayerDuplicated
	}
	if err != nil {
		return fmt.Errorf("failed to add player to db: %w", err)
	}
	return nil

}

func (s *mongoStore) GetPlayer(ctx context.Context, playerName string) ([]byte, error) {
	collection := s.client.Database(s.dbName).Collection(s.playerCollection)

	// get the player specified by playerName
	filter := bson.M{"name": playerName}

	result := collection.FindOne(ctx, filter)
	if result == nil {
		return nil, fmt.Errorf("failed to retrieve player: %w", result)
	}

	player := &Player{}
	if err := result.Decode(player); err != nil {
		return nil, ErrNoPlayerFound
	}

	return json.Marshal(player)
}

func (s *mongoStore) GetRanking(ctx context.Context) ([]byte, error) {
	collection := s.client.Database(s.dbName).Collection(s.playerCollection)

	// get all players, ordered by max(win_count) and min(match_count) and alphabetical(name)
	filter := bson.M{}

	order := bson.D{{"win_count", -1}, {"match_count", 1}, {"name", 1}}
	options := options.Find().SetSort(order)

	results, err := collection.Find(ctx, filter, options)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve ranking of players: %w", err)
	}

	var players []Player

	for results.Next(ctx) {
		player := Player{}
		if err := results.Decode(&player); err != nil {
			return nil, fmt.Errorf("failed to retrieve player: %w", err)
		}
		players = append(players, player)
	}

	if len(players) == 0 {
		return nil, ErrNoPlayerFound
	}

	return json.Marshal(players)
}
