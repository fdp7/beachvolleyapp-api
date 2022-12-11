package match

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/fdp7/beachvolleyapp-api/store"
)


func AddMatch(ctx *gin.Context) {
	if store.DB == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "store is not initialized",
		})

		return
	}

	err := store.DB.AddMatch(ctx, &store.Match{})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to add match",
		})
	}

	ctx.JSON(http.StatusOK, gin.H{})
}
