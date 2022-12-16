package match

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/fdp7/beachvolleyapp-api/store"
)

const (
	playerQueryParam = "player"
)

func AddMatch(ctx *gin.Context) {
	if store.DB == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "store is not initialized",
		})

		return
	}

	match := &Match{}
	if err := ctx.BindJSON(match); err != nil {
		return
	}

	storeMatch := matchToStoreMatch(match)

	err := store.DB.AddMatch(ctx, storeMatch)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to add match",
		})
	}
	ctx.JSON(http.StatusCreated, gin.H{})
}

func GetMatches(ctx *gin.Context) {
	if store.DB == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "store is not initialized",
		})

		return
	}

	player := ctx.Request.URL.Query().Get(playerQueryParam)

	result, err := store.DB.GetMatches(ctx, player)
	if errors.Is(err, store.ErrNoMatchFound) {
		ctx.JSON(http.StatusNotFound, gin.H{
			"message": "no match found",
		})

		return
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to retrieve matches",
		})

		return
	}

	matches := &[]Match{}

	if err := json.Unmarshal(result, matches); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to unmarshal matches",
		})

		return
	}
	ctx.JSON(http.StatusOK, gin.H{"matches": matches})
}

func matchToStoreMatch(m *Match) *store.Match {
	return &store.Match{
		TeamA:  m.TeamA,
		TeamB:  m.TeamB,
		ScoreA: m.ScoreA,
		ScoreB: m.ScoreB,
		Date:   m.Date,
	}
}
