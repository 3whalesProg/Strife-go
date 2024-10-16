package routes

import (
	"github.com/gin-gonic/gin"

	controllers "github.com/3whalesProg/Strife-go/src/modules/user"
)

// AuthRouter создает роутер для аутентификации
func UserRouter(router *gin.RouterGroup) {
	userController := controllers.NewUserController()

	router.GET("/getUserInfo", userController.GetUserInfo)
	router.PATCH("/cname", userController.CName)
	router.PATCH("/description", userController.UpdateDescription)
}
