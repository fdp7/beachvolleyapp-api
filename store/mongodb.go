package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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
		dbName:           viper.GetString("DATABASE_NAME"),
		matchCollection:  viper.GetString("MATCH_COLLECTION_NAME"),
		playerCollection: viper.GetString("PLAYER_COLLECTION_NAME"),
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
	players := append(m.TeamA, m.TeamB...)
	// update player stats based on played match
	onDeletedMatch := false
	for _, player := range players {
		s.updatePlayer(ctx, m, player, onDeletedMatch)
	}

	return nil
}

// update player stats based on played or deleted match
func (s *mongoStore) updatePlayer(ctx context.Context, m *Match, p string, onDeletedMatch bool) error {
	// GetPlayer where _id/name is equal to p.
	// if added match: +1 to match_count; if he's won, +1 win_count
	// if deleted match: -1 to match_count; if he's won, -1 win_count

	collection := s.client.Database(s.dbName).Collection(s.playerCollection)

	playerInTeamA := false
	isTeamAWinner := false
	isPlayerWinner := false

	// check which team player played for
	for _, i := range m.TeamA {
		if i == p {
			playerInTeamA = true
			break
		}
	}
	// check which team won
	if m.ScoreA > m.ScoreB {
		isTeamAWinner = true
	}
	//check if player won
	if playerInTeamA && isTeamAWinner {
		isPlayerWinner = true
	} else if !playerInTeamA && !isTeamAWinner {
		isPlayerWinner = true
	}

	// get player
	filter := bson.M{"name": p}

	result := collection.FindOne(ctx, filter)
	if result == nil {
		return fmt.Errorf("failed to retrieve player: %w", result)
	}

	player := &Player{}
	if err := result.Decode(player); err != nil {
		return ErrNoPlayerFound
	}

	// update player stats
	newMatchCount := player.MatchCount
	newWinCount := player.WinCount
	if onDeletedMatch {
		newMatchCount = player.MatchCount - 1
		if isPlayerWinner {
			newWinCount = player.WinCount - 1
		}
	} else {
		newMatchCount = player.MatchCount + 1
		if isPlayerWinner {
			newWinCount = player.WinCount + 1
		}
	}
	update := bson.D{{"$set",
		bson.D{
			{"match_count", newMatchCount},
			{"win_count", newWinCount},
		},
	}}
	opts := options.Update().SetUpsert(false)
	_, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to update player: %w", err)
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

func (s *mongoStore) DeleteMatch(ctx context.Context, matchDate time.Time) error {
	collection := s.client.Database(s.dbName).Collection(s.matchCollection)

	// get match by date and delete; update stats of players that played the deleted match
	filter := bson.M{"date": matchDate}

	result := collection.FindOne(ctx, filter)

	match := &Match{}

	if err := result.Decode(match); err != nil {
		return ErrNoMatchFound
	}

	players := append(match.TeamA, match.TeamB...)

	deletedCount, err := collection.DeleteOne(ctx, filter)

	if deletedCount.DeletedCount == 0 {
		return ErrNoMatchFound
	}
	if err != nil {
		return fmt.Errorf("failed to delete match %w, err")
	}

	for _, player := range players {
		s.updatePlayer(ctx, match, player, true)
	}

	return nil
}

func (s *mongoStore) AddPlayer(ctx context.Context, p *Player) error {
	collection := s.client.Database(s.dbName).Collection(s.playerCollection)

	_, err := collection.InsertOne(ctx, bson.M{
		"_id":         p.Name,
		"name":        p.Name,
		"password":    p.Password,
		"match_count": p.MatchCount,
		"win_count":   p.WinCount,
	})
	if mongo.IsDuplicateKeyError(err) {
		return ErrPlayerDuplicated
	}
	if len(p.Name) < 2 {
		return ErrNotValidName
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
