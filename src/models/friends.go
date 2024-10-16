package models

import (
	"gorm.io/gorm"
)

type Friends struct {
	gorm.Model
	//id
	ID uint `gorm:"primaryKey;autoIncrement"`
	// id ключи к userу и другу
	UserID   uint `gorm:"not null"`
	FriendID uint `gorm:"not null"`

	User   Users `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Friend Users `gorm:"foreignKey:FriendID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}
