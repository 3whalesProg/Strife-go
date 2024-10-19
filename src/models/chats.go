package models

import (
	"gorm.io/gorm"
)

type Chats struct {
	gorm.Model
	ID          uint       `gorm:"primaryKey;autoIncrement"`
	Title       string     `gorm:"size:256;not null"`
	IsTetATet   bool       `gorm:"not null;default:false"` // Добавлено поле для обозначения личного чата
	Recipient   *Users     `gorm:"foreignKey:RecipientID"` // Получатель в случае личного чата
	RecipientID uint       // ID получателя, если чат личный
	Messages    []Messages `gorm:"foreignKey:ChatID"`
	Users       []*Users   `gorm:"many2many:user_chats;"`
}
