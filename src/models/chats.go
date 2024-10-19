package models

import (
	"gorm.io/gorm"
)

type Chats struct {
	gorm.Model
	ID       uint       `gorm:"primaryKey;autoIncrement"`
	Title    string     `gorm:"size:256;not null"`
	Messages []Messages `gorm:"foreignKey:ChatID"`
	Users    []*Users   `gorm:"many2many:user_chats;"`
}
