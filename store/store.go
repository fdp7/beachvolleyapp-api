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
	AddPlayer(ctx context.Context, player *Player, sport Sport) error
	GetPlayers(ctx context.Context, sport Sport) ([]byte, error)
	GetPlayer(ctx context.Context, playerName string, sport Sport) ([]byte, error)
	GetRanking(ctx context.Context, sport Sport) ([]byte, error)
}

type Sport string

const (
	Beachvolley Sport = "beachvolley"
	Basket      Sport = "basket"
)

var EnabledSport = map[Sport]struct{}{
	Beachvolley: {},
	Basket:      {},
}

var DBUser UserStore
var DBSport SportStore

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
)

func InitializeDB(ctx context.Context, t StoreType) error {
	switch t {
	case MongoDB:
		connectionUri := viper.GetString("CONNECTIONSTRING_MONGODB")
		dbUser, dbSport, err := NewMongoDBStore(ctx, connectionUri)
		if err != nil {
			return fmt.Errorf("failed to initialize mongoDB: %w", err)
		}
		DBUser = dbUser
		DBSport = dbSport

	default:
		return errors.New("unknown DB type")
	}

	return nil
}

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
