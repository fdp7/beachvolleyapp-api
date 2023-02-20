package store

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
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
	err = s.updatePlayer(ctx, m, players, false)
	if err != nil {
		return fmt.Errorf("failed to update playes stats: %w", err)
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
	sort := options.Find().SetSort(orderDate).SetLimit(10)

	results, err := collection.Find(ctx, filter, sort)
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

	err = s.updatePlayer(ctx, match, players, true)
	if err != nil {
		return fmt.Errorf("failed to update player stats: %w", err)
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
		"elo":         100.0, //default Elo for a new player
		"last_elo":    100.0, //default Elo for a new player
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

	// get all players, ordered by max(last_elo), max(win_count) and min(match_count) and alphabetical(name)
	filter := bson.M{}

	order := bson.D{{"last_elo", -1}, {"win_count", -1}, {"match_count", 1}, {"name", 1}}
	sort := options.Find().SetSort(order)

	results, err := collection.Find(ctx, filter, sort)
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

// update player stats (match_count, win_count, elo) based on played or deleted match
func (s *mongoStore) updatePlayer(ctx context.Context, m *Match, players []string, onDeletedMatch bool) error {

	collection := s.client.Database(s.dbName).Collection(s.playerCollection)

	// check which team won
	isTeamAWinner := false
	if m.ScoreA > m.ScoreB {
		isTeamAWinner = true
	}

	// get list of players entities in the match
	var playersList []*Player

	for _, p := range players {

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
		playersList = append(playersList, player)
	}

	// get teamA and teamB total ratings
	var teamARating float64
	var teamBRating float64

	for _, player := range playersList {
		for _, playerName := range m.TeamA {
			if playerName == player.Name {
				teamARating = teamARating + player.Elo[len(player.Elo)-1] // use last elo
			}
		}
		for _, playerName := range m.TeamB {
			if playerName == player.Name {
				teamBRating = teamBRating + player.Elo[len(player.Elo)-1]
			}
		}
	}

	// compute updated stats
	for _, p := range playersList {

		playerInTeamA := false
		isPlayerWinner := false

		// check which team player played for
		for _, i := range m.TeamA {
			if i == p.Name {
				playerInTeamA = true
				break
			}
		}

		//check if player won
		if playerInTeamA && isTeamAWinner {
			isPlayerWinner = true
		} else if !playerInTeamA && !isTeamAWinner {
			isPlayerWinner = true
		}

		// edit player stats
		if onDeletedMatch {
			p.MatchCount = p.MatchCount - 1
			if isPlayerWinner {
				p.WinCount = p.WinCount - 1
			}
			p.LastElo, p.Elo, _ = s.computeElo(p, teamARating, teamBRating, playerInTeamA, isPlayerWinner, true)
		} else {
			p.MatchCount = p.MatchCount + 1
			if isPlayerWinner {
				p.WinCount = p.WinCount + 1
			}
			p.LastElo, p.Elo, _ = s.computeElo(p, teamARating, teamBRating, playerInTeamA, isPlayerWinner, false)
		}
	}

	// update players stats
	for _, p := range playersList {

		filter := bson.M{"name": p.Name}
		update := bson.D{{"$set",
			bson.D{
				{"match_count", p.MatchCount},
				{"win_count", p.WinCount},
				{"elo", p.Elo},
				{"last_elo", p.LastElo},
			},
		}}
		opts := options.Update().SetUpsert(false)

		_, err := collection.UpdateOne(ctx, filter, update, opts)
		if err != nil {
			return fmt.Errorf("failed to update player: %w", err)
		}
	}

	return nil
}

// compute updated elo for player according to the following formula:
// r^ = r + k(s-e)alpha , where
//
//	r^ is the updated elo
//	r is the previous elo
//	k = 32
//	s = {1,0} is the assigned score for win/loss
//	e = 1 / ( 1 + 10 ^(( Rb - Ra) / d) ) is the expected probability that player wins the match, where
//		R is team total elo (sum of elo per team); d = 400
//	alpha = r / R is the player weight/importance for his team
func (s *mongoStore) computeElo(p *Player, teamARating float64, teamBRating float64,
	playerInTeamA bool, isPlayerWinner bool, onDeletedMatch bool) (float64, []float64, error) {

	var playerWeight float64
	var expectedResult float64
	var score float64
	var d float64
	var k float64

	d = 400
	k = 32

	// calculate player weight in team [0,1] and expected match result based on team total ratings
	if playerInTeamA {
		playerWeight = p.Elo[len(p.Elo)-1] / teamARating
		expectedResult = 1 / (1 + math.Pow(10, (teamBRating-teamARating)/d))
	} else {
		playerWeight = p.Elo[len(p.Elo)-1] / teamBRating
		expectedResult = 1 / (1 + math.Pow(10, (teamARating-teamBRating)/d))
	}

	// define score values
	if isPlayerWinner {
		score = 1
	} else {
		score = 0
	}

	// compute updated player rating
	if onDeletedMatch {

		// if deleting match --> rollback the updated rating based on the removed match.
		// actually you would need the previous teamA and teamB exact ratings to be precise, while I can only have the already updated ratings.
		// so there is a small difference after deleting the match with respect to the real previous elo
		p.LastElo = p.Elo[len(p.Elo)-1] - k*(score-expectedResult)*playerWeight
		p.Elo = append(p.Elo, p.LastElo)
	} else {
		p.LastElo = p.Elo[len(p.Elo)-1] + k*(score-expectedResult)*playerWeight
		p.Elo = append(p.Elo, p.LastElo)
	}

	return p.LastElo, p.Elo, nil
}
