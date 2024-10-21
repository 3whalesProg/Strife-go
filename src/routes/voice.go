package routes

import (
	"github.com/gin-gonic/gin"

	controllers "github.com/3whalesProg/Strife-go/src/modules/voice"
)

// AuthRouter создает роутер для аутентификации
func VoiceRouter(router *gin.RouterGroup) {
	voiceController := controllers.NewVoiceController()

	router.POST("/createRoom", voiceController.CreateRoom)
	router.POST("/joinRoom", voiceController.JoinRoom)
}
