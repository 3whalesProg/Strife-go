package controllers

import (
	"net/http"

	"github.com/3whalesProg/Strife-go/src/db"
	"github.com/3whalesProg/Strife-go/src/models"
	"github.com/3whalesProg/Strife-go/src/socket"
	"github.com/3whalesProg/Strife-go/src/utils"
	"github.com/gin-gonic/gin"
)

// AuthController содержит методы для аутентификации
type VoiceController struct{}

// NewAuthController создает новый экземпляр AuthController
func NewVoiceController() *VoiceController {
	return &VoiceController{}
}

// Register обрабатывает регистрацию пользователя
func (vc *VoiceController) CreateRoom(c *gin.Context) {
	var json struct {
		ChatID uint `json:"chat_id" binding:"required"` // ID чата
	}
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token not provided"})
		return
	}

	// Убираем "Bearer " из начала токена, если оно есть
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// Проверяем токен и получаем информацию из Claims
	claims, err := utils.CheckUser(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем пользователя из базы данных по ID, который хранится в JWT токене
	var user models.Users
	if err := db.DB.First(&user, claims.ID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	socket.CreateRoom(claims.ID, json.ChatID)
	// Возвращаем информацию о пользователе
	c.JSON(http.StatusOK, gin.H{
		"RoomID": json.ChatID,
	})
}

func (vc *VoiceController) JoinRoom(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token not provided"})
		return
	}

	// Убираем "Bearer " из начала токена, если оно есть
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// Проверяем токен и получаем информацию из Claims
	claims, err := utils.CheckUser(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Получаем пользователя из базы данных по ID, который хранится в JWT токене
	var user models.Users
	if err := db.DB.First(&user, claims.ID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Возвращаем информацию о пользователе
	c.JSON(http.StatusOK, gin.H{
		"id":       user.ID,
		"login":    user.Login,
		"email":    user.Email,
		"nickname": user.Nickname,
		"role":     user.Role,
	})
}

// RegisterRoutes регистрирует маршруты контроллера
func (vc *VoiceController) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/createRoom", vc.CreateRoom)
	router.POST("/joinRoom", vc.JoinRoom)
}
