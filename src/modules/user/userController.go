package controllers

import (
	"fmt"
	"net/http"

	"github.com/3whalesProg/Strife-go/src/db"
	"github.com/3whalesProg/Strife-go/src/models"
	"github.com/3whalesProg/Strife-go/src/utils"
	"github.com/gin-gonic/gin"
)

// UserController содержит методы для управления пользователями
type UserController struct{}

var userCache = make(map[uint]models.Users) // используем uint для ключей

// NewUserController создает новый экземпляр UserController
func NewUserController() *UserController {
	return &UserController{}
}

// getToken извлекает токен из заголовка Authorization
func getToken(c *gin.Context) (string, error) {
	token := c.GetHeader("Authorization")
	if token == "" {
		return "", fmt.Errorf("токен не предоставлен")
	}
	if len(token) > 7 && token[:7] == "Bearer " {
		return token[7:], nil
	}
	return token, nil
}

// getUserByClaims получает пользователя по Claims
func (uc *UserController) getUserByClaims(claims *utils.Claims) (models.Users, error) {
	// Проверяем кэш
	if user, found := userCache[claims.ID]; found {
		return user, nil // Возвращаем пользователя из кэша
	}

	var user models.Users
	if err := db.DB.First(&user, claims.ID).Error; err != nil {
		return user, err
	}

	// Сохраняем пользователя в кэш
	userCache[claims.ID] = user
	return user, nil
}

// handleUserInfo обрабатывает общий код для получения информации о пользователе
func (uc *UserController) handleUserInfo(c *gin.Context) (models.Users, error) {
	token, err := getToken(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return models.Users{}, err
	}

	claims, err := utils.CheckUser(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return models.Users{}, err
	}

	return uc.getUserByClaims(claims)
}

// GetUserInfo обрабатывает запрос на получение информации о пользователе
func (uc *UserController) GetUserInfo(c *gin.Context) {
	user, err := uc.handleUserInfo(c)
	if err != nil {
		return // Ошибка уже обработана в handleUserInfo
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       user.ID,
		"login":    user.Login,
		"email":    user.Email,
		"nickname": user.Nickname,
		"role":     user.Role,
	})
}

// CName обновляет никнейм пользователя
func (uc *UserController) CName(c *gin.Context) {
	var json struct {
		Nickname string `json:"nickname" binding:"required"`
	}

	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ввод"})
		return
	}

	user, err := uc.handleUserInfo(c)
	if err != nil {
		return // Ошибка уже обработана в handleUserInfo
	}

	if user.Nickname == json.Nickname {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Новый никнейм должен отличаться от текущего"})
		return
	}

	user.Nickname = json.Nickname
	if err := db.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось обновить никнейм"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       user.ID,
		"login":    user.Login,
		"email":    user.Email,
		"nickname": user.Nickname,
		"role":     user.Role,
	})
}

// UpdateDescription обновляет описание пользователя
func (uc *UserController) UpdateDescription(c *gin.Context) {
	var json struct {
		Description string `json:"description" binding:"required"`
	}

	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ввод"})
		return
	}

	user, err := uc.handleUserInfo(c)
	if err != nil {
		return // Ошибка уже обработана в handleUserInfo
	}

	user.Description = json.Description
	if err := db.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось обновить описание"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":          user.ID,
		"login":       user.Login,
		"email":       user.Email,
		"nickname":    user.Nickname,
		"role":        user.Role,
		"description": user.Description,
	})
}

// RegisterRoutes регистрирует маршруты контроллера
func (uc *UserController) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/user", uc.GetUserInfo)
	router.PATCH("/cname", uc.CName)
	router.PATCH("/description", uc.UpdateDescription)
}
