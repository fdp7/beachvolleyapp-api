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

	query := `SELECT * FROM "User" WHERE "Name" = $1`

	err := s.client.QueryRow(ctx, query, userName).
		Scan(&user.Id, &user.Name, &user.Password, &user.Email)

	if err != nil {
		return nil, ErrNoUserFound
	}

	return json.Marshal(user)
}

func (s *PostgresStore) AddUser(ctx context.Context, u *UserP) error {

	query := `INSERT INTO "User" ("Id", "Name", "Password", "Email") VALUES ($u.Name, $u.Password, $u.Email)
			on conflict ("Name") do nothing
            returning "Id"`

	err := s.client.QueryRow(ctx, query)

	if len(u.Name) < 2 || len(u.Name) >= 11 {
		return ErrNotValidName
	}
	if err != nil {
		return fmt.Errorf("failed to add user to db: %w", err)
	}
	return nil
}

func (s *PostgresStore) GetPlayers(ctx context.Context, leagueId string, sportId string) ([]byte, error) {

	// get all players ordered by alphabetical name for given league and sport
	var players []PlayerP

	query := `SELECT  u."Id", us."UserId", u."Name", us."LeagueId", us."SportId", us."MatchCount", us."WinCount", us."Elo"
			 FROM "UserStats" as us
			 inner join "User" as u on us."UserId" = u."Id"
			 WHERE us."LeagueId" = $1 and us."SportId" = $2
			 order by u."Name" asc`

	rows, err := s.client.Query(ctx, query, leagueId, sportId)

	if err != nil {
		return nil, ErrNoPlayerFound
	}

	defer rows.Close()

	for rows.Next() {
		var player PlayerP
		err := rows.Scan(
			&player.UserStats.Id,
			&player.UserStats.UserId,
			&player.Name,
			&player.UserStats.LeagueId,
			&player.UserStats.SportId,
			&player.UserStats.MatchCount,
			&player.UserStats.WinCount,
			&player.UserStats.Elo,
		)
		if err != nil {
			return nil, ErrNoPlayerFound
		}
		players = append(players, player)
	}

	return json.Marshal(&players)
}

func (s *PostgresStore) GetPlayer(ctx context.Context, leagueId string, sportId string, userId string) ([]byte, error) {

	// get player with given name for given league and sport
	var player PlayerP

	query := `SELECT  u."Id", us."UserId", u."Name", us."LeagueId", us."SportId", us."MatchCount", us."WinCount", us."Elo"
			 FROM "UserStats" as us
			 inner join "User" as u on us."UserId" = u."Id"
			 WHERE us."LeagueId" = $1 and us."SportId" = $2 and us."UserId" = $3`

	err := s.client.QueryRow(ctx, query, leagueId, sportId, userId).
		Scan(
			&player.UserStats.Id,
			&player.UserStats.UserId,
			&player.Name,
			&player.UserStats.LeagueId,
			&player.UserStats.SportId,
			&player.UserStats.MatchCount,
			&player.UserStats.WinCount,
			&player.UserStats.Elo,
		)

	if err != nil {
		return nil, ErrNoPlayerFound
	}

	return json.Marshal(&player)
}

func (s *PostgresStore) GetRanking(ctx context.Context, leagueId string, sportId string) ([]byte, error) {

	// get all players with at least 1 match played, ordered by max(elo), max(win count) and min(match count) and alphabetical(name)
	var players []PlayerP

	query := `SELECT  u."Id", us."UserId", u."Name", us."LeagueId", us."SportId", us."MatchCount", us."WinCount", us."Elo"
			 FROM "UserStats" as us
			 inner join "User" as u on us."UserId" = u."Id"
			 WHERE us."LeagueId" = $1 and us."SportId" = $2 and us."MatchCount" >= 1
			 order by us."Elo"[array_upper(us."Elo", 1)] desc, us."WinCount" desc, us."MatchCount" asc, u."Name" asc`

	rows, err := s.client.Query(ctx, query, leagueId, sportId)

	if err != nil {
		return nil, ErrNoPlayerFound
	}

	defer rows.Close()

	for rows.Next() {
		var player PlayerP
		err := rows.Scan(
			&player.UserStats.Id,
			&player.UserStats.UserId,
			&player.Name,
			&player.UserStats.LeagueId,
			&player.UserStats.SportId,
			&player.UserStats.MatchCount,
			&player.UserStats.WinCount,
			&player.UserStats.Elo,
		)
		if err != nil {
			return nil, ErrNoPlayerFound
		}
		players = append(players, player)
	}

	return json.Marshal(&players)
}

/*func (s *PostgresStore) GetFriendNFoe(ctx context.Context, leagueId string, sportId string, name string) (*FriendNFoe, error){
}*/

func (s *PostgresStore) GetMatches(ctx context.Context, leagueId string, sportId string, userId string) ([]byte, error) {

	// get all, ordered by descending date, limit number of samples, with query filters
	var matches []MatchPV2

	var query string

	if userId != "" {
		query = `SELECT  m."Id", m."LeagueId", m."SportId",
					ARRAY(
						SELECT u."Name"
						FROM "User" AS u
						WHERE u."Id" = ANY(m."TeamA")
					) AS "TeamA_Names",
					ARRAY(
						SELECT u."Name"
						FROM "User" AS u
						WHERE u."Id" = ANY(m."TeamB")
					) AS "TeamB_Names",
					m."ScoreA", m."ScoreB", m."Date"
				 FROM "Match" as m
				 WHERE m."LeagueId" = $1 and m."SportId" = $2 and ($3 = ANY(m."TeamA") OR $3 = ANY(m."TeamB"))
				 order by m."Date" asc`
	} else {
		query = `SELECT  m."Id", m."LeagueId", m."SportId",
					ARRAY(
						SELECT u."Name"
						FROM "User" AS u
						WHERE u."Id" = ANY(m."TeamA")
					) AS "TeamA_Names",
					ARRAY(
						SELECT u."Name"
						FROM "User" AS u
						WHERE u."Id" = ANY(m."TeamB")
					) AS "TeamB_Names",
					m."ScoreA", m."ScoreB", m."Date"
				 FROM "Match" as m
				 WHERE m."LeagueId" = $1 and m."SportId" = $2
				 order by m."Date" asc`
	}

	rows, err := s.client.Query(ctx, query, leagueId, sportId, userId)

	if err != nil {
		return nil, ErrNoMatchFound
	}

	defer rows.Close()

	for rows.Next() {
		var match MatchPV2
		err := rows.Scan(
			&match.Id,
			&match.LeagueId,
			&match.SportId,
			&match.TeamA,
			&match.TeamB,
			&match.ScoreA,
			&match.ScoreB,
			&match.Date,
		)
		if err != nil {
			return nil, ErrNoMatchFound
		}
		matches = append(matches, match)
	}

	if len(matches) == 0 {
		return nil, ErrNoMatchFound
	}

	return json.Marshal(matches)
}
