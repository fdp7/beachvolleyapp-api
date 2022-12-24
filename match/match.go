package match

import "time"

type Match struct {
	TeamA  []string  `json:"team_a"`
	TeamB  []string  `json:"team_b"`
	ScoreA int       `json:"score_a"`
	ScoreB int       `json:"score_b"`
	Date   time.Time `json:"date"`
}
