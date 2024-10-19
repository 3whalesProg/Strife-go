package main

import (
	"github.com/3whalesProg/Strife-go/src/db"
	"github.com/3whalesProg/Strife-go/src/migrate"
	"github.com/3whalesProg/Strife-go/src/routes"
	"github.com/3whalesProg/Strife-go/src/socket"
	"github.com/gin-contrib/cors" // Импортируем пакет cors
	"github.com/gin-gonic/gin"
)

func init() {
	db.ConnectToDB()
}

func main() {
	migrate.Migrate()
	router := gin.Default()

	// Настройка CORS
	config := cors.DefaultConfig()                                            // Используем стандартную конфигурацию
	config.AllowOrigins = []string{"*"}                                       // Разрешаем доступ со всех доменов (в реальных приложениях используйте конкретные домены)
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"} // Разрешаем все основные HTTP-методы
	config.AllowHeaders = []string{"Content-Type", "Authorization"}           // Разрешаем заголовки Content-Type и Authorization
	router.Use(cors.New(config))                                              // Включаем CORS для всего приложения

	// Создаем группу маршрутов для API v1
	apiGroup := router.Group("/api/v1")
	routes.IndexRouter(apiGroup) // Подключаем маршруты из IndexRouter

	router.GET("/ws", func(c *gin.Context) {
		socket.HandleConnections(c.Writer, c.Request)
	})

	// Маршрутизация для WebSocket
	// router.GET("/socket.io/", gin.WrapH(server)) // Обработка запросов на WebSocket

	// go server.Serve()
	// Запускаем сервер на порту 8080
	router.Run(":8080") // по умолчанию запускается на http://localhost:8080
}
