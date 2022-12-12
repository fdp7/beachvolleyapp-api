package store

import (
	"context"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"time"
)

type Store interface {
	AddMatch(context.Context, *Match) error
}

var DB Store

type StoreType int

const (
	MongoDB StoreType = iota
)

func InitializeDB(ctx context.Context, t StoreType) error {
	switch t {
	case MongoDB:
		connectionUri := viper.Get("CONNECTIONSTRING_MONGODB").(string)
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
	TeamA  []string
	TeamB  []string
	ScoreA int
	ScoreB int
	Date   time.Time
}
