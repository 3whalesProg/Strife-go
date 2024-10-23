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
	router.POST("/getCurrentChat", chatController.GetCurrentChat)
	router.GET("/getUserChats", chatController.GetUserChats)
	router.POST("/editMessage", chatController.EditChatMessages) // Новый маршрут для редактирования сообщений
	router.POST("/deleteMessage", chatController.DeleteChatMessage)
}
