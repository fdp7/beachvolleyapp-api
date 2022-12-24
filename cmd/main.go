package main

import (
	"context"
	"github.com/fdp7/beachvolleyapp-api/auth"
	"github.com/fdp7/beachvolleyapp-api/match"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"github.com/fdp7/beachvolleyapp-api/player"
	"github.com/fdp7/beachvolleyapp-api/store"
)

func main() {

	viper.SetConfigFile("app.env")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error while reading configuration file %s", err)
	}

	ctx := context.Background()

	if err := store.InitializeDB(ctx, store.MongoDB); err != nil {
		panic(err)
	}

	router := gin.Default()

	router.POST("/player/login", auth.GenerateToken)

	// in secured all api that must be checked using a valid token
	secured := router.Use(auth.Auth())
	{
		secured.GET("/ping", player.Ping)

		secured.GET("/matches", match.GetMatches)
		secured.POST("/match", match.AddMatch)
		secured.DELETE("/match", match.DeleteMatch)

		secured.GET("/player/:name", player.GetPlayer)
		secured.GET("/ranking", player.GetRanking)
		secured.POST("/player/signup", player.RegisterPlayer)
	}

	router.Run()
}
