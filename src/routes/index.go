package routes

import (
	"github.com/gin-gonic/gin"
)

// IndexRouter создает роутер для API v1
func IndexRouter(router *gin.RouterGroup) {
	// Группа маршрутов для аутентификации
	authGroup := router.Group("/auth")
	AuthRouter(authGroup) // Подключаем роуты аутентификации

	userGroup := router.Group("/user")
	UserRouter(userGroup)
	friendGroup := router.Group("/friend")
	FriendsRouter(friendGroup)

	chatGroup := router.Group("/chat")
	ChatRouter(chatGroup)
	// Здесь можно добавить другие группы маршрутов для v1
}
