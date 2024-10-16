package models

import (
	"gorm.io/gorm"
)

type Users struct {
	gorm.Model
	ID       uint       `gorm:"primaryKey;autoIncrement"` // ID будет Primary Key с автоинкрементом
	Login    string     `gorm:"unique;size:256;not null"` // Логин, уникальный и обязательный (varchar(256))
	Email    string     `gorm:"unique;size:256;not null"` // Email, уникальный и обязательный (varchar(256))
	Nickname string     `gorm:"size:256"`                 // Никнейм (varchar(256))
	Password string     `gorm:"not null"`                 // Пароль, обязательный (text)
	Role     string     `gorm:"size:50;default:'user'"`
	Chats    []*Chats   `gorm:"many2many:user_chats;"`
	Messages []Messages `gorm:"foreignKey:SenderID"`
	// Включает стандартные поля: CreatedAt, UpdatedAt, DeletedAt
}
