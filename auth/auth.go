package auth

import (
	"encoding/json"
	"errors"
	"github.com/dgrijalva/jwt-go"
	"github.com/fdp7/beachvolleyapp-api/player"
	"github.com/fdp7/beachvolleyapp-api/store"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"time"
)

var jwtKey = []byte("key")

type JWTClaim struct {
	Name string `json:"name" bson:"name"`
	jwt.StandardClaims
}

type TokenRequest struct {
	Name     string `json:"name" bson:"name"`
	Password string `json:"password" bson:"password"`
}

func GenerateJWT(name string) (tokenString string, err error) {
	expirationTime := time.Now().Add(1 * time.Hour)
	claims := &JWTClaim{
		Name: name,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString(jwtKey)
	return
}

func ValidateToken(signedToken string) (err error) {
	sToken := strings.Split(signedToken, "Bearer ")
	token, err := jwt.ParseWithClaims(
		sToken[1],
		&JWTClaim{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtKey), nil
		},
	)
	if err != nil {
		return
	}
	claims, ok := token.Claims.(*JWTClaim)
	if !ok {
		err = errors.New("couldn't parse claims")
		return
	}
	if claims.ExpiresAt < time.Now().Local().Unix() {
		err = errors.New("token expired")
		return
	}
	return
}

func GenerateToken(ctx *gin.Context) {

	var request TokenRequest

	if err := ctx.BindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "failed to generate token",
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

	player := &player.Player{}
	if err := json.Unmarshal(record, player); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to unmarshal player",
		})
		return
	}

	if credentialError := player.CheckPassword(request.Password); credentialError != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"message": "invalid credential",
		})
		return
	}

	tokenString, err := GenerateJWT(player.Name)
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
