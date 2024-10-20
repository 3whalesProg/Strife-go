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

	for _, userID := range json.UserIDs {
		socket.AddUserToChat(chat.ID, userID)
	}
	// Возвращаем успешный ответ
	c.JSON(http.StatusOK, gin.H{
		"message":      "Chat created successfully",
		"chat_id":      chat.ID,
		"recipient_id": chat.RecipientID,
		"is_tet_a_tet": chat.IsTetATet,
	})
}

func containsUsers(users []*models.Users, userID uint, user2ID uint) bool {
	var foundUser1, foundUser2 bool

	// Проходим по всем пользователям в чате
	for _, user := range users {
		// Проверяем наличие userID
		if user.ID == userID {
			foundUser1 = true
		}
		// Проверяем наличие user2ID
		if user.ID == user2ID {
			foundUser2 = true
		}

		// Если оба найдены, можно сразу вернуть true
		if foundUser1 && foundUser2 {
			return true
		}
	}

	// Если один из пользователей не найден, возвращаем false
	return false
}
func (ac *ChatController) GetCurrentChat(c *gin.Context) {
	var json struct {
		UserID uint `json:"user_ids" binding:"required"` // ID получателя
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

	var user models.Users
	if err := db.DB.
		Preload("Chats.Users"). // Загружаем связанные чаты и пользователей в этих чатах
		First(&user, claims.ID).Error; err != nil {
		log.Println("Ошибка получения пользователя:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving user"})
		return
	}

	var targetChat *models.Chats
	for _, chat := range user.Chats {
		if chat.IsTetATet {
			if chat.RecipientID == json.UserID || containsUsers(chat.Users, claims.ID, json.UserID) {
				targetChat = chat
				break
			}
		}
	}

	var user2 models.Users
	if err := db.DB.
		Preload("Chats.Users"). // Загружаем связанные чаты и пользователей в этих чатах
		First(&user2, json.UserID).Error; err != nil {
		log.Println("Ошибка получения пользователя:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving user"})
		return
	}

	for _, chat := range user2.Chats {
		if chat.IsTetATet {
			if chat.RecipientID == claims.ID || containsUsers(chat.Users, claims.ID, json.UserID) {
				targetChat = chat
				break
			}
		}
	}

	UserIDs := []uint{claims.ID} // Используем ID текущего пользователя и получателя
	var users []models.Users
	if err := db.DB.Where("id IN ?", UserIDs).Find(&users).Error; err != nil || len(users) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Some users not found"})
		return
	}

	var userPointers []*models.Users
	for i := range users {
		userPointers = append(userPointers, &users[i])
	}

	var recipient models.Users
	if err := db.DB.Where("id = ?", json.UserID).First(&recipient).Error; err != nil {
		// Если пользователь не найден, возвращаем ошибку
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if targetChat == nil {
		newChat := models.Chats{
			Users:       userPointers,
			Title:       "Tet-a-tet chat", // Название чата
			IsTetATet:   true,             // Чат "тет-а-тет"
			RecipientID: json.UserID,      // Привязываем получателя
		}

		// Сохраняем новый чат в базе данных
		if err := db.DB.Create(&newChat).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create chat"})
			return
		}
		// Возвращаем информацию о созданном чате
		socket.AddUserToChat(newChat.ID, claims.ID)
		c.JSON(http.StatusOK, gin.H{
			"message":      "Chat created successfully",
			"chat_id":      newChat.ID,
			"recipient_id": newChat.RecipientID,
			"is_tet_a_tet": newChat.IsTetATet,
			"recipient":    recipient,
		})
		return
	}
	socket.AddUserToChat(targetChat.ID, claims.ID)
	// Возвращаем найденный чат
	c.JSON(http.StatusOK, gin.H{
		"message":      "Chat found successfully",
		"chat_id":      targetChat.ID,
		"recipient_id": targetChat.RecipientID,
		"is_tet_a_tet": targetChat.IsTetATet,
		"recipient":    recipient,
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
	socket.AddUserToChat(chat.ID, json.UserID)
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
	if err := db.DB.
		Preload("Chats.Users").
		First(&user, claims.ID).Error; err != nil {
		log.Println("Ошибка получения пользователя:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error retrieving user"})
		return
	}

	for i, chat := range user.Chats {
		// Если чат личный (IsTetATet == true)
		if chat.IsTetATet {
			for _, chatUser := range chat.Users {
				// Если пользователь не тот, что в claims (собеседник)
				if chatUser.ID != claims.ID {
					user.Chats[i].RecipientID = chatUser.ID
					user.Chats[i].Recipient = chatUser
					// Заменяем title на ник собеседника
					user.Chats[i].Title = chatUser.Nickname
					break
				}
			}
		}
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
		Offset int  `json:"offset" binding:"required`   // Смещение для пагинации (по умолчанию 0)
	}

	// Привязка входящих данных JSON
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "chat_id and limit are required"})
		return
	}

	// Проверяем, что лимит положительный
	if json.Offset <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Offset must be greater than 0"})
		return
	}

	// Загрузка чата с привязанными сообщениями с использованием offset и limit
	var chat models.Chats
	if err := db.DB.Preload("Messages.Sender").
		First(&chat, json.ChatID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Chat not found"})
		return
	}

	// Загрузка сообщений с использованием offset и limit
	var messages []models.Messages
	if err := db.DB.Where("chat_id = ?", json.ChatID).
		Preload("Sender").
		Order("created_at DESC"). // Сортировка по времени создания сообщений
		Limit(50).
		Offset(json.Offset).
		Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load messages"})
		return
	}

	// Возвращаем сообщения с учетом лимита и смещения
	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
	})
}

func (cc *ChatController) EditChatMessages(c *gin.Context) {
	var json struct {
		MessageID uint   `json:"message_id" binding:"required"` // ID сообщения
		Content   string `json:"content" binding:"required"`    // Новый текст сообщения
	}

	// Привязка входящих данных JSON
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data"})
		return
	}

	// Проверка токена
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token not provided"})
		return
	}
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// Получаем данные пользователя из токена
	claims, err := utils.CheckUser(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// Ищем сообщение по его ID
	var message models.Messages
	if err := db.DB.First(&message, json.MessageID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}

	// Проверяем, что текущий пользователь является автором сообщения
	if message.SenderID != claims.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only edit your own messages"})
		return
	}

	// Обновляем текст сообщения
	message.Content = json.Content
	if err := db.DB.Save(&message).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update message"})
		return
	}

	// Возвращаем обновленное сообщение
	c.JSON(http.StatusOK, gin.H{
		"message_id": message.ID,
		"content":    message.Content,
		"updated_at": message.UpdatedAt,
	})
}

func (cc *ChatController) DeleteChatMessage(c *gin.Context) {
	var json struct {
		MessageID uint `json:"message_id" binding:"required"` // ID сообщения
	}

	// Привязка входящих данных JSON
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input data"})
		return
	}

	// Проверка токена
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token not provided"})
		return
	}
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// Получаем данные пользователя из токена
	claims, err := utils.CheckUser(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// Ищем сообщение по его ID
	var message models.Messages
	if err := db.DB.First(&message, json.MessageID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}

	// Проверяем, что текущий пользователь является автором сообщения
	if message.SenderID != claims.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only delete your own messages"})
		return
	}

	// Удаляем сообщение
	if err := db.DB.Delete(&message).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete message"})
		return
	}

	// Возвращаем успешный ответ
	c.JSON(http.StatusOK, gin.H{
		"message":    "Message deleted successfully",
		"message_id": message.ID,
	})
}

func (ac *ChatController) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/createChat", ac.CreateChat)
	router.POST("/addUserToChat", ac.AddUserToChat)
	router.POST("/sendMessage", ac.SendMessage)
	router.POST("/getChatMessages", ac.GetChatMessages)
	router.POST("/getCurrentChat", ac.GetCurrentChat)
	router.GET("/getUserChats", ac.GetUserChats)
	router.POST("/editMessage", ac.EditChatMessages)    // Новый маршрут для редактирования сообщений
	router.POST("/deleteMessage", ac.DeleteChatMessage) // Новый маршрут для удаления сообщений
}
