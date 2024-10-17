package routes

import (
	controllers "github.com/3whalesProg/Strife-go/src/modules/friends"

	"github.com/gin-gonic/gin"
)

func FriendsRouter(router *gin.RouterGroup) {
	friendsController := controllers.NewFriendController()

	router.GET("/friends", friendsController.GetFriendRequests)
	router.POST("/request", friendsController.SendFriendRequest)
	router.GET("/requests", friendsController.GetFriendRequests)
	router.POST("/response", friendsController.RespondToFriendRequest)
}
