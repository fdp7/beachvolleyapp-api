package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"github.com/fdp7/beachvolleyapp-api/auth"
	"github.com/fdp7/beachvolleyapp-api/match"
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

	if err := store.InitializeDB(ctx, store.MongoDB); err != nil {
		log.Fatalf("failed to initialize DB: %s", err.Error())
	}

	router := gin.Default()

	router.POST("/user/signup", user.RegisterUser)
	router.POST("/user/login", auth.GenerateToken)

	// in secured all api that must be checked using a valid token
	secured := router.Use(auth.Auth())
	{
		secured.GET("/:sport/matches", match.GetMatches)
		secured.POST("/:sport/match", match.AddMatch)
		secured.DELETE("/:sport/match", match.DeleteMatch)

		secured.GET("/:sport/player/:name", player.GetPlayer)
		secured.GET("/:sport/player/ranking", player.GetRanking)
	}

	router.Run()
}
