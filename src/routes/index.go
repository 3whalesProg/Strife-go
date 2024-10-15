package routes

import (
	"github.com/gin-gonic/gin"
)

// IndexRouter создает роутер для API v1
func IndexRouter(router *gin.RouterGroup) {
	// Группа маршрутов для аутентификации
	authGroup := router.Group("/auth")
	AuthRouter(authGroup) // Подключаем роуты аутентификации

	// Здесь можно добавить другие группы маршрутов для v1
}
