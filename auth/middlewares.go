package auth

import (
	"github.com/gin-gonic/gin"
)

func Auth() gin.HandlerFunc {

	return func(ctx *gin.Context) {
		tokenString := ctx.GetHeader("Authorization")
		if tokenString == "" {
			ctx.JSON(401, gin.H{
				"message": "request does not contain an access token",
			})
			return
		}

		err := ValidateToken(tokenString)
		if err != nil {
			ctx.JSON(401, gin.H{
				"message": "token validation failed",
			})
			return
		}
		ctx.Next()
	}
}
