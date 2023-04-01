package player

type Player struct {
	ID         string    `json:"_id"`
	Name       string    `json:"name"`
	MatchCount int       `json:"match_count"`
	WinCount   int       `json:"win_count"`
	Elo        []float64 `json:"elo"`
	LastElo    float64   `json:"last_elo"`
}
