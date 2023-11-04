package player

type Player struct {
	ID         string    `json:"_id"`
	Name       string    `json:"name"`
	MatchCount int       `json:"match_count"`
	WinCount   int       `json:"win_count"`
	Elo        []float64 `json:"elo"`
	LastElo    float64   `json:"last_elo"`
}

type UserStats struct {
	Id         int       `json:"Id"`
	UserId     int       `json:"UserId"`
	SportId    int       `json:"SportId"`
	MatchCount int       `json:"MatchCount"`
	WinCount   int       `json:"WinCount"`
	Elo        []float64 `json:"Elo"`
}
