package routes

import (
	controllers "github.com/3whalesProg/Strife-go/src/modules/friends"

	"github.com/gin-gonic/gin"
)

// FriendsRouter Регистрируем маршруты для работы с друзьями
func FriendsRouter(router *gin.RouterGroup) {
	friendsController := controllers.NewFriendController()

	// Регистрация всех маршрутов, соответствующих методам FriendController
	router.GET("/friends", friendsController.GetFriendsByUserId)             // Получение списка друзей
	router.POST("/request", friendsController.SendFriendRequest)             // Отправка запроса в друзья
	router.GET("/reqlis", friendsController.GetFriendRequests)               // Получение списка запросов
	router.POST("/response", friendsController.RespondToFriendRequest)       // Ответ на запрос в друзья
	router.DELETE("/dfriend", friendsController.RemoveFriend)                // Удаление друга
	router.POST("/favorites/toggle", friendsController.ToggleFavoriteFriend) // Добавление в избранные
	router.GET("/favorites", friendsController.GetFavoriteFriends)
	router.POST("/common_friends", friendsController.GetCommonFriends)

}
