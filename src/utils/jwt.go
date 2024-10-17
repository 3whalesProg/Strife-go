package utils

import (
	"errors"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var jwtSecret = []byte("your-secret-key")

// Claims структура для хранения информации о пользователе
type Claims struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.StandardClaims
}

// CheckUser проверяет токен пользователя и возвращает claims
func CheckUser(token string) (*Claims, error) {
	// Удаляем префикс "Bearer " если он присутствует
	if strings.HasPrefix(token, "Bearer ") {
		token = strings.TrimPrefix(token, "Bearer ")
	}

	claims := &Claims{}
	tokenParsed, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}

	if !tokenParsed.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// GenerateJWT генерирует новый JWT токен
func GenerateJWT(id uint, email string, role string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour) // Токен будет действовать 24 часа
	claims := &Claims{
		ID:    id,
		Email: email,
		Role:  role,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
