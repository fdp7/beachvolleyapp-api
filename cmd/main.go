package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"github.com/fdp7/beachvolleyapp-api/match"
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
	router.POST("/match", match.AddMatch)

	router.Run()
}