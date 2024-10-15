package main

import (
	"github.com/3whalesProg/Strife-go/src/db"
	"github.com/3whalesProg/Strife-go/src/migrate"
	"github.com/3whalesProg/Strife-go/src/models"
	"github.com/3whalesProg/Strife-go/src/routes"
	"github.com/gin-gonic/gin"
)

func init() {
	db.ConnectToDB()
}

func main() {
	db.DB.Migrator().DropTable(&models.Users{})
	migrate.Migrate()
	router := gin.Default()

	// Создаем группу маршрутов для API v1
	apiGroup := router.Group("/api/v1")
	routes.IndexRouter(apiGroup) // Подключаем маршруты из IndexRouter

	// Запускаем сервер на порту 8080
	router.Run(":8080") // по умолчанию запускается на http://localhost:8080
}
