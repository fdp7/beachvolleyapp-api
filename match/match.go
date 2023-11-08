package match

import "time"

type Match struct {
	TeamA  []string  `json:"team_a"`
	TeamB  []string  `json:"team_b"`
	ScoreA int       `json:"score_a"`
	ScoreB int       `json:"score_b"`
	Date   time.Time `json:"date"`
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
