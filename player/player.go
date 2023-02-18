package player

type Player struct {
	ID         string  `json:"_id"`
	Name       string  `json:"name"`
	Password   string  `json:"password"`
	MatchCount int     `json:"match_count"`
	WinCount   int     `json:"win_count"`
	Elo        float64 `json:"elo"`
}
