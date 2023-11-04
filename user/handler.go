package user

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/fdp7/beachvolleyapp-api/store"
)

func HashPassword(user *UserP) error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(user.Password), 1)
	if err != nil {
		return nil
	}
	user.Password = string(bytes)
	return nil
}

func CheckPassword(user *UserP, providedPassword string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(providedPassword)); err != nil {
		return err
	}
	return nil
}

func RegisterUser(ctx *gin.Context) {
	if store.DBUser == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "store is not initialized",
		})
		return
	}

	user := &UserP{}
	if err := ctx.BindJSON(user); err != nil {
		return
	}

	if err := HashPassword(user); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to encrypt password",
		})

		return
	}

	storeUser := UserToStoreUserP(user)

	//check duplicate name, then add new user
	u, err := store.DBUser.GetUser(ctx, storeUser.Name)

	if errors.Is(err, store.ErrNoUserFound) {

		err = store.DBUser.AddUser(ctx, storeUser)

		if errors.Is(err, store.ErrUserDuplicated) {
			ctx.JSON(http.StatusForbidden, gin.H{
				"message": "user already exists",
			})

			return
		}
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": "failed to add user",
			})

			return
		}
	}
	if len(u) != 0 {
		ctx.JSON(http.StatusForbidden, gin.H{
			"message": "user already exists",
		})

		return
	}

	/*err = store.DBSport.AddUserToSportDBs(ctx, storeUser)
	if errors.Is(err, store.ErrPlayerDuplicated) {
		ctx.JSON(http.StatusForbidden, gin.H{
			"message": "player already exist",
		})

		return
	}
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to add player",
		})

		return
	}*/

	ctx.JSON(http.StatusCreated, gin.H{})

}

func UserToStoreUser(u *User) *store.User {
	return &store.User{
		ID:       u.ID,
		Name:     u.Name,
		Password: u.Password,
	}
}

func UserToStoreUserP(u *UserP) *store.UserP {
	return &store.UserP{
		Id:       u.Id,
		Name:     u.Name,
		Password: u.Password,
		Email:    u.Email,
	}
}
