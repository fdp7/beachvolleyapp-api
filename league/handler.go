package league

import (
	"fmt"
	"github.com/fdp7/beachvolleyapp-api/store"
	"github.com/gin-gonic/gin"
	"net/http"
)

const (
	userQueryParam       = "user"
	targetUserQueryParam = "targetUser"
)

/*
- The user that creates the league becomes its first participant automatically
- The user that creates the league becomes admin and founder
- An admin can use AddUser and RemoveUser api
- Founder only can use DeleteLeague api
- Founder only can make others admin
- Anyone can abandon the league, but if the founder abandon the league, then an admin should become founder,
while if no admin is present, a normal user should be made both founder and admin. If no one is left in the league, then
it's deleted
*/
func AddLeague(ctx *gin.Context) {

	if store.DBSql == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "store is not initialized",
		})

		return
	}

	league := &League{}
	if err := ctx.BindJSON(league); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid league data",
		})
		return
	}

	user := ctx.Request.URL.Query().Get(userQueryParam)

	storeLeagueU := leagueToStoreLeagueU(league, user)

	err := store.DBSql.AddLeague(ctx, storeLeagueU)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to add league",
		})
	}
	ctx.JSON(http.StatusCreated, gin.H{
		"message": fmt.Sprintf("league named %s has been created", league.Name),
	})
}

func AbandonLeague(ctx *gin.Context) {
	leagueId := ctx.Param("leagueId")
	userName := ctx.Request.URL.Query().Get(userQueryParam)

	if store.DBSql == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "store is not initialized",
		})

		return
	}

	err := store.DBSql.AbandonLeague(ctx, userName, leagueId)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to abandon league",
		})

		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("user %s successfully abandoned league %s", userName, leagueId),
	})

}

func GetUserLeagues(ctx *gin.Context) {
	// get all the leagues the user is part of, so that it can select one to navigate.
	// Must exploit UserLeague table
}

// delete league if you're the founder
func DeleteLeague(ctx *gin.Context) {
	leagueId := ctx.Param("leagueId")
	userName := ctx.Request.URL.Query().Get(userQueryParam)

	if store.DBSql == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "store is not initialized",
		})

		return
	}

	if isFounder, _ := store.DBSql.IsFounder(ctx, userName, leagueId); isFounder == true {
		err := store.DBSql.DeleteLeague(ctx, leagueId)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": "failed to delete league",
			})
		}
		ctx.JSON(http.StatusOK, gin.H{
			"message": "league has been deleted",
		})
	} else {
		ctx.JSON(http.StatusForbidden, gin.H{
			"message": fmt.Sprintf("user %s can't delete the league", userName),
		})
	}

}

// when adding a user to a league you must initialize his stats for that league with default values, for all sports
func AddUserToLeague(ctx *gin.Context) {

	leagueId := ctx.Param("leagueId")
	userName := ctx.Request.URL.Query().Get(userQueryParam)
	userToAdd := ctx.Request.URL.Query().Get(targetUserQueryParam)

	if store.DBSql == nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "store is not initialized",
		})

		return
	}

	if isAdmin, _ := store.DBSql.IsAdmin(ctx, userName, leagueId); isAdmin == true {
		err := store.DBSql.AddUserToLeague(ctx, userToAdd, leagueId)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": "failed to add user to league",
			})

			return
		}
		ctx.JSON(http.StatusOK, gin.H{
			"message": "user has been added",
		})
	} else {
		ctx.JSON(http.StatusForbidden, gin.H{
			"message": fmt.Sprintf("user %s can't add users to league %s", userName, leagueId),
		})

	}
}

func RemoveUserFromLeague(ctx *gin.Context) {
	// deleting a user from a league should be a soft delete so that if you add it back it will regain all its historical stats
}

func MakeAdmin(ctx *gin.Context) {
	// founder can make another user of the league as admin. The user should already be added
	// as query param there is the username of the one invoking the api (that must be the founder); in the body there
	// must be the LIST of usernames to be made admin
}

func leagueToStoreLeagueU(l *League, user string) *store.LeagueU {
	return &store.LeagueU{
		Name: l.Name,
		User: user,
	}
}
