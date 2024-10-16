package socket

import (
	"fmt"
	"log"
	"net/url"

	"github.com/3whalesProg/Strife-go/src/db"
	"github.com/3whalesProg/Strife-go/src/models"
	"github.com/3whalesProg/Strife-go/src/utils"
	socketio "github.com/googollee/go-socket.io"
)

var activeClients = make(map[uint]socketio.Conn) // Ключ - ID пользователя, значение - соединение
type ActiveChat struct {
	UserID uint // ID пользователя
}

var activeChats = make(map[uint][]uint)

// CreateServer создает и настраивает сервер WebSocket
func CreateServer() *socketio.Server {
	server := socketio.NewServer(nil)

	server.OnConnect("/", func(s socketio.Conn) error {
		// Получаем URL
		rawQuery := s.URL().RawQuery
		log.Println("RawQuery:", rawQuery)

		// Разбираем URL и получаем токен
		values, err := url.ParseQuery(rawQuery)
		if err != nil {
			log.Println("Ошибка разбора URL:", err)
			return nil // Или вернуть ошибку
		}

		// Предполагаем, что токен передается в параметре "token"
		token := values.Get("token")
		if token == "" {
			log.Println("Токен не найден в запросе.")
			return nil // Или вернуть ошибку
		}

		log.Println("Токен:", token)

		// Проверяем токен и получаем информацию из Claims
		claims, err := utils.CheckUser(token) // Функция для декодирования токена
		if err != nil {
			log.Println("Ошибка проверки токена:", err)
			return nil // Или вернуть ошибку
		}

		userID := claims.ID // Предполагаем, что ID пользователя находится в claims

		log.Println("Клиент подключен:", userID)
		activeClients[userID] = s // Сохраняем соединение под ID пользователя

		var user models.Users
		if err := db.DB.Preload("Chats").First(&user, claims.ID).Error; err != nil {
			log.Println("Ошибка получения пользователя:", err)
			return nil
		}
		fmt.Println("Пользователь:", user)
		fmt.Println("Чаты пользователя:")
		for _, chat := range user.Chats {
			activeChats[chat.ID] = append(activeChats[chat.ID], claims.ID)
		}

		fmt.Println(activeChats)
		return nil
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		// Удаляем соединение по ID пользователя
		for userID, conn := range activeClients {
			if conn.ID() == s.ID() {
				log.Println("Клиент отключен:", userID)
				delete(activeClients, userID) // Удаляем соединение при отключении
				break
			}
		}
	})

	// SetupChatRoutes(server)

	return server
}
