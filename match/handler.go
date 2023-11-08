package match

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/fdp7/beachvolleyapp-api/store"
)

const (
	playerQueryParam    = "player"
	matchDateQueryParam = "date"
)

func AddMatch(ctx *gin.Context) {

	leagueId := ctx.Param("leagueId")
	sportId := ctx.Param("sportId")

	/*sportStr := ctx.Param("sport")
	sport := store.Sport(sportStr)
	_, ok := store.EnabledSport[sport]
	if !ok {
		ctx.JSON(http.StatusNotAcceptable, gin.H{
			"message": "sport is not enabled",
		})
		return
	}*/

	if store.DBSql == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "store is not initialized",
		})

		return
	}

	match := &Match{}
	if err := ctx.BindJSON(match); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid match data",
		})
		return
	}

	storeMatch := matchToStoreMatchP(leagueId, sportId, match)

	err := store.DBSql.AddMatch(ctx, storeMatch)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to add match",
		})
	}
	ctx.JSON(http.StatusCreated, gin.H{})
}

func GetMatches(ctx *gin.Context) {

	leagueId := ctx.Param("leagueId")
	sportId := ctx.Param("sportId")

	/*
		sportStr := ctx.Param("sport")
		sport := store.Sport(sportStr)
		_, ok := store.EnabledSport[sport]
		if !ok {
			ctx.JSON(http.StatusNotAcceptable, gin.H{
				"message": "sport is not enabled",
			})

			return
		}*/

	if store.DBSql == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "store is not initialized",
		})

		return
	}

	name := ctx.Request.URL.Query().Get(playerQueryParam)

	result, err := store.DBSql.GetMatches(ctx, leagueId, sportId, name)
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

	matches := &[]MatchP{}

	if err := json.Unmarshal(result, matches); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to unmarshal matches",
		})

		return
	}
	ctx.JSON(http.StatusOK, gin.H{"matches": matches})
}

func DeleteMatch(ctx *gin.Context) {
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

	matchDate := ctx.Request.URL.Query().Get(matchDateQueryParam)
	FormattedMatchDate, err := time.Parse(time.RFC3339, matchDate)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to format match date",
		})
	}

	err = store.DBSport.DeleteMatch(ctx, FormattedMatchDate, sport)
	if errors.Is(err, store.ErrNoMatchFound) {
		ctx.JSON(http.StatusNoContent, gin.H{
			"message": "no match found",
		})
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to delete match",
		})

		return
	}

	ctx.JSON(http.StatusOK, gin.H{})
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

func matchToStoreMatchP(leagueId string, sportId string, m *Match) *store.MatchP {

	sportIdInt, _ := strconv.Atoi(sportId)
	leagueIdInt, _ := strconv.Atoi(leagueId)

	return &store.MatchP{
		SportId:  sportIdInt,
		LeagueId: leagueIdInt,
		TeamA:    m.TeamA,
		TeamB:    m.TeamB,
		ScoreA:   m.ScoreA,
		ScoreB:   m.ScoreB,
		Date:     m.Date,
	}
}
