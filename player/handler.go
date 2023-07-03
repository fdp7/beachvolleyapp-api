package player

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/fdp7/beachvolleyapp-api/store"
)

func GetPlayers(ctx *gin.Context) {
	sportStr := ctx.Param("sport")

	sport := store.Sport(sportStr)
	_, ok := store.EnabledSport[sport]

	if !ok {
		ctx.JSON(http.StatusNotAcceptable, gin.H{
			"message": "sport is not enabled",
		})
		return
	}

	if store.DBSport == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "store is not initialized",
		})

		return
	}

	result, err := store.DBSport.GetPlayers(ctx, sport)
	if errors.Is(err, store.ErrNoPlayerFound) {
		ctx.JSON(http.StatusNotFound, gin.H{
			"message": "failed to retrieve players",
		})

		return
	}

	players := &[]Player{}

	if err := json.Unmarshal(result, players); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to unmarshal players",
		})

		return
	}
	ctx.JSON(http.StatusOK, gin.H{"players": players})
}

func GetPlayer(ctx *gin.Context) {
	name := ctx.Param("name")
	sportStr := ctx.Param("sport")

	sport := store.Sport(sportStr)
	_, ok := store.EnabledSport[sport]
	if !ok {
		ctx.JSON(http.StatusNotAcceptable, gin.H{
			"message": "sport is not enabled",
		})

		return
	}

	if store.DBSport == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "store is not initialized",
		})
		return
	}

	result, err := store.DBSport.GetPlayer(ctx, name, sport)
	if errors.Is(err, store.ErrNoPlayerFound) {
		ctx.JSON(http.StatusNotFound, gin.H{
			"message": "no player found",
		})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "can't find player",
		})
		return
	}

	player := &Player{}

	if err := json.Unmarshal(result, player); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to unmarshal player",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"player": player})
}

func GetRanking(ctx *gin.Context) {
	sportStr := ctx.Param("sport")

	sport := store.Sport(sportStr)
	_, ok := store.EnabledSport[sport]
	if !ok {
		ctx.JSON(http.StatusNotAcceptable, gin.H{
			"message": "sport is not enabled",
		})
		return
	}

	if store.DBSport == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "store is not initialized",
		})

		return
	}

	result, err := store.DBSport.GetRanking(ctx, sport)
	if errors.Is(err, store.ErrNoPlayerFound) {
		ctx.JSON(http.StatusNotFound, gin.H{
			"message": "ranking is empty",
		})

		return
	}

	players := &[]Player{}

	if err := json.Unmarshal(result, players); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to unmarshal players",
		})

		return
	}
	ctx.JSON(http.StatusOK, gin.H{"ranking": players})
}

func GenerateBalancedTeams(ctx *gin.Context) {
	sportStr := ctx.Param("sport")

	sport := store.Sport(sportStr)
	_, ok := store.EnabledSport[sport]
	if !ok {
		ctx.JSON(http.StatusNotAcceptable, gin.H{
			"message": "sport is not enabled",
		})

		return
	}

	if store.DBSport == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "store is not initialized",
		})

		return
	}

	var body struct {
		Players []Player `json:"players"`
	}

	if err := json.NewDecoder(ctx.Request.Body).Decode(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid player data",
		})

		return
	}

	players := body.Players

	/*players := &[]Player{}
	if err := ctx.BindJSON(players); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid player data",
		})

		return
	}
	if len(*players) <= 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "not enough players",
		})

		return
	}*/

	// turn data into storage type
	storePlayers := make([]store.Player, len(players))
	for i, player := range players {
		storePlayers[i] = playerToStorePlayer(player)
	}

	team1, team2, rtValueDiff, swaps, err := store.DBSport.GenerateBalancedTeams(ctx, storePlayers, sport)
	if err != nil {
		ctx.JSON(http.StatusNoContent, gin.H{
			"message": "failed to generate balanced teams",
		})

		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"balancedTeam1":       team1,
		"balancedTeam2":       team2,
		"teamValueDifference": rtValueDiff,
		"swaps":               swaps},
	)
}

func playerToStorePlayer(p Player) store.Player {
	return store.Player{
		ID:         p.ID,
		Name:       p.Name,
		MatchCount: p.MatchCount,
		WinCount:   p.WinCount,
		Elo:        p.Elo,
		LastElo:    p.LastElo,
	}
}
