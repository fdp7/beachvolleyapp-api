package main

import (
	"context"
	"github.com/fdp7/beachvolleyapp-api/match"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"github.com/fdp7/beachvolleyapp-api/auth"
	"github.com/fdp7/beachvolleyapp-api/player"
	"github.com/fdp7/beachvolleyapp-api/store"
	"github.com/fdp7/beachvolleyapp-api/user"
)

func main() {
	ctx := context.Background()

	viper.SetConfigFile("app.env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("error while reading configuration file: %s\n", err.Error())
		log.Println("application will run without configurations from config file")
	}

	/*if err := store.InitializeDB(ctx, store.MongoDB); err != nil {
		log.Fatalf("failed to initialize DB: %s", err.Error())
	}*/

	if err := store.InitializeDB(ctx, store.Postgres); err != nil {
		log.Fatalf("failed to initialize DB: %s", err.Error())
	}

	router := gin.Default()

	// the same for MongoDB and PostgreSql
	router.POST("/user/signup", user.RegisterUser)
	router.POST("/user/login", auth.GenerateToken)

	// in secured all api that must be checked using a valid token

	//MongoDB API
	/*secured := router.Use(auth.Auth())
	{
		// MATCH
		secured.GET("/:sport/matches", match.GetMatches)

		secured.POST("/:sport/match", match.AddMatch)
		secured.DELETE("/:sport/match", match.DeleteMatch)

		// PLAYER
		secured.GET("/:sport/players", player.GetPlayers)
		secured.POST("/:sport/players/balanceTeams", player.GenerateBalancedTeams)

		secured.GET("/:sport/player/:name", player.GetPlayer)
		secured.GET("/:sport/player/ranking", player.GetRanking)
		secured.GET("/:sport/player/:name/mates", player.GetMates)
	}*/

	// PostgreSql API
	secured := router.Use(auth.Auth())
	{
		secured.GET("/:leagueId/:sportId/players", player.GetPlayers)
		secured.GET("/:leagueId/:sportId/players/:name", player.GetPlayer)
		secured.GET("/:leagueId/:sportId/players/ranking", player.GetRanking)
		secured.GET("/:leagueId/:sportId/players/:name/mates", player.GetMates)
		secured.POST("/:leagueId/:sportId/players/balanceTeams", player.GenerateBalancedTeams)
		secured.GET("/:leagueId/:sportId/matches", match.GetMatches)
	}

	router.Run()
}
