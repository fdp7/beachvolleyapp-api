package match

import (
	"encoding/json"
	"errors"
	"github.com/fdp7/beachvolleyapp-api/player"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/fdp7/beachvolleyapp-api/store"
)

const (
	playerQueryParam    = "player"
	matchDateQueryParam = "date"
)

func AddMatch(ctx *gin.Context) {
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

	match := &Match{}
	if err := ctx.BindJSON(match); err != nil {
		return
	}

	storeMatch := matchToStoreMatch(match)

	err := store.DBSport.AddMatch(ctx, storeMatch, sport)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to add match",
		})
	}
	ctx.JSON(http.StatusCreated, gin.H{})
}

func GetMatches(ctx *gin.Context) {
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

	player := ctx.Request.URL.Query().Get(playerQueryParam)

	result, err := store.DBSport.GetMatches(ctx, player, sport)
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

	// DEVE STARE IN MATCH O PLAYER?? PERCHÈ IO GLI PASSERÒ DEI PLAYER,NON MATCH.
	// MODIFICARE GenerateBalancedTeams PERCHÈ ACCOLGA DEI PLAYER E NON DELLE STRINGHE? TANT POI DENTRO SE LI ANDAVA A RECUPERARE DA SOLO COMUNQUE, TANTO VALE PASSARLI DIRETTAMENTE IO?
	players := []player.Player{}
	if err := ctx.BindJSON(players); err != nil {
		return
	}

	var playerNames []string
	for player := range players {
		playerNames = append(playerNames)
	}

	team1, team2, swaps, err := store.DBSport.GenerateBalancedTeams(ctx, players, sport)
	if err != nil {
		ctx.JSON(http.StatusNoContent, gin.H{
			"message": "failed to generate balanced teams",
		})

		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"balancedTeam1": team1,
		"balancedTeam2": team2,
		"swaps":         swaps})
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
