package match

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/fdp7/beachvolleyapp-api/store"
)

type Handler struct {
	store store.Store
}

func NewHandler(store store.Store) Handler {
	return Handler{store: store}
}

func (h *Handler) AddMatch(ctx *gin.Context) {
	err := h.store.AddMatch(ctx, &store.Match{})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to add match",
			"error": err.Error(),
		})
	}

	ctx.JSON(http.StatusOK, gin.H{})
}
