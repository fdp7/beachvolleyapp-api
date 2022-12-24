package player

import (
	"encoding/json"
	"errors"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"github.com/gin-gonic/gin"

	"github.com/fdp7/beachvolleyapp-api/store"
)

func HashPassword(player *Player) error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(player.Password), 1)
	if err != nil {
		return nil
	}
	player.Password = string(bytes)
	return nil
}

func CheckPassword(player *Player, providedPassword string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(player.Password), []byte(providedPassword)); err != nil {
		return err
	}
	return nil
}

func RegisterPlayer(ctx *gin.Context) {
	if store.DB == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "store is not initialized",
		})
		return
	}

	player := &Player{}
	if err := ctx.BindJSON(player); err != nil {
		return
	}

	if err := HashPassword(player); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to encrypt password",
		})
		return
	}

	storePlayer := playerToStorePlayer(player)
	err := store.DB.AddPlayer(ctx, storePlayer)
	if errors.Is(err, store.ErrPlayerDuplicated) {
		ctx.JSON(http.StatusForbidden, gin.H{
			"message": "player already exists",
		})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to add player",
		})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{})
}

func GetPlayer(ctx *gin.Context) {

	name := ctx.Param("name")

	if store.DB == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "store is not initialized",
		})
		return
	}

	result, err := store.DB.GetPlayer(ctx, name)
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
	if store.DB == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "store is not initialized",
		})

		return
	}

	result, err := store.DB.GetRanking(ctx)
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

func playerToStorePlayer(p *Player) *store.Player {
	return &store.Player{
		ID:         p.ID,
		Name:       p.Name,
		Password:   p.Password,
		MatchCount: p.MatchCount,
		WinCount:   p.WinCount,
	}
}
