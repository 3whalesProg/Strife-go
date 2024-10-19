package models

import (
	"gorm.io/gorm"
)

// Friends модель для хранения информации о дружбе
type Friends struct {
	gorm.Model
	UserID     uint `gorm:"not null"`      // ID пользователя
	FriendID   uint `gorm:"not null"`      // ID друга
	IsFavorite bool `gorm:"default:false"` // Поле для отметки избранных друзей
}
