package match

import "time"

type Match struct {
	TeamA []string
	TeamB []string
	ScoreA int
	ScoreB int
	Date time.Time
}
