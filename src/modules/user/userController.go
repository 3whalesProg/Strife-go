package controllers

import (
	"net/http"

	"github.com/3whalesProg/Strife-go/src/db"
	"github.com/3whalesProg/Strife-go/src/models"
	"github.com/3whalesProg/Strife-go/src/utils"
	"github.com/gin-gonic/gin"
)

// AuthController содержит методы для аутентификации
type UserController struct{}

// NewAuthController создает новый экземпляр AuthController
func NewUserController() *UserController {
	return &UserController{}
}

// Register обрабатывает регистрацию пользователя
func (ac *UserController) GetUserInfo(c *gin.Context) {
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

func (ac *UserController) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/register", ac.GetUserInfo)
}
