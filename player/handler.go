package player

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/fdp7/beachvolleyapp-api/store"
)

func GetPlayers(ctx *gin.Context) {
	//sportStr := ctx.Param("sport")

	leagueId := ctx.Param("leagueId")
	sportId := ctx.Param("sportId")

	/*sport := store.Sport(sportStr)
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

	result, err := store.DBSql.GetPlayers(ctx, leagueId, sportId)
	if errors.Is(err, store.ErrNoPlayerFound) {
		ctx.JSON(http.StatusNotFound, gin.H{
			"message": "failed to retrieve players",
		})

		return
	}

	players := &[]PlayerP{}

	if err := json.Unmarshal(result, players); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to unmarshal players",
		})

		return
	}
	ctx.JSON(http.StatusOK, gin.H{"players": players})
}

func GetPlayer(ctx *gin.Context) {

	leagueId := ctx.Param("leagueId")
	sportId := ctx.Param("sportId")
	name := ctx.Param("name")

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

	result, err := store.DBSql.GetPlayer(ctx, leagueId, sportId, name)
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

	player := &PlayerP{}

	if err := json.Unmarshal(result, player); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to unmarshal player",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"player": player})
}

func GetRanking(ctx *gin.Context) {

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

	result, err := store.DBSql.GetRanking(ctx, leagueId, sportId)
	if errors.Is(err, store.ErrNoPlayerFound) {
		ctx.JSON(http.StatusNotFound, gin.H{
			"message": "ranking is empty",
		})

		return
	}

	players := &[]PlayerP{}

	if err := json.Unmarshal(result, players); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to unmarshal players",
		})

		return
	}
	ctx.JSON(http.StatusOK, gin.H{"ranking": players})
}

func GetFriendNFoe(ctx *gin.Context) {

	leagueId := ctx.Param("leagueId")
	sportId := ctx.Param("sportId")
	name := ctx.Param("name")

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

	fnf, err := store.DBSql.GetFriendNFoe(ctx, leagueId, sportId, name)
	if errors.Is(err, store.ErrNoPlayerFound) {
		ctx.JSON(http.StatusNotFound, gin.H{
			"message": "no player found",
		})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "can't get player's mates stats",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"bestfriend": fnf.BestFriend,
		"worstfoe":   fnf.WorstFoe,
	})
}

func GenerateBalancedTeams(ctx *gin.Context) {

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

	var body struct {
		Players []string `json:"players"`
	}

	if err := json.NewDecoder(ctx.Request.Body).Decode(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid player data",
		})

		return
	}

	playersNames := body.Players
	var players []PlayerP

	for _, playerName := range playersNames {

		result, err := store.DBSql.GetPlayer(ctx, leagueId, sportId, playerName)

		if errors.Is(err, store.ErrNoPlayerFound) {
			ctx.JSON(http.StatusNotFound, gin.H{
				"message": "no player found",
			})
			return
		}
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": "failed to get player",
			})
			return
		}

		p := &PlayerP{}

		if err := json.Unmarshal(result, p); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": "failed to unmarshal player",
			})
			return
		}
		players = append(players, *p)
	}

	// turn data into storage type
	storePlayers := make([]*store.PlayerP, len(players))

	for i, player := range players {
		storePlayers[i] = playerToStorePlayerP(leagueId, sportId, player)
	}

	team1, team2, rtValueDiff, swaps, err := store.DBSql.GenerateBalancedTeams(ctx, leagueId, sportId, storePlayers)
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

func playerToStorePlayerP(leagueIdStr string, sportIdStr string, p PlayerP) *store.PlayerP {

	leagueId, _ := strconv.Atoi(leagueIdStr)
	sportId, _ := strconv.Atoi(sportIdStr)

	return &store.PlayerP{
		Name: p.Name,
		UserStats: store.UserStats{
			Id:         p.UserStats.Id,
			UserId:     p.UserStats.UserId,
			LeagueId:   leagueId,
			SportId:    sportId,
			MatchCount: p.UserStats.MatchCount,
			WinCount:   p.UserStats.WinCount,
			Elo:        p.UserStats.Elo,
		},
	}
}
