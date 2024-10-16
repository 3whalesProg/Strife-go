package routes

import (
	"github.com/gin-gonic/gin"

	controllers "github.com/3whalesProg/Strife-go/src/modules/auth"
)

// AuthRouter создает роутер для аутентификации
func AuthRouter(router *gin.RouterGroup) {
	authController := controllers.NewAuthController()

	router.POST("/login", authController.Login)

	router.POST("/register", authController.Register)
}
