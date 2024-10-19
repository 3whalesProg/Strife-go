package controllers

import (
	"github.com/3whalesProg/Strife-go/src/db"
	"github.com/3whalesProg/Strife-go/src/models"
	"github.com/3whalesProg/Strife-go/src/utils"
	"github.com/gin-gonic/gin"
	"net/http"
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
func (ac *UserController) CName(c *gin.Context) {
	var json struct {
		Nickname string `json:"nickname" binding:"required"`
	}

	// Привязываем данные из JSON к структуре
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

func (ac *UserController) GetUserChats(c *gin.Context) {


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
	if err := db.DB.Preload("Chats").First(&user, claims.ID).Error; err != nil {
		return
	}

	// Возвращаем информацию о пользователе
	c.JSON(http.StatusOK, gin.H{
		"chats": user,
	})
}

func (ac *UserController) Hello(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"chats": 111,
	})
}

func (ac *UserController) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/register", ac.GetUserInfo)
	router.GET("/getUserChats", ac.GetUserChats)
	router.GET("/hello", ac.Hello)
	if err := db.DB.First(&user, claims.ID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Проверяем, не совпадает ли новый ник с текущим
	if user.Nickname == json.Nickname {
		c.JSON(http.StatusBadRequest, gin.H{"error": "New nickname must be different from the current one"})
		return
	}

	// Обновляем поле Nickname в базе данных
	user.Nickname = json.Nickname
	if err := db.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update nickname"})
		return
	}

	// Возвращаем информацию о пользователе с обновленным ником
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

	// Привязываем данные из JSON к структуре
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
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

	// Получаем пользователя из базы данных по ID, который хранится в JWT токене
	var user models.Users
	if err := db.DB.First(&user, claims.ID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Обновляем поле Description в базе данных
	user.Description = json.Description
	if err := db.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update description"})
		return
	}

	// Возвращаем информацию о пользователе с обновленным описанием
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
	router.POST("/register", uc.GetUserInfo)
	router.PATCH("/cname", uc.CName)
	router.PATCH("/description", uc.UpdateDescription) // Добавляем маршрут для обновления описания
}
