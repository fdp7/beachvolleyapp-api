package user

import (
	"errors"
	"github.com/fdp7/beachvolleyapp-api/store"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

func HashPassword(user *User) error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(user.Password), 1)
	if err != nil {
		return nil
	}
	user.Password = string(bytes)
	return nil
}

func CheckPassword(user *User, providedPassword string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(providedPassword)); err != nil {
		return err
	}
	return nil
}

func RegisterUser(ctx *gin.Context) {
	if store.DB == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "store is not initialized",
		})
		return
	}

	user := &User{}
	if err := ctx.BindJSON(user); err != nil {
		return
	}

	if err := HashPassword(user); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to encrypt password",
		})
		return
	}

	storeUser := userToStoreUser(user)
	err := store.DB.AddUser(ctx, storeUser)
	if errors.Is(err, store.ErrUserDuplicated) {
		ctx.JSON(http.StatusForbidden, gin.H{
			"message": "player already exists",
		})
		return
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to add player",
		})
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{})
}

func userToStoreUser(u *User) *store.User {
	return &store.User{
		ID:       u.ID,
		Name:     u.Name,
		Password: u.Password,
	}
}
