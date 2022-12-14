package match

import (
	"github.com/gin-gonic/gin"
	"net/http"

	"encoding/json"

	"github.com/fdp7/beachvolleyapp-api/store"
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
	ctx.JSON(http.StatusOK, gin.H{})
}

func GetMatches(ctx *gin.Context) {
	if store.DB == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "store is not initialized",
		})

		return
	}

	matches, err := store.DB.GetMatches(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to retrieve matches",
		})
	}

	var result []Match
	if err := json.Unmarshal(matches, &result); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to unmarshal matches",
		})
	}
	ctx.JSON(http.StatusOK, gin.H{"matches": result})
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
