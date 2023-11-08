package store

import (
	"context"
	"errors"
	"fmt"
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
	//GetFriendNFoe(ctx context.Context, leagueId string, sportId string, name string) (*FriendNFoe, error)
	//GenerateBalancedTeams(ctx context.Context, players []Player, sport Sport) ([]string, []string, float64, int, error)

	GetMatches(ctx context.Context, leagueId string, sportId string, name string) ([]byte, error)
	AddMatch(ctx context.Context, m *MatchP) error
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
