package controllers

import (
	"net/http"

	"github.com/3whalesProg/Strife-go/src/db"
	"github.com/3whalesProg/Strife-go/src/models"
	"github.com/3whalesProg/Strife-go/src/utils"
	"github.com/gin-gonic/gin"
)

// ChatController содержит методы для работы с чатами
type ChatController struct{}

// NewChatController создает новый экземпляр ChatController
func NewChatController() *ChatController {
	return &ChatController{}
}

// CreateChat создает новый чат
func (ac *ChatController) CreateChat(c *gin.Context) {
	var json struct {
		Title   string `json:"title" binding:"required"`    // Название чата
		UserIDs []uint `json:"user_ids" binding:"required"` // Список ID пользователей
	}

	// Привязываем JSON к структуре
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверяем, что передан хотя бы один пользователь
	if len(json.UserIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one user must be specified"})
		return
	}

	// Получаем список пользователей по их ID
	var users []models.Users
	if err := db.DB.Where("id IN ?", json.UserIDs).Find(&users).Error; err != nil || len(users) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Some users not found"})
		return
	}

	var userPointers []*models.Users
	for i := range users {
		userPointers = append(userPointers, &users[i])
	}

	// Создаем новый чат и связываем его с пользователями
	chat := models.Chats{
		Title: json.Title,
		Users: userPointers, // Связываем указатели на пользователей с чатом
	}

	// Сохраняем чат в базе данных
	if err := db.DB.Create(&chat).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create chat"})
		return
	}

	// Возвращаем успешный ответ
	c.JSON(http.StatusOK, gin.H{
		"message": "Chat created successfully",
		"chat_id": chat.ID,
	})
}

func (ac *ChatController) AddUserToChat(c *gin.Context) {
	var json struct {
		ChatID uint `json:"chat_id" binding:"required"` // ID чата
		UserID uint `json:"user_id" binding:"required"` // ID пользователя
	}

	// Привязываем JSON к структуре
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ищем чат по его ID
	var chat models.Chats
	if err := db.DB.First(&chat, json.ChatID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chat not found"})
		return
	}

	// Ищем пользователя по его ID
	var user models.Users
	if err := db.DB.First(&user, json.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Добавляем пользователя в чат (связь many2many)
	if err := db.DB.Model(&chat).Association("Users").Append(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add user to chat"})
		return
	}

	// Возвращаем успешный ответ
	c.JSON(http.StatusOK, gin.H{
		"message": "User added to chat successfully",
		"chat_id": chat.ID,
		"user_id": user.ID,
	})
}

func (cc *ChatController) SendMessage(c *gin.Context) {
	var json struct {
		ChatID  uint   `json:"chat_id" binding:"required"` // ID чата
		Content string `json:"content" binding:"required"` // Текст сообщения
	}

	// Привязываем JSON к структуре
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
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

	// Ищем чат по его ID
	var chat models.Chats
	if err := db.DB.First(&chat, json.ChatID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chat not found"})
		return
	}

	// Ищем отправителя по его ID
	var sender models.Users
	if err := db.DB.First(&sender, claims.ID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sender not found"})
		return
	}

	// Создаем новое сообщение
	message := models.Messages{
		Content:  json.Content,
		SenderID: sender.ID,
		ChatID:   chat.ID,
	}

	// Сохраняем сообщение в базе данных
	if err := db.DB.Create(&message).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message"})
		return
	}
	// socket.Hello()
	// Возвращаем успешный ответ с данными о сообщении
	c.JSON(http.StatusOK, gin.H{
		"message_id": message.ID,
		"content":    message.Content,
		"chat_id":    chat.ID,
		"sender_id":  sender.ID,
		"created_at": message.CreatedAt,
	})
}

func (cc *ChatController) GetChatMessages(c *gin.Context) {
	var json struct {
		ChatID uint `json:"chat_id" binding:"required"` // ID чата
	}

	// Привязка входящих данных JSON
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "chat_id is required"})
		return
	}

	// Загрузка чата с привязанными сообщениями
	var chat models.Chats
	if err := db.DB.Preload("Messages.Sender").First(&chat, json.ChatID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chat not found"})
		return
	}

	// Возвращаем сообщения, привязанные к чату
	c.JSON(http.StatusOK, gin.H{
		"messages": chat.Messages,
	})
}

// RegisterRoutes регистрирует маршруты для ChatController
func (ac *ChatController) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/createChat", ac.CreateChat)
	router.POST("/addUserToChat", ac.AddUserToChat)
	router.POST("/sendMessage", ac.AddUserToChat)
	router.POST("/getChatMessages", ac.GetChatMessages)
}
