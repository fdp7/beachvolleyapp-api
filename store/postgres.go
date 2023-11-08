package store

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v5"
	"math"
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

func (s *PostgresStore) GetPlayer(ctx context.Context, leagueId string, sportId string, name string) ([]byte, error) {

	// get player with given name for given league and sport
	var player PlayerP

	query := `SELECT  u."Id", us."UserId", u."Name", us."LeagueId", us."SportId", us."MatchCount", us."WinCount", us."Elo"
			 FROM "UserStats" as us
			 inner join "User" as u on us."UserId" = u."Id"
			 WHERE us."LeagueId" = $1 and us."SportId" = $2 and u."Name" = $3`

	err := s.client.QueryRow(ctx, query, leagueId, sportId, name).
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

func (s *PostgresStore) GetMatches(ctx context.Context, leagueId string, sportId string, name string) ([]byte, error) {

	// get all, ordered by descending date, limit number of samples, with query filters
	var matches []MatchP

	var query string
	var err error
	var rows pgx.Rows

	if name != "" {
		query = `SELECT  m."Id", m."LeagueId", m."SportId", m."TeamA", m."TeamB", m."ScoreA", m."ScoreB", m."Date"
				 FROM "Match" as m
				 WHERE m."LeagueId" = $1 and m."SportId" = $2 and ($3 = ANY(m."TeamA") OR $3 = ANY(m."TeamB"))
				 order by m."Date" asc`
		rows, err = s.client.Query(ctx, query, leagueId, sportId, name)
	} else {
		query = `SELECT  m."Id", m."LeagueId", m."SportId", m."TeamA", m."TeamB", m."ScoreA", m."ScoreB", m."Date"
				 FROM "Match" as m
				 WHERE m."LeagueId" = $1 and m."SportId" = $2
				 order by m."Date" asc`
		rows, err = s.client.Query(ctx, query, leagueId, sportId)
	}

	if err != nil {
		return nil, ErrNoMatchFound
	}

	defer rows.Close()

	for rows.Next() {
		var match MatchP
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

func (s *PostgresStore) AddMatch(ctx context.Context, m *MatchP) error {
	query := `INSERT INTO "Match" ("LeagueId", "SportId", "TeamA", "TeamB", "ScoreA", "ScoreB", "Date")
				VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := s.client.Exec(ctx, query, m.LeagueId, m.SportId, m.TeamA, m.TeamB, m.ScoreA, m.ScoreB, m.Date)

	if err != nil {
		return fmt.Errorf("failed to add a new match: %w", err)
	}
	players := append(m.TeamA, m.TeamB...)

	// update player stats based on played match
	err = s.updateUserStats(ctx, m, players, false)
	if err != nil {
		return fmt.Errorf("failed to update playes stats: %w", err)
	}

	return nil
}

// update player stats (match_count, win_count, elo) based on played or deleted match
func (s *PostgresStore) updateUserStats(ctx context.Context, m *MatchP, players []string, onDeletedMatch bool) error {

	// check which team won
	isTeamAWinner := false
	if m.ScoreA > m.ScoreB {
		isTeamAWinner = true
	}

	// get list of players entities in the match
	var playersList []*PlayerP

	for _, p := range players {

		// get player by name
		var player *PlayerP

		query := `SELECT  u."Id", us."UserId", u."Name", us."LeagueId", us."SportId", us."MatchCount", us."WinCount", us."Elo"
			 FROM "UserStats" as us
			 inner join "User" as u on us."UserId" = u."Id"
			 WHERE us."LeagueId" = $1 and us."SportId" = $2 and u."Name" = $3`

		err := s.client.QueryRow(ctx, query, m.LeagueId, m.SportId, p).
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
				teamARating = teamARating + player.UserStats.Elo[len(player.UserStats.Elo)-1] // use last elo
			}
		}
		for _, playerName := range m.TeamB {
			if playerName == player.Name {
				teamBRating = teamBRating + player.UserStats.Elo[len(player.UserStats.Elo)-1] // use last elo
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
			p.UserStats.MatchCount = p.UserStats.MatchCount - 1
			if p.UserStats.MatchCount < 0 {
				p.UserStats.MatchCount = 0
			}
			if isPlayerWinner {
				p.UserStats.WinCount = p.UserStats.WinCount - 1
			}
			if p.UserStats.WinCount < 0 {
				p.UserStats.WinCount = 0
			}
			p.UserStats.Elo, _ = s.computeElo(p, teamARating, teamBRating, playerInTeamA, isPlayerWinner, true)
		} else {
			p.UserStats.MatchCount = p.UserStats.MatchCount + 1
			if isPlayerWinner {
				p.UserStats.WinCount = p.UserStats.WinCount + 1
			}
			p.UserStats.Elo, _ = s.computeElo(p, teamARating, teamBRating, playerInTeamA, isPlayerWinner, false)
		}
	}

	// update players stats
	for _, p := range playersList {

		query := `update "UserStats" as us
				set "MatchCount" = $1, "WinCount" = $2, "Elo" = $3
				from "User" as u
				where u."Id" = us."UserId" and us."LeagueId" = $4 and us."SportId" = $5 and u."Name" = $6`

		_, err := s.client.Exec(ctx, query, p.UserStats.MatchCount, p.UserStats.WinCount, p.UserStats.Elo, m.LeagueId, m.SportId, p.Name)

		if err != nil {
			return fmt.Errorf("failed to update %s UserStats: %w", p.Name, err)
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
func (s *PostgresStore) computeElo(p *PlayerP, teamARating float64, teamBRating float64,
	playerInTeamA bool, isPlayerWinner bool, onDeletedMatch bool) ([]float64, error) {

	var playerWeight float64
	var expectedResult float64
	var score float64
	var d float64
	var k float64

	d = 400
	k = 32

	lastElo := p.UserStats.Elo[len(p.UserStats.Elo)-1]

	// calculate player weight in team [0,1] and expected match result based on team total ratings
	if playerInTeamA {
		playerWeight = lastElo / teamARating
		expectedResult = 1 / (1 + math.Pow(10, (teamBRating-teamARating)/d))
	} else {
		playerWeight = lastElo / teamBRating
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
		lastElo = lastElo - k*(score-expectedResult)*playerWeight
		p.UserStats.Elo = append(p.UserStats.Elo, lastElo)
	} else {
		lastElo = lastElo + k*(score-expectedResult)*playerWeight
		p.UserStats.Elo = append(p.UserStats.Elo, lastElo)
	}

	return p.UserStats.Elo, nil
}
