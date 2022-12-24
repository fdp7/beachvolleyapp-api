package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"github.com/fdp7/beachvolleyapp-api/player"
	"github.com/fdp7/beachvolleyapp-api/store"
)

var jwtKey []byte

type JWTClaim struct {
	Name string `json:"name"`
	jwt.StandardClaims
}

type TokenRequest struct {
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func init() {
	jwtKeyString := viper.GetString("JWT_KEY")
	jwtKey = []byte(jwtKeyString)
}

func GenerateJWT(name string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &JWTClaim{
		Name: name,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func ValidateToken(signedToken string) error {
	sToken := strings.TrimPrefix(signedToken, "Bearer ")

	token, err := jwt.ParseWithClaims(
		sToken,
		&JWTClaim{},
		func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		},
	)
	if err != nil {
		return fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*JWTClaim)
	if !ok {
		return fmt.Errorf("couldn't parse claims: %w", err)
	}

	if claims.ExpiresAt < time.Now().Local().Unix() {
		return fmt.Errorf("token expired: %w", err)
	}

	return nil
}

func GenerateToken(ctx *gin.Context) {
	var request TokenRequest

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	if store.DB == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "store is not initialized",
		})
		return
	}

	record, err := store.DB.GetPlayer(ctx, request.Name)
	if errors.Is(err, store.ErrNoPlayerFound) || len(record) == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{
			"message": "no player found",
		})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "can't find player",
		})
		return
	}

	user := &player.Player{}
	if err := json.Unmarshal(record, user); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to unmarshal player",
		})
		return
	}

	if credentialError := player.CheckPassword(user, request.Password); credentialError != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"message": "invalid credential",
		})
		return
	}

	tokenString, err := GenerateJWT(user.Name)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to generate token",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func Auth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenString := ctx.GetHeader("Authorization")
		if tokenString == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "request does not contain an access token",
			})
			return
		}

		if err := ValidateToken(tokenString); err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "token validation failed",
			})
			return
		}

		ctx.Next()
	}
}
