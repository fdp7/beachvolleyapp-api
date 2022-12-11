package main

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/fdp7/beachvolleyapp-api/match"
	"github.com/fdp7/beachvolleyapp-api/store/mongodb"
)

func main() {
	ctx := context.Background()

	store, err := mongodb.New(ctx, "mongodb+srv://federico:kObR1eLZfNg0WeN9@federicocluster.2bvfp2w.mongodb.net/test")
	if err != nil {
		panic(err)
	}

	matchHdl := match.NewHandler(store)

	router := gin.Default()
	router.POST("/match", matchHdl.AddMatch)

	router.Run()
}
