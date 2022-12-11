package store

import (
	"context"
)

type Store interface {
	AddMatch(context.Context, *Match) error
}


type Match struct {}
