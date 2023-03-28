package player

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/fdp7/beachvolleyapp-api/store"
)

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
	ctx.JSON(http.StatusOK, gin.H{"ranking": players})
}

/*func userToStorePlayer(p *Player) *store.Player {
	return &store.Player{
		ID:         p.ID,
		Name:       p.Name,
		MatchCount: p.MatchCount,
		WinCount:   p.WinCount,
		Elo:        p.Elo,
		LastElo:    p.LastElo,
	}
}*/
