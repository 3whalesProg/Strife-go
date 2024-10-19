package controllers

import (
	"log"
	"net/http"

	"github.com/3whalesProg/Strife-go/src/db"
	"github.com/3whalesProg/Strife-go/src/models"
	"github.com/3whalesProg/Strife-go/src/socket"
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
		Title       string `json:"title" binding:"required"`    // Название чата
		UserIDs     []uint `json:"user_ids" binding:"required"` // Список ID пользователей
		IsTetATet   *bool  `json:"is_tet_a_tet"`                // Опциональный параметр: является ли чат личным
		RecipientID *uint  `json:"recipient_id"`
	}

	// Привязываем JSON к структуре
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if (json.IsTetATet != nil && json.RecipientID == nil) || (json.IsTetATet == nil && json.RecipientID != nil) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Both is_tet_a_tet and recipient_id must be provided together"})
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
		"message":      "Chat created successfully",
		"chat_id":      chat.ID,
		"recipient_id": chat.RecipientID,
		"is_tet_a_tet": chat.IsTetATet,
	})
}

func (ac *ChatController) GetCurrentChat(c *gin.Context) {
	var json struct {
		UserID uint `json:"user_ids" binding:"required"` // Список ID пользователей
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
	var chats []models.Chats

	// Сначала находим все чаты текущего пользователя
	if err := db.DB.
		Joins("JOIN user_chats ON user_chats.chat_id = chats.id").
		Where("user_chats.user_id = ?", claims.ID).
		Find(&chats).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user chats"})
		return
	}

	// Ищем среди чатов пользователя нужный, где is_tet_a_tet = true и recipient_id = переданному
	var targetChat *models.Chats
	for _, chat := range chats {
		if chat.IsTetATet && chat.RecipientID == json.UserID {
			targetChat = &chat
			break
		}
	}
	if targetChat == nil {
		newChat := models.Chats{
			Users: []*models.Users{ // Срез указателей на Users
				{ID: claims.ID}, // Указатель на текущего пользователя
			},
			Title:       "Tet-a-tet chat", // Можно передать любое значение для названия
			IsTetATet:   true,             // Чат "тет-а-тет"
			RecipientID: json.UserID,      // Привязываем получателя
		}
		c.JSON(http.StatusOK, gin.H{"message": "Chat created successfully",
			"chat_id":      newChat.ID,
			"recipient_id": newChat.RecipientID,
			"is_tet_a_tet": newChat.IsTetATet})
	}

	// Возвращаем найденный чат
	c.JSON(http.StatusOK, gin.H{"message": "Chat created successfully",
		"chat_id":      targetChat.ID,
		"recipient_id": targetChat.RecipientID,
		"is_tet_a_tet": targetChat.IsTetATet})
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

func (cc *ChatController) GetUserChats(c *gin.Context) {
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
	// Привязка входящих данных JSON
	var user models.Users
	if err := db.DB.Preload("Chats").First(&user, claims.ID).Error; err != nil {
		log.Println("Ошибка получения пользователя:", err)
	}

	// Возвращаем сообщения, привязанные к чату
	c.JSON(http.StatusOK, gin.H{
		"chats": user.Chats,
	})
}

func (ac *ChatController) SendMessage(c *gin.Context) {
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
	socket.SendMessageToChat(chat.ID, message)
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
	router.POST("/getCurrentChat", ac.GetCurrentChat)
	router.GET("/getUserChats", ac.GetUserChats)
}
