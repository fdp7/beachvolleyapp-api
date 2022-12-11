package match

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/fdp7/beachvolleyapp-api/store"
)

type Hdl struct {
	store store.Store
}

func New(store store.Store) Hdl {
	return Hdl{store: store}
}

func (h *Hdl) AddMatch(ctx *gin.Context) {
	err := h.store.AddMatch(ctx, &store.Match{})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to add match",
		})
	}

	ctx.JSON(http.StatusOK, gin.H{})
}
