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

// GetUserByLogin Поиск по логинку
func GetUserByLogin(login string) (*models.Users, error) {
	var user models.Users
	if err := db.DB.Where("login = ?", login).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// AddFriends Добавление в друзья
func AddFriends(user *models.Users, friend *models.Users) error {
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
		var wg sync.WaitGroup
		errs := make(chan error, len(friendships))

		for _, friendship := range friendships {
			wg.Add(1)
			go func(f models.Friends) {
				defer wg.Done()
				if err := tx.Create(&f).Error; err != nil {
					errs <- err
				}
			}(friendship)
		}

		wg.Wait()
		close(errs)

		for err := range errs {
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// DeleteFriendRequest удаляет запрос на дружбу
func (fc *FriendController) DeleteFriendRequest(c *gin.Context) {
	var request struct {
		RecipientID uint `json:"recipient_id"` // ID пользователя, который получил запрос
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token := c.Request.Header.Get("Authorization")
	claims, err := utils.CheckUser(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	senderID := claims.ID

	// Удаление запроса на дружбу
	if err := db.DB.Where("sender_id = ? AND recipient_id = ?", senderID, request.RecipientID).Delete(&models.FriendRequest{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Friend request deleted"})
}

// RemoveFriend удаляет друга (разрывает связь)
func (fc *FriendController) RemoveFriend(c *gin.Context) {
	var request struct {
		FriendID uint `json:"friend_id"` // ID друга для удаления
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token := c.Request.Header.Get("Authorization")
	claims, err := utils.CheckUser(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	userID := claims.ID

	// Удаление обеих записей дружбы (двусторонняя связь)
	if err := db.DB.Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)",
		userID, request.FriendID, request.FriendID, userID).Delete(&models.Friends{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Friend removed"})
}

// GetFriendsByUserId список друзей по айди юзера
func (fc *FriendController) GetFriendsByUserId(c *gin.Context) {
	token := c.Request.Header.Get("Authorization")
	claims, err := utils.CheckUser(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	userId := claims.ID

	var friends []models.Friends
	if err := db.DB.Where("user_id = ?", userId).Find(&friends).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	friendList := make([]models.Users, len(friends))
	var wg sync.WaitGroup

	for i, friend := range friends {
		wg.Add(1)
		go func(i int, friendID uint) {
			defer wg.Done()
			user, err := GetUserByID(friendID)
			if err == nil {
				friendList[i] = *user
			}
		}(i, friend.FriendID)
	}

	wg.Wait()
	c.JSON(http.StatusOK, friendList)
}

// SendFriendRequest запрос в добавку в др
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
	sender, err := GetUserByID(senderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// проверка есть или нет frindreq
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

// GetFriendRequests список запросов на дружбу
func (fc *FriendController) GetFriendRequests(c *gin.Context) {
	token := c.Request.Header.Get("Authorization")
	claims, err := utils.CheckUser(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	userId := claims.ID

	var friendRequests []models.FriendRequest
	if err := db.DB.Where("recipient_id = ?", userId).Find(&friendRequests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, friendRequests)
}

// обрабатывает ответ на запрос в др
func (fc *FriendController) RespondToFriendRequest(c *gin.Context) {
	var response struct {
		Accepted bool `json:"accepted"`
	}
	if err := c.ShouldBindJSON(&response); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token := c.Request.Header.Get("Authorization")
	claims, err := utils.CheckUser(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	userId := claims.ID

	var friendRequest models.FriendRequest
	if err := db.DB.Where("recipient_id = ?", userId).First(&friendRequest).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Friend request not found"})
		return
	}

	if response.Accepted {
		sender, err := GetUserByID(friendRequest.SenderID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		recipient, err := GetUserByID(friendRequest.RecipientID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if err := AddFriends(sender, recipient); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if err := db.DB.Delete(&friendRequest).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Response processed"})
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

// RegisterRoutes Регистрируем новые пути в маршрутизаторе
func (fc *FriendController) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/friends", fc.GetFriendsByUserId)
	router.POST("/request", fc.SendFriendRequest)
	router.GET("/reqlis", fc.GetFriendRequests)
	router.POST("/response", fc.RespondToFriendRequest)
	router.DELETE("/drequest", fc.DeleteFriendRequest) // Удаление запроса
	router.DELETE("/dfriend", fc.RemoveFriend)         // Удаление друга
}
