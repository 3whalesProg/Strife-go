package socket

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/3whalesProg/Strife-go/src/utils"
	"github.com/gorilla/websocket"
)

var activeClients = make(map[uint]*websocket.Conn) // Ключ - ID пользователя, значение - соединение
var mu sync.Mutex                                  // Для защиты активных клиентов
type ActiveChat struct {
	UserID uint // ID пользователя
}

type Claims struct {
	ID uint
	// Другие поля токена...
}

var activeChats = make(map[uint][]uint)

// Обновление для Gorilla WebSocket
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Разрешаем запросы из любого источника (по необходимости)
		return true
	},
}

// Он конэкшн
func HandleConnections(w http.ResponseWriter, r *http.Request) {
	// Апгрейд функции не убираем
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Ошибка апгрейда до WebSocket:", err)
		return
	}
	defer conn.Close() //тут закрываем сокет если токена нет

	// получаем айди
	claims, err := utils.ExtractTokenAndVerify(r)
	if err != nil {
		log.Println("Ошибка проверки токена:", err)
		return
	}
	userID := claims.ID
	log.Println("Клиент подключен:", userID)
	// Сохраняем соединение
	mu.Lock()
	activeClients[userID] = conn
	mu.Unlock()

	//подписываемся на чаты
	if err := subscribeChatNotifications(userID); err != nil {
		log.Println("Ошибка подписки на уведомления чатов:", err)
		return
	}

	fmt.Println(activeChats)

	// Цикл чтения сообщений
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Ошибка чтения сообщения:", err)
			break
		}
		log.Printf("Сообщение от пользователя %d: %s", userID, message)
	}

	// Отключение клиента
	DisconnectClient(userID, conn)
}

func DisconnectClient(userID uint, conn *websocket.Conn) {
	// Закрываем соединение
	err := conn.Close()
	if err != nil {
		log.Println("Ошибка при закрытии соединения:", err)
	}
	//отписываемся от чатов
	unsubscribeChatsNotifications(userID)

	//удаляем из эктив
	mu.Lock()
	delete(activeClients, userID)
	mu.Unlock()

	log.Printf("Клиент отключен: %d", userID)
}
