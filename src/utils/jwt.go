package utils

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var jwtSecret = []byte("your-secret-key")

type Claims struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.StandardClaims
}

// Генерация JWT токена
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

// Валидация JWT токена
func ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, err
	}

	return claims, nil
}

func CheckUser(tokenString string) (*Claims, error) {
	claims, err := ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	// Проверяем, истек ли срок действия токена
	if claims.ExpiresAt < time.Now().Unix() {
		return nil, errors.New("token has expired")
	}

	return claims, nil
}

func ExtractTokenAndVerify(r *http.Request) (*Claims, error) {
	// Получаем строку запроса
	rawQuery := r.URL.RawQuery
	log.Println("RawQuery:", rawQuery)

	// Разбираем строку запроса и получаем токен
	values, err := url.ParseQuery(rawQuery)
	if err != nil {
		log.Println("Ошибка разбора URL:", err)
		return nil, err
	}

	// Получаем токен из параметра "token"
	token := values.Get("token")
	if token == "" {
		log.Println("Токен не найден в запросе.")
		return nil, fmt.Errorf("токен не найден")
	}

	log.Println("Токен:", token)

	// Проверяем токен и получаем информацию из Claims
	claims, err := CheckUser(token) // Предполагается, что у вас уже есть функция CheckUser для проверки токена
	if err != nil {
		log.Println("Ошибка проверки токена:", err)
		return nil, err
	}

	return claims, nil
}
