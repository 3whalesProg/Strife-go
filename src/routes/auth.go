package routes

import (
	"github.com/gin-gonic/gin"

	controllers "github.com/3whalesProg/Strife-go/src/modules/auth"
)

// AuthRouter создает роутер для аутентификации
func AuthRouter(router *gin.RouterGroup) {
	authController := controllers.NewAuthController()

	router.POST("/login", func(c *gin.Context) {
		// Логика для входа пользователя
		c.JSON(200, gin.H{"message": "Login successful"})
	})

	router.POST("/register", authController.Register)
}
