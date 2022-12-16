package player

type Player struct {
	ID         string `json:"_id" bson:"_id"`
	Name       string `json:"name" bson:"name"`
	MatchCount int    `json:"match_count" bson:"match_count"`
	WinCount   int    `json:"win_count" bson:"win_count"`
}
