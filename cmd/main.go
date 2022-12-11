package main

import (
	"context"
	"github.com/fdp7/beachvolleyapp-api/store"

	"github.com/gin-gonic/gin"

	"github.com/fdp7/beachvolleyapp-api/match"
)

func main() {
	ctx := context.Background()

	if err := store.InitializeDB(ctx, store.MongoDB); err != nil {
		panic(err)
	}

	router := gin.Default()
	router.POST("/match", match.AddMatch)

	router.Run()
}
