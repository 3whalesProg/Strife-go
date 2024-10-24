package models

import (
	"gorm.io/gorm"
)

type Messages struct {
	gorm.Model
	ID       uint   `gorm:"primaryKey;autoIncrement"`
	Content  string `gorm:"type:text;not null"`  // Текст сообщения
	SenderID uint   `gorm:"not null"`            // ID отправителя сообщения
	ChatID   uint   `gorm:"not null"`            // ID чата, в который отправлено сообщение
	Sender   Users  `gorm:"foreignKey:SenderID"` // Связь с отправителем (User)
	Chat     Chats  `gorm:"foreignKey:ChatID"`   // Связь с чатом (Chat)
	FileURL  string `json:"file_url"`
	ImageURL string `json:"image_url"`
}
