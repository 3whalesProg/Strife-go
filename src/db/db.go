package db

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB // глобальная переменная для хранения подключения к базе данных

func ConnectToDB() {
	var err error
	dsn := "host=localhost user=postgres password=19283746 dbname=Strife port=5432 sslmode=disable"
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{}) // просто присваиваем глобальной переменной
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
}
