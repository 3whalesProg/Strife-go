package models

import (
	"errors"
	"gorm.io/gorm"
	"regexp"
)

type Users struct {
	gorm.Model
	ID          uint   `gorm:"primaryKey;autoIncrement"` // ID будет Primary Key с автоинкрементом
	Login       string `gorm:"unique;size:20;not null"`  // Логин, уникальный и обязательный (varchar(256))
	Email       string `gorm:"unique;size:256;not null"` // Email, уникальный и обязательный (varchar(256))
	Nickname    string `gorm:"size:20"`                  // Никнейм (varchar(20))
	Password    string `gorm:"not null"`                 // Пароль, обязательный (text)
	Description string `gorm:"size:256"`
	Role        string `gorm:"size:20;default:'user'"`
	// Включает стандартные поля: CreatedAt, UpdatedAt, DeletedAt
}

// ValidateEmail проверяет, что email имеет правильный формат
func ValidateEmail(email string) bool {
	re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
	return re.MatchString(email)
}

// BeforeCreate выполняет валидации перед созданием записи
func (u *Users) BeforeCreate(tx *gorm.DB) (err error) {
	if !ValidateEmail(u.Email) {
		return errors.New("invalid email format")
	}

	if len(u.Password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	// Если Nickname не задан, установить его равным Login
	if u.Nickname == "" {
		u.Nickname = u.Login
	}

	return nil
}
