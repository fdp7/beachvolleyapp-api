package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Store interface {
	AddMatch(context.Context, *Match) error
	GetMatches(context.Context, string) ([]byte, error)
	DeleteMatch(context.Context, time.Time) error

	AddPlayer(context.Context, *Player) error
	GetPlayer(context.Context, string) ([]byte, error)
	GetRanking(context.Context) ([]byte, error)
}

var DB Store

var (
	ErrNoMatchFound     = errors.New("no match found")
	ErrPlayerDuplicated = errors.New("player already registered")
	ErrNotValidName     = errors.New("player name is not valid")
	ErrNoPlayerFound    = errors.New("no player found")
)

type StoreType int

const (
	MongoDB StoreType = iota
)

func InitializeDB(ctx context.Context, t StoreType) error {
	switch t {
	case MongoDB:
		connectionUri := viper.GetString("CONNECTIONSTRING_MONGODB")
		db, err := NewMongoDBStore(ctx, connectionUri)
		if err != nil {
			return fmt.Errorf("failed to initialize mongoDB: %w", err)
		}
		DB = db

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
	ID         string  `json:"_id" bson:"_id"`
	Name       string  `json:"name" bson:"name"`
	Password   string  `json:"password" bson:"password"`
	MatchCount int     `json:"match_count" bson:"match_count"`
	WinCount   int     `json:"win_count" bson:"win_count"`
	Elo        float64 `json:"elo" bson:"elo"`
}
