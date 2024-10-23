package models

import (
	"time"

	"gorm.io/gorm"
)

type FriendRequest struct {
	gorm.Model
	SenderID    uint      `json:"sender_id" gorm:"not null"`
	RecipientID uint      `json:"recipient_id" gorm:"not null"`
	CreatedAt   time.Time `json:"created_at"` // когда создалосб
	UpdatedAt   time.Time `json:"updated_at"` // когда обновлено
}
