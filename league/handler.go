package league

import (
	"github.com/gin-gonic/gin"
)

func AddLeague(ctx *gin.Context) {
	// the user that creates the league becomes its first participant automatically
	// Maybe should also be made admin and founder with the power to make others admin too.
	// An admin can use AddUser and DeleteUser api. Founder (only!!!) can use DeleteLeague api.
	// Anyone can abandon the league

	// If a care about the role, I must change UserLeague table adding two bool columns: IsFounder and IsAdmin
	// or only one column for Role where it's written "founder"/"admin"/"user"
	// or one column for PermissionLevel 300/200/100, so that I can control the
	// usage of api by checking these columns
}

func AbandonLeague(ctx *gin.Context) {
}

func GetUserLeagues(ctx *gin.Context) {
	// get all the leagues the user is part of, so that it can select one to navigate.
	// Must exploit UserLeague table
}

func DeleteLeague(ctx *gin.Context) {
}

func AddUser(ctx *gin.Context) {
	// when adding a user to a league you must initialize his stats for that league with default values, for all sports
}

func DeleteUser(ctx *gin.Context) {
	// deleting a user from a league should be a soft delete so that if you add it back it will regain all its historical stats
}
