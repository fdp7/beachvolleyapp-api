package store

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoSportStore struct {
	client           *mongo.Client
	matchCollection  string
	playerCollection string
	sportDBs         map[Sport]string
}

type MongoUserStore struct {
	client         *mongo.Client
	dbName         string
	userCollection string
}

func NewMongoDBStore(ctx context.Context, connectionURI string) (*MongoUserStore, *MongoSportStore, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionURI))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create mongoDB client: %w", err)
	}

	mus := MongoUserStore{
		client:         client,
		dbName:         viper.GetString("DB_USER_NAME"),
		userCollection: viper.GetString("COLLECTION_USER_NAME"),
	}

	sportDBs := map[Sport]string{}

	for sport := range EnabledSport {
		s := strings.ToUpper(string(sport))
		dbEnv := fmt.Sprintf("DB_%s_NAME", s)
		dbName := viper.GetString(dbEnv)

		if dbName == "" {
			return nil, nil, fmt.Errorf("env %s not set", dbEnv)
		}

		sportDBs[sport] = dbName
	}

	mss := MongoSportStore{
		client:           client,
		matchCollection:  viper.GetString("COLLECTION_MATCH_NAME"),
		playerCollection: viper.GetString("COLLECTION_PLAYER_NAME"),
		sportDBs:         sportDBs,
	}

	return &mus, &mss, nil
}

// --------------------- OPERATIONS

func (s *MongoUserStore) GetUser(ctx context.Context, userName string) ([]byte, error) {
	collection := s.client.Database(s.dbName).Collection(s.userCollection)

	// get the player specified by playerName
	filter := bson.M{"name": userName}

	result := collection.FindOne(ctx, filter)
	if result == nil {
		return nil, fmt.Errorf("failed to retrieve user: %w", result)
	}

	user := &User{}
	if err := result.Decode(user); err != nil {
		return nil, ErrNoUserFound
	}

	return json.Marshal(user)
}

func (s *MongoUserStore) AddUser(ctx context.Context, u *User) error {
	collection := s.client.Database(s.dbName).Collection(s.userCollection)

	_, err := collection.InsertOne(ctx, bson.M{
		"_id":      u.Name,
		"name":     u.Name,
		"password": u.Password,
	})
	if mongo.IsDuplicateKeyError(err) {
		return ErrUserDuplicated
	}
	if len(u.Name) < 2 || len(u.Name) >= 11 {
		return ErrNotValidName
	}
	if err != nil {
		return fmt.Errorf("failed to add user to db: %w", err)
	}
	return nil
}

func (s *MongoSportStore) AddMatch(ctx context.Context, m *Match, sport Sport) error {
	dbName := s.sportDBs[sport]
	collection := s.client.Database(dbName).Collection(s.matchCollection)

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
	err = s.updatePlayer(ctx, m, players, sport, false)
	if err != nil {
		return fmt.Errorf("failed to update playes stats: %w", err)
	}

	return nil
}

func (s *MongoSportStore) GetMatches(ctx context.Context, player string, sport Sport) ([]byte, error) {
	dbName := s.sportDBs[sport]
	collection := s.client.Database(dbName).Collection(s.matchCollection)

	// get all, ordered by descending date, limit number of samples, with query filters
	filter := bson.M{}
	if player != "" {
		filterTeamA := bson.M{"team_a": player}
		filterTeamB := bson.M{"team_b": player}
		filter = bson.M{"$or": []bson.M{filterTeamA, filterTeamB}}
	}

	orderDate := bson.D{{"date", -1}}
	sorting := options.Find().SetSort(orderDate).SetLimit(10)

	results, err := collection.Find(ctx, filter, sorting)
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

func (s *MongoSportStore) DeleteMatch(ctx context.Context, matchDate time.Time, sport Sport) error {
	dbName := s.sportDBs[sport]
	collection := s.client.Database(dbName).Collection(s.matchCollection)

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

	err = s.updatePlayer(ctx, match, players, sport, true)
	if err != nil {
		return fmt.Errorf("failed to update player stats: %w", err)
	}

	return nil
}

func (s *MongoSportStore) GenerateBalancedTeams(ctx context.Context, players []Player, sport Sport) ([]string, []string, float64, int, error) {
	dbName := s.sportDBs[sport]
	collection := s.client.Database(dbName).Collection(s.playerCollection)

	// retrieve players stats
	playersStats := make(map[string]float64)

	for _, p := range players {

		filter := bson.M{"name": p.Name}

		result := collection.FindOne(ctx, filter)
		if result == nil {
			return nil, nil, 0, 0, fmt.Errorf("failed to retrieve player: %v", result)
		}

		player := &Player{}
		if err := result.Decode(player); err != nil {
			return nil, nil, 0, 0, ErrNoPlayerFound
		}

		// compute RealTimeValue (rtValue)
		rtValue := computeRealTimePlayerValue(player.LastElo, player.Elo, player.MatchCount, 3)

		// fill the map with names and rtValues
		playersStats[player.Name] = rtValue
	}

	// generate Balanced Teams
	team1, team2, rtValueDiff, swaps := balanceTeams(playersStats, 1, 10)

	if len(team1)+len(team2) < len(playersStats) {
		return nil, nil, 0, 0, fmt.Errorf("balance teams generation failed")
	}

	return team1, team2, rtValueDiff, swaps, nil
}

func (s *MongoSportStore) AddUserToSportDBs(ctx context.Context, user *User) error {

	player := userToStorePlayer(user)

	for sport := range s.sportDBs {

		err := s.AddPlayer(ctx, player, sport)

		if mongo.IsDuplicateKeyError(err) {
			return ErrPlayerDuplicated
		}
		if err != nil {
			return fmt.Errorf("failed to add player to db: %w", err)
		}
	}

	return nil
}

func (s *MongoSportStore) AddExistingUserToNewSportDBs(ctx context.Context, user *User) error {

	player := userToStorePlayer(user)
	var newSports []Sport

	// find for which sports the logged user is not registered then add to them
	for sport := range s.sportDBs {

		_, err := s.GetPlayer(ctx, player.Name, sport)

		if err != nil {
			newSports = append(newSports, sport)
		}
	}

	for _, newSport := range newSports {

		err := s.AddPlayer(ctx, player, newSport)

		if mongo.IsDuplicateKeyError(err) {
			return ErrPlayerDuplicated
		}
		if err != nil {
			return fmt.Errorf("failed to add player to db: %w", err)
		}
	}

	return nil
}

func (s *MongoSportStore) AddPlayer(ctx context.Context, player *Player, sport Sport) error {
	dbName := s.sportDBs[sport]
	collection := s.client.Database(dbName).Collection(s.playerCollection)

	_, err := collection.InsertOne(ctx, bson.M{
		"_id":         player.Name,
		"name":        player.Name,
		"match_count": player.MatchCount,
		"win_count":   player.WinCount,
		"elo":         player.Elo,
		"last_elo":    player.LastElo,
	})
	if mongo.IsDuplicateKeyError(err) {
		return ErrPlayerDuplicated
	}
	if err != nil {
		return fmt.Errorf("failed to add player to db: %w", err)
	}
	return nil
}

func (s *MongoSportStore) GetPlayers(ctx context.Context, sport Sport) ([]byte, error) {
	dbName := s.sportDBs[sport]
	collection := s.client.Database(dbName).Collection(s.playerCollection)

	// get all players ordered by alphabetical(name)
	filter := bson.M{}
	order := bson.D{{"name", 1}}
	sorting := options.Find().SetSort(order)

	results, err := collection.Find(ctx, filter, sorting)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve players: %w", err)
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

func (s *MongoSportStore) GetPlayer(ctx context.Context, playerName string, sport Sport) ([]byte, error) {
	dbName := s.sportDBs[sport]
	collection := s.client.Database(dbName).Collection(s.playerCollection)

	// get the player specified by playerName
	filter := bson.M{"name": playerName}

	result := collection.FindOne(ctx, filter)
	if result == nil {
		return nil, fmt.Errorf("failed to retrieve player: %v", result)
	}

	player := &Player{}
	if err := result.Decode(player); err != nil {
		return nil, ErrNoPlayerFound
	}

	return json.Marshal(player)
}

func (s *MongoSportStore) GetRanking(ctx context.Context, sport Sport) ([]byte, error) {
	dbName := s.sportDBs[sport]
	collection := s.client.Database(dbName).Collection(s.playerCollection)

	// get all players with at least 1 match played, ordered by max(last_elo), max(win_count) and min(match_count) and alphabetical(name)
	filter := bson.D{{"match_count", bson.D{{"$gt", 0}}}}
	order := bson.D{{"last_elo", -1}, {"win_count", -1}, {"match_count", 1}, {"name", 1}}
	sorting := options.Find().SetSort(order)

	results, err := collection.Find(ctx, filter, sorting)
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

// --------------------- FUNCTIONS

func userToStorePlayer(user *User) *Player {
	return &Player{
		ID:         user.Name,
		Name:       user.Name,
		MatchCount: 0,                //default for a new player
		WinCount:   0,                //default for a new player
		Elo:        []float64{100.0}, //default for a new player
		LastElo:    100.0,            //default for a new player
	}
}

// update player stats (match_count, win_count, elo) based on played or deleted match
func (s *MongoSportStore) updatePlayer(ctx context.Context, m *Match, players []string, sport Sport, onDeletedMatch bool) error {

	dbName := s.sportDBs[sport]
	collection := s.client.Database(dbName).Collection(s.playerCollection)

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
				teamARating = teamARating + player.LastElo // use last elo
			}
		}
		for _, playerName := range m.TeamB {
			if playerName == player.Name {
				teamBRating = teamBRating + player.LastElo
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
			if p.MatchCount < 0 {
				p.MatchCount = 0
			}
			if isPlayerWinner {
				p.WinCount = p.WinCount - 1
			}
			if p.WinCount < 0 {
				p.WinCount = 0
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
func (s *MongoSportStore) computeElo(p *Player, teamARating float64, teamBRating float64,
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
		playerWeight = p.LastElo / teamARating
		expectedResult = 1 / (1 + math.Pow(10, (teamBRating-teamARating)/d))
	} else {
		playerWeight = p.LastElo / teamBRating
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
		p.LastElo = p.LastElo - k*(score-expectedResult)*playerWeight
		p.Elo = append(p.Elo, p.LastElo)
	} else {
		p.LastElo = p.LastElo + k*(score-expectedResult)*playerWeight
		p.Elo = append(p.Elo, p.LastElo)
	}

	return p.LastElo, p.Elo, nil
}

// Generate two teams such that : rtValue(team1) - rtValue(team2) =(about) 0
func balanceTeams(players map[string]float64, teamsValueMaxDifference float64, maxSwaps int) ([]string, []string, float64, int) {
	// sort players from higher to lower rtValue
	keys := make([]string, 0, len(players))
	for key := range players {
		keys = append(keys, key)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return players[keys[i]] > players[keys[j]]
	})

	var team1 []string
	var team2 []string
	var team1rtValue float64
	var team2rtValue float64

	// make 2 teams from the sorted list and compute team rtValues
	for _, key := range keys {
		playerRtValue := players[key]
		if team1rtValue <= team2rtValue {
			team1 = append(team1, key)
			team1rtValue += playerRtValue
		} else {
			team2 = append(team2, key)
			team2rtValue += playerRtValue
		}
	}

	// attempt to swap players to minimize the difference between teams' rtValues until threshold teamsValueMaxDifference or maxSwaps is reached
	swaps := 0
	rtValueDiff := math.Abs(team1rtValue - team2rtValue)
	for rtValueDiff >= teamsValueMaxDifference {
		if swaps >= maxSwaps {
			break
		} else {
			maxIdx, minIdx := findPlayersMaxMinValue(team1, team2, players)

			player1 := team1[maxIdx]
			player2 := team2[minIdx]

			team1rtValueNew := team1rtValue - players[player1] + players[player2]
			team2rtValueNew := team2rtValue + players[player1] - players[player2]

			// if no improvement, that's already the best balance between teams
			if rtValueDiff < math.Abs(team1rtValueNew-team2rtValueNew) {
				break
			} else {
				// swap players, update team values and their difference
				team1[maxIdx], team2[minIdx] = player2, player1

				team1rtValue, team2rtValue = team1rtValueNew, team2rtValueNew
				rtValueDiff = math.Abs(team1rtValue - team2rtValue)

				swaps++
			}
		}
	}

	return team1, team2, rtValueDiff, swaps
}

// find the higher value for team1 and lower value for team2
// return the indexes of the players owing such values
func findPlayersMaxMinValue(team1 []string, team2 []string, players map[string]float64) (int, int) {
	maxIdx := 0
	minIdx := 0
	maxValue := players[team1[0]]
	minValue := players[team2[0]]

	for i := 1; i < len(team1); i++ {
		if players[team1[i]] > maxValue {
			maxValue = players[team1[i]]
			maxIdx = i
		}
	}

	for i := 1; i < len(team2); i++ {
		if players[team2[i]] < minValue {
			minValue = players[team2[i]]
			minIdx = i
		}
	}

	return maxIdx, minIdx
}

// compute RealTimeValue for a player considering :
// last elo, historic avg, latest avg, match played, latest period to analyze
func computeRealTimePlayerValue(lastElo float64, elo []float64, matchCount int, latestPeriod int) float64 {
	var rtValue float64
	e := make([]float64, latestPeriod) // sub-elo trend to analyze

	if len(elo) > latestPeriod {
		// starting from second-latest match fill the sub-elo trend
		for i := 1; i <= latestPeriod; i++ {
			e[i-1] = elo[len(elo)-1-i]
		}

		// compute historic and latest elo average
		totalAvg := computeAvg(elo)
		latestAvg := computeAvg(e)

		// RealTimeVPlayerValue computed as
		// last elo
		// adding bonus/malus based on latest period performance
		// weighted on number of games played wrt the considered latest period length
		// such that the higher the matchCount, the higher the confidence in bonus/malus, the higher the bonus/malus incidence in RTV
		rtValue = lastElo + ((latestAvg - totalAvg) / float64(matchCount/(matchCount-latestPeriod)))
	} else {
		rtValue = lastElo
	}

	return rtValue
}

// compute average of values in a slice
func computeAvg(n []float64) float64 {
	var sum float64
	for i := range n {
		sum = sum + n[i]
	}

	average := sum / float64(len(n))

	return average
}
