package store

import (
	"context"
	"errors"
	"fmt"
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
		db, err := NewMongoDBStore(ctx, "mongodb+srv://federico:kObR1eLZfNg0WeN9@federicocluster.2bvfp2w.mongodb.net/test")
		if err != nil {
			return fmt.Errorf("failed to initialize mongoDB: %w", err)
		}
		DB = db

	default:
		return errors.New("unknown DB type")
	}

	return nil
}

type Match struct{}
