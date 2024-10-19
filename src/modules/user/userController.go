package controllers

import (
	"fmt"
	"net/http"
	"regexp"

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

func (uc *UserController) GetUserByLoginController(c *gin.Context) {
	// Структура для принятия JSON запроса
	var request struct {
		Login string `json:"login"` // Получаем логин пользователя
	}

	// Проверяем правильность JSON
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Проверяем, передан ли логин
	if request.Login == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Login is required"})
		return
	}

	// Ищем пользователя по логину
	var user models.Users
	if err := db.DB.Where("login = ?", request.Login).First(&user).Error; err != nil {
		// Если пользователь не найден, возвращаем ошибку
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Возвращаем успешный ответ с информацией о пользователе
	c.JSON(http.StatusOK, gin.H{
		"message": "User found",
		"user": gin.H{
			"id":       user.ID,
			"login":    user.Login,
			"email":    user.Email,
			"nickname": user.Nickname,
		},
	})
}

func (uc *UserController) GetUserByIDController(c *gin.Context) {
	// Структура для принятия JSON запроса
	var request struct {
		ID uint `json:"id"` // Получаем ID пользователя
	}

	// Проверяем правильность JSON
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Проверяем, передан ли ID
	if request.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}

	// Ищем пользователя по ID
	var user models.Users
	if err := db.DB.Where("id = ?", request.ID).First(&user).Error; err != nil {
		// Если пользователь не найден, возвращаем ошибку
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Возвращаем успешный ответ с информацией о пользователе
	c.JSON(http.StatusOK, gin.H{
		"message": "User found",
		"user": gin.H{
			"id":       user.ID,
			"login":    user.Login,
			"email":    user.Email,
			"nickname": user.Nickname,
		},
	})
}

func (ac *UserController) CName(c *gin.Context) {
	var json struct {
		Nickname string `json:"nickname" binding:"required"`
	}
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ввод"})
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
	if err := db.DB.First(&user, claims.ID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Проверяем, не совпадает ли новый ник с текущим
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

// обновление аватара пользователя
func (uc *UserController) UpdateAvatar(c *gin.Context) {
	var json struct {
		AvatarURL string `json:"avatar_url" binding:"required"`
	}

	// привязка данных
	if err := c.ShouldBindJSON(&json); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// валидносьб url
	if !isValidURL(json.AvatarURL) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL format"})
		return
	}

	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token not provided"})
		return
	}

	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	// проверка токена
	claims, err := utils.CheckUser(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
		return
	}

	// получаем пользователся из бд по клаймсу JWT токана
	var user models.Users
	if err := db.DB.First(&user, claims.ID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// обновление avatar url в тайпе модели юзера
	user.AvatarURL = json.AvatarURL
	if err := db.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update avatar"})
		return
	}

	// по примеру возращаем измененое тело как влад сверху писал
	c.JSON(http.StatusOK, gin.H{
		"id":         user.ID,
		"login":      user.Login,
		"email":      user.Email,
		"nickname":   user.Nickname,
		"role":       user.Role,
		"avatar_url": user.AvatarURL,
	})
}

// валидация ссылки
func isValidURL(url string) bool {
	re := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	return re.MatchString(url)
}

// RegisterRoutes регистрирует маршруты контроллера
func (uc *UserController) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/user", uc.GetUserInfo)
	router.POST("/getUserByLogin", uc.GetUserByLoginController)
	router.POST("/getUserById", uc.GetUserByLoginController)
	router.PATCH("/cname", uc.CName)
	router.PATCH("/description", uc.UpdateDescription) // Сосем член по кд у гпт
	router.PATCH("/avatar", uc.UpdateAvatar)
	router.PATCH("/description", uc.UpdateDescription)
}
