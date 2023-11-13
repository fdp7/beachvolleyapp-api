package store

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/spf13/viper"
)

type UserStore interface {
	GetUser(ctx context.Context, userName string) ([]byte, error)
	AddUser(ctx context.Context, user *User) error
}

type SportStore interface {
	AddMatch(ctx context.Context, match *Match, sport Sport) error
	GetMatches(ctx context.Context, playerName string, sport Sport) ([]byte, error)
	DeleteMatch(ctx context.Context, date time.Time, sport Sport) error

	AddUserToSportDBs(ctx context.Context, user *User) error
	AddExistingUserToNewSportDBs(ctx context.Context, user *User) error

	AddPlayer(ctx context.Context, player *Player, sport Sport) error
	GetPlayers(ctx context.Context, sport Sport) ([]byte, error)
	GetPlayer(ctx context.Context, playerName string, sport Sport) ([]byte, error)
	GetRanking(ctx context.Context, sport Sport) ([]byte, error)
	GenerateBalancedTeams(ctx context.Context, players []Player, sport Sport) ([]string, []string, float64, int, error)
	GetMates(ctx context.Context, playerName string, sport Sport) (*Mate, *Mate, error)
}

type SqlDbStore interface {
	GetUser(ctx context.Context, userName string) ([]byte, error)
	AddUser(ctx context.Context, user *UserP) error

	GetPlayers(ctx context.Context, leagueId string, sportId string) ([]byte, error)
	GetPlayer(ctx context.Context, leagueId string, sportId string, name string) ([]byte, error)
	GetRanking(ctx context.Context, leagueId string, sportId string) ([]byte, error)
	GetFriendNFoe(ctx context.Context, leagueId string, sportId string, name string) (*FriendNFoe, error)
	GenerateBalancedTeams(ctx context.Context, leagueId string, sportId string, players []*PlayerP) ([]string, []string, float64, int, error)

	GetMatches(ctx context.Context, leagueId string, sportId string, name string) ([]byte, error)
	AddMatch(ctx context.Context, m *MatchP) error
	DeleteMatch(ctx context.Context, leagueId string, sportId string, date time.Time) error

	AddLeague(ctx context.Context, l *LeagueU) error
	DeleteLeague(ctx context.Context, leagueId string) error
	IsFounder(ctx context.Context, userName string, leagueId string) (error, bool)
	IsAdmin(ctx context.Context, userName string, leagueId string) (error, bool)
}

type Sport string

const (
	Beachvolley Sport = "beachvolley"
	Basket      Sport = "basket"
	Pool        Sport = "pool"
)

var EnabledSport = map[Sport]struct{}{
	Beachvolley: {},
	Basket:      {},
	Pool:        {},
}

var DBUser UserStore
var DBSport SportStore
var DBSql SqlDbStore

var (
	ErrNoUserFound      = errors.New("no user found")
	ErrUserDuplicated   = errors.New("user already registered")
	ErrNotValidName     = errors.New("user name is not valid")
	ErrNoPlayerFound    = errors.New("no player found")
	ErrPlayerDuplicated = errors.New("player already registered")
	ErrNoMatchFound     = errors.New("no match found")
	ErrNoLeagueFound    = errors.New("no league found")
)

type StoreType int

const (
	MongoDB StoreType = iota
	Postgres
)

func InitializeDB(ctx context.Context, t StoreType) error {
	switch t {
	/*case MongoDB:
	connectionUri := viper.GetString("CONNECTIONSTRING_MONGODB")
	dbUser, dbSport, err := NewMongoDBStore(ctx, connectionUri)
	if err != nil {
		return fmt.Errorf("failed to initialize mongoDB: %w", err)
	}
	DBUser = dbUser
	DBSport = dbSport*/
	case Postgres:
		connectionUri := viper.GetString("CONNECTIONSTRING_POSTGRES")
		dbUser, err := NewPostgresStore(ctx, connectionUri)
		if err != nil {
			return fmt.Errorf("failed to initialize posgres: %w", err)
		}
		DBSql = dbUser

	default:
		return errors.New("unknown DB type")
	}

	return nil
}

//---------------------------------------------------------------------- MongoDB classes

type Match struct {
	TeamA  []string  `json:"team_a" bson:"team_a"`
	TeamB  []string  `json:"team_b" bson:"team_b"`
	ScoreA int       `json:"score_a" bson:"score_a"`
	ScoreB int       `json:"score_b" bson:"score_b"`
	Date   time.Time `json:"date" bson:"date"`
}

type Player struct {
	ID         string    `json:"_id" bson:"_id"`
	Name       string    `json:"name" bson:"name"`
	MatchCount int       `json:"match_count" bson:"match_count"`
	WinCount   int       `json:"win_count" bson:"win_count"`
	Elo        []float64 `json:"elo" bson:"elo"`
	LastElo    float64   `json:"last_elo" bson:"last_elo"`
}

type User struct {
	ID       string `json:"_id" bson:"_id"`
	Name     string `json:"name" bson:"name"`
	Password string `json:"password" bson:"password"`
}

//---------------------------------------------------------------------- PostgresSQL classes

type UserP struct {
	Id       int    `json:"Id"`
	Name     string `json:"Name"`
	Password string `json:"Password"`
	Email    string `json:"Email"`
}

type UserStats struct {
	Id         int       `json:"Id"`
	UserId     int       `json:"UserId"`
	LeagueId   int       `json:"LeagueId"`
	SportId    int       `json:"SportId"`
	MatchCount int       `json:"MatchCount"`
	WinCount   int       `json:"WinCount"`
	Elo        []float64 `json:"Elo"`
}

type UserMatch struct {
	Id      int `json:"Id"`
	UserId  int `json:"UserId"`
	MatchId int `json:"MatchId"`
}

type UserLeague struct {
	Id       int `json:"Id"`
	UserId   int `json:"UserId"`
	LeagueId int `json:"LeagueId"`
}

type SportP struct {
	Id   int    `json:"Id"`
	Name string `json:"Name"`
}

type MatchP struct {
	Id       int       `json:"Id"`
	SportId  int       `json:"SportId"`
	LeagueId int       `json:"LeagueId"`
	TeamA    []string  `json:"TeamA"`
	TeamB    []string  `json:"TeamB"`
	ScoreA   int       `json:"ScoreA"`
	ScoreB   int       `json:"ScoreB"`
	Date     time.Time `json:"Date"`
}

type League struct {
	Id   int    `json:"Id"`
	Name string `json:"Name"`
}

//---------------------------------------------------------------------- other classes

type PlayerP struct {
	Name      string `json:"Name"`
	UserStats UserStats
}

type Mate struct {
	Name              string `json:"name"`
	WonLossCount      int    `json:"won_loss_count"`
	TotalMatchesCount int    `json:"total_matches_count"`
}

type FriendNFoe struct {
	BestFriend Mate `json:"bestfriend"`
	WorstFoe   Mate `json:"worstfoe"`
}

type LeagueU struct {
	Id   int    `json:"Id"`
	Name string `json:"Name"`
	User string `json:"User"`
}

//---------------------------------------------------------------------- utilities functions

// compute average of values in a slice
func computeAvg(n []float64) float64 {
	var sum float64
	for i := range n {
		sum = sum + n[i]
	}

	average := sum / float64(len(n))

	return average
}

// check if a string is contained in a list
func containsString(list []string, target string) bool {
	for _, str := range list {
		if str == target {
			return true
		}
	}
	return false
}

// find the string that has the most occurrences in a list and count them
func findMaxOccurrences(list []string) (string, int) {
	maxString := ""
	maxCount := 0

	for _, str := range list {
		count := countOccurrences(list, str)
		if count > maxCount {
			maxString = str
			maxCount = count
		}
	}

	return maxString, maxCount
}

// count occurrences of a string in a list
func countOccurrences(list []string, target string) int {
	count := 0
	for _, str := range list {
		if str == target {
			count++
		}
	}
	return count
}

// remove all occurrences of a string from a list
func removeString(list []string, target string) []string {
	result := make([]string, 0)

	for _, str := range list {
		if str != target {
			result = append(result, str)
		}
	}

	return result
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
