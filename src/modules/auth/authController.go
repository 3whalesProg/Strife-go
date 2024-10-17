package controllers

import (
	"net/http"

	"github.com/3whalesProg/Strife-go/src/db"
	"github.com/3whalesProg/Strife-go/src/models"
	"github.com/3whalesProg/Strife-go/src/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// AuthController содержит методы для аутентификации
type AuthController struct{}

// NewAuthController создает новый экземпляр AuthController
func NewAuthController() *AuthController {
	return &AuthController{}
}

// Register обрабатывает регистрацию пользователя
func (ac *AuthController) Register(c *gin.Context) {
	var json struct {
		Login       string `json:"login" binding:"required"`
		Email       string `json:"email" binding:"required"`
		Password    string `json:"password" binding:"required"`
		Description string `json:"description" `
		Nickname    string `json:"nickname"`
	}

	// Привязываем JSON к структуре
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(json.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := models.Users{
		Login:    json.Login,
		Email:    json.Email,
		Password: string(hashedPassword)}

	if err := db.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	token, err := utils.GenerateJWT(user.ID, user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
func (ac *AuthController) Login(c *gin.Context) {
	var json struct {
		LoginOrEmail string `json:"loginOrEmail" binding:"required"`
		Password     string `json:"password" binding:"required"`
	}

	// Привязываем JSON к структуре
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Ищем пользователя по логину или email
	var user models.Users
	if err := db.DB.Where("login = ? OR email = ?", json.LoginOrEmail, json.LoginOrEmail).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid login or email"})
		return
	}

	// Проверяем пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(json.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
		return
	}

	// Генерируем JWT токен
	token, err := utils.GenerateJWT(user.ID, user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Возвращаем токен
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (ac *AuthController) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/register", ac.Register)
	router.POST("/login", ac.Login)
}
