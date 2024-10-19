package migrate

import (
	"github.com/3whalesProg/Strife-go/src/db"
	"github.com/3whalesProg/Strife-go/src/models"
)

func Migrate() {
	db.ConnectToDB()
	db.DB.AutoMigrate(&models.Users{})
	db.DB.AutoMigrate(&models.Messages{})
	db.DB.AutoMigrate(&models.Chats{})
	db.DB.AutoMigrate(&models.Friends{})
	db.DB.AutoMigrate(&models.FriendRequest{})
}
