package routes

import (
	"github.com/gin-gonic/gin"

	controllers "github.com/3whalesProg/Strife-go/src/modules/chat"
)

// AuthRouter создает роутер для аутентификации
func ChatRouter(router *gin.RouterGroup) {
	chatController := controllers.NewChatController()

	router.POST("/createChat", chatController.CreateChat)
	router.POST("/addUserToChat", chatController.AddUserToChat)
	router.POST("/sendMessage", chatController.SendMessage)
	router.POST("/getChatMessages", chatController.GetChatMessages)
	router.GET("/getUserChats", chatController.GetUserChats)
}
