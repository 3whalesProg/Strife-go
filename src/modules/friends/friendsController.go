package controllers

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/3whalesProg/Strife-go/src/db"
	"github.com/3whalesProg/Strife-go/src/models"
	"github.com/3whalesProg/Strife-go/src/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FriendController struct{}

func NewFriendController() *FriendController {
	return &FriendController{}
}

// GetUserByID Поиск по id
func GetUserByID(id uint) (*models.Users, error) {
	var user models.Users
	if err := db.DB.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByLogin Поиск по логину
func GetUserByLogin(login string) (*models.Users, error) {
	var user models.Users
	if err := db.DB.Where("login = ?", login).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// AddFriends Добавление в друзья
// AddFriends Добавление в друзья
func AddFriends(user *models.Users, friend *models.Users) error {
	// Проверка: нельзя добавить себя в друзья
	if user.ID == friend.ID {
		return fmt.Errorf("cannot add yourself as a friend")
	}

	// Проверка на существующую дружбу
	exists, err := FriendsExist(user.ID, friend.ID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("friendship already exists")
	}

	// Создаем дружеские связи (двусторонние)
	friendships := []models.Friends{
		{UserID: user.ID, FriendID: friend.ID},
		{UserID: friend.ID, FriendID: user.ID},
	}

	// Используем транзакцию для сохранения целостности данных
	return db.DB.Transaction(func(tx *gorm.DB) error {
		for _, friendship := range friendships {
			if err := tx.Create(&friendship).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// RemoveFriend удаляет друга по логину (жесткое удаление связи)
func (fc *FriendController) RemoveFriend(c *gin.Context) {
	var request struct {
		FriendLogin string `json:"friend_login"` // Логин друга для удаления
	}

	// Парсинг JSON запроса
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем токен пользователя
	token := c.Request.Header.Get("Authorization")
	claims, err := utils.CheckUser(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// Получаем ID текущего пользователя из токена
	userID := claims.ID

	// Поиск друга по логину
	friend, err := GetUserByLogin(request.FriendLogin)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Friend not found"})
		return
	}

	// Полное удаление всех записей дружбы (жесткое удаление)
	if err := db.DB.Unscoped().Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
		userID, friend.ID, friend.ID, userID).Delete(&models.Friends{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Успешное удаление
	c.JSON(http.StatusOK, gin.H{"message": "Friend fully removed"})
}

// GetFriendsByUserId список друзей по ID пользователя
// UserResponse структура для ответа без поля Role и Password
type UserResponse struct {
	ID          uint   `json:"id"`
	Login       string `json:"login"`
	Email       string `json:"email"`
	Nickname    string `json:"nickname"`
	Description string `json:"description"`
	AvatarURL   string `json:"avatar_url"`
	IsFavorite  bool   `json:"is_favorite"` // Поле для обозначения избранного
}

// GetFriendsByUserId список друзей по ID пользователя без роли и пароля
func (fc *FriendController) GetFriendsByUserId(c *gin.Context) {
	token := c.Request.Header.Get("Authorization")
	claims, err := utils.CheckUser(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	userId := claims.ID

	// Получаем список друзей
	var friends []models.Friends
	if err := db.DB.Where("user_id = ?", userId).Find(&friends).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Создаем список друзей с необходимыми данными
	friendList := make([]UserResponse, 0, len(friends))
	var wg sync.WaitGroup

	for _, friend := range friends {
		wg.Add(1)
		go func(friend models.Friends) {
			defer wg.Done()
			var user models.Users
			if err := db.DB.First(&user, friend.FriendID).Error; err == nil {
				friendList = append(friendList, UserResponse{
					ID:          user.ID,
					Login:       user.Login,
					Email:       user.Email,
					Nickname:    user.Nickname,
					Description: user.Description,
					AvatarURL:   user.AvatarURL,
					IsFavorite:  friend.IsFavorite, // Добавляем поле is_favorite
				})
			}
		}(friend)
	}

	wg.Wait()
	c.JSON(http.StatusOK, friendList)
}

// SendFriendRequest отправляет запрос на добавление в друзья
// SendFriendRequest отправляет запрос на добавление в друзья
func (fc *FriendController) SendFriendRequest(c *gin.Context) {
	var request struct {
		RecipientLogin string `json:"recipient_login"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	recipient, err := GetUserByLogin(request.RecipientLogin)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	token := c.Request.Header.Get("Authorization")
	claims, err := utils.CheckUser(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	senderID := claims.ID
	if senderID == recipient.ID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "You cannot send a friend request to yourself"})
		return
	}

	sender, err := GetUserByID(senderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Проверка на существующий запрос на дружбу
	var existingRequest models.FriendRequest
	if err := db.DB.Where("sender_id = ? AND recipient_id = ?", sender.ID, recipient.ID).First(&existingRequest).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Request already exists"})
		return
	} else if err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	friendRequest := models.FriendRequest{
		SenderID:    sender.ID,
		RecipientID: recipient.ID,
	}

	if err := db.DB.Create(&friendRequest).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Friend request sent"})
}

// GetFriendRequests возвращает список запросов на дружбу
func (fc *FriendController) GetFriendRequests(c *gin.Context) {
	token := c.Request.Header.Get("Authorization")
	claims, err := utils.CheckUser(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	userId := claims.ID

	// Создаем структуру для хранения запроса с логинами отправителей
	type FriendRequestWithLogin struct {
		ID          uint   `json:"id"`
		SenderID    uint   `json:"sender_id"`
		RecipientID uint   `json:"recipient_id"`
		SenderLogin string `json:"sender_login"` // Логин отправителя
		CreatedAt   string `json:"created_at"`
		UpdatedAt   string `json:"updated_at"`
	}

	var requestsWithLogins []FriendRequestWithLogin

	// Выполняем запрос с указанием полей (без повторяющихся временных меток GORM)
	if err := db.DB.Table("friend_requests").
		Select("friend_requests.id, friend_requests.sender_id, friend_requests.recipient_id, friend_requests.created_at, friend_requests.updated_at, users.login AS sender_login").
		Joins("join users on users.id = friend_requests.sender_id").
		Where("friend_requests.recipient_id = ?", userId).
		Scan(&requestsWithLogins).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, requestsWithLogins)
}

func (fc *FriendController) RespondToFriendRequest(c *gin.Context) {
	var response struct {
		Accepted    bool   `json:"accepted"`
		SenderLogin string `json:"sender_login"` // Логин отправителя
	}
	if err := c.ShouldBindJSON(&response); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Проверка токена
	token := c.Request.Header.Get("Authorization")
	claims, err := utils.CheckUser(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	userId := claims.ID

	// Получаем отправителя по логину
	var sender models.Users
	if err := db.DB.Where("login = ?", response.SenderLogin).First(&sender).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sender not found"})
		return
	}

	// Поиск запроса на добавление в друзья по ID получателя и ID отправителя
	var friendRequest models.FriendRequest
	if err := db.DB.Where("sender_id = ? AND recipient_id = ?", sender.ID, userId).First(&friendRequest).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Friend request not found"})
		return
	}

	// Проверка существования уже существующего запроса
	existingRequest := models.FriendRequest{}
	if err := db.DB.Where("sender_id = ? AND recipient_id = ?", userId, sender.ID).First(&existingRequest).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Friend request already exists"})
		return
	}

	// Если запрос принят, добавляем друзей
	if response.Accepted {
		recipient, err := GetUserByID(friendRequest.RecipientID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Добавляем в друзья
		if err := AddFriends(&sender, recipient); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Полное удаление запроса (без soft delete)
	if err := db.DB.Unscoped().Delete(&friendRequest).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Response processed"})
}

// ToggleFavoriteFriend переключает статус избранного друга
func (fc *FriendController) ToggleFavoriteFriend(c *gin.Context) {
	var request struct {
		FriendLogin string `json:"friend_login"` // Логин друга
		FriendID    uint   `json:"friend_id"`    // ID друга
	}

	// Парсинг запроса
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Получение пользователя из JWT токена
	token := c.Request.Header.Get("Authorization")
	claims, err := utils.CheckUser(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}
	userID := claims.ID

	var friend models.Users
	// Проверка по ID или логину
	if request.FriendID != 0 {
		// Поиск друга по ID
		friendPtr, err := GetUserByID(request.FriendID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Friend not found by ID"})
			return
		}
		friend = *friendPtr // Разыменовываем указатель на структуру
	} else if request.FriendLogin != "" {
		// Поиск друга по логину
		friendPtr, err := GetUserByLogin(request.FriendLogin)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Friend not found by login"})
			return
		}
		friend = *friendPtr // Разыменовываем указатель на структуру
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Friend ID or login must be provided"})
		return
	}

	// Проверка существования дружбы
	var friendship models.Friends
	if err := db.DB.Where("user_id = ? AND friend_id = ?", userID, friend.ID).First(&friendship).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Friendship not found"})
		return
	}

	// Переключение is_favorite
	friendship.IsFavorite = !friendship.IsFavorite

	// Сохранение изменений
	if err := db.DB.Save(&friendship).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Успешный ответ с новым статусом
	c.JSON(http.StatusOK, gin.H{
		"message":      "Favorite status updated",
		"is_favorite":  friendship.IsFavorite,
		"friend_login": friend.Login,
		"friend_id":    friend.ID,
	})
}

// GetFavoriteFriends возвращает список избранных друзей в формате UserResponse
func (fc *FriendController) GetFavoriteFriends(c *gin.Context) {
	// Извлекаем токен пользователя
	token := c.Request.Header.Get("Authorization")
	claims, err := utils.CheckUser(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	userID := claims.ID

	// Находим всех избранных друзей пользователя
	var favoriteFriends []models.Friends
	if err := db.DB.Where("user_id = ? AND is_favorite = ?", userID, true).Find(&favoriteFriends).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Формируем список в формате UserResponse
	friendList := make([]UserResponse, len(favoriteFriends))
	var wg sync.WaitGroup

	for i, friend := range favoriteFriends {
		wg.Add(1)
		go func(i int, friendID uint) {
			defer wg.Done()
			user, err := GetUserByID(friendID)
			if err == nil {
				friendList[i] = UserResponse{
					ID:          user.ID,
					Login:       user.Login,
					Email:       user.Email,
					Nickname:    user.Nickname,
					Description: user.Description,
					AvatarURL:   user.AvatarURL,
				}
			}
		}(i, friend.FriendID)
	}

	wg.Wait()

	// Возвращаем список избранных друзей в формате UserResponse
	c.JSON(http.StatusOK, friendList)
}

func FriendsExist(userID, friendID uint) (bool, error) {
	var friend models.Friends
	err := db.DB.Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
		userID, friendID, friendID, userID).First(&friend).Error
	if err == gorm.ErrRecordNotFound {
		return false, nil
	}
	return err == nil, err
}

// RegisterRoutes Регистрируем маршруты в контроллере
func (fc *FriendController) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/friends", fc.GetFriendsByUserId)
	router.POST("/request", fc.SendFriendRequest)
	router.GET("/reqlis", fc.GetFriendRequests)
	router.POST("/response", fc.RespondToFriendRequest)
	router.POST("/favorites/toggle", fc.ToggleFavoriteFriend) // Добавление в избранные
	router.GET("/favorites", fc.GetFavoriteFriends)           // Получение списка избранных
}
