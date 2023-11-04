package store

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
)

type PostgresStore struct {
	client *pgx.Conn
}

func NewPostgresStore(ctx context.Context, connectionUri string) (*PostgresStore, error) {
	client, err := pgx.Connect(ctx, connectionUri)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres client: %w", err)
	}
	//defer client.Close(ctx)

	ps := PostgresStore{
		client: client,
	}

	return &ps, nil
}

func (s *PostgresStore) GetUser(ctx context.Context, userName string) ([]byte, error) {

	user := &UserP{}

	err := s.client.QueryRow(ctx, `SELECT * FROM "User" WHERE "Name" = $1`, userName).Scan(&user.Id, &user.Name, &user.Password, &user.Email)
	if err != nil {
		return nil, ErrNoUserFound
	}

	return json.Marshal(user)
}

func (s *PostgresStore) AddUser(ctx context.Context, u *UserP) error {

	err := s.client.QueryRow(ctx, `INSERT INTO "User" ("Id", "Name", "Password", "Email") VALUES ($u.Name, $u.Password, $u.Email)`)

	/*if mongo.IsDuplicateKeyError(err) {
		return ErrUserDuplicated
	}*/
	if len(u.Name) < 2 || len(u.Name) >= 11 {
		return ErrNotValidName
	}
	if err != nil {
		return fmt.Errorf("failed to add user to db: %w", err)
	}
	return nil
}

func (s *PostgresStore) GetPlayers(ctx context.Context, sport SportP, league League) ([]byte, error) {

	// get all players ordered by alphabetical name for given league and sport
	players := &UserStats{}

	err := s.client.QueryRow(ctx, `SELECT * FROM "UserStats" as us inner join "User" as u WHERE "SportId" = $1 and "LeagueId" = $2`, sport.Id, league.Id).
		Scan(&players.Id, &players.Name, &players.UserId, &players.SportId, &players.MatchCount, &players.WinCount, &players.Elo)

	if err != nil {
		return nil, ErrNoPlayerFound
	}

	return json.Marshal(players)
}
