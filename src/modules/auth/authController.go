package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
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
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	// Привязываем JSON к структуре
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Здесь добавь логику для сохранения пользователя в базе данных

	c.JSON(http.StatusOK, gin.H{"message": "Registration successful", "user": json.Username})
}

func (ac *AuthController) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/register", ac.Register)
}
