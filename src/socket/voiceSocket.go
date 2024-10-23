package socket

import (
	"fmt"
	"log"

	"github.com/3whalesProg/Strife-go/src/db"
	"github.com/3whalesProg/Strife-go/src/models"
	"github.com/gin-gonic/gin"
)

func LeaveRoom(roomId uint, userID uint) {
	// Удаляем пользователя из activeRooms[roomId], если он там был
	if users, exists := activeRooms[roomId]; exists {
		for i, id := range users {
			if id == userID {
				activeRooms[roomId] = append(users[:i], users[i+1:]...)
				break
			}
		}
		// Если в комнате больше нет пользователей, можно удалить саму комнату
		if len(activeRooms[roomId]) == 0 {
			delete(activeRooms, roomId)
		}
	}

	// Удаляем пользователя из requestRooms[roomId], если он там был
	if users, exists := requestRooms[roomId]; exists {
		for i, id := range users {
			if id == userID {
				requestRooms[roomId] = append(users[:i], users[i+1:]...)
				break
			}
		}
		// Если в списке запросов больше нет пользователей, удаляем запись
		if len(requestRooms[roomId]) == 0 {
			delete(requestRooms, roomId)
		}
	}

	// Отправляем уведомление другим пользователям в activeRooms
	if userIDs, ok := activeRooms[roomId]; ok {
		for _, otherUserID := range userIDs {
			if client, exists := activeClients[otherUserID]; exists {
				err := client.WriteJSON(gin.H{
					"event":  "leaveRoom", // Уведомление о выходе
					"UserID": userID,
				})
				if err != nil {
					log.Printf("Ошибка отправки сообщения пользователю %d: %v", otherUserID, err)
				}
			}
		}
	} else {
		log.Printf("Нет активных пользователей для комнаты с ID %d", roomId)
	}
}

func CreateRoom(userID uint, chatID uint) {
	fmt.Println(chatID)
	activeRooms[chatID] = append(activeRooms[chatID], userID)
	fmt.Println(activeRooms)
	if userIDs, exists := activeChats[chatID]; exists {
		mu.Lock()
		fmt.Println(userIDs)
		for _, otherUserID := range userIDs {
			// Пропускаем текущего пользователя
			if otherUserID == userID {
				continue
			}
			// Вызываем RequestJoinRoom для всех остальных пользователей
			RequestJoinRoom(chatID, otherUserID)
			mu.Unlock()
		}
	} else {
		log.Printf("Чат с ID %d не найден в activeChats", chatID)
	}
}

func RequestJoinRoom(roomId uint, userID uint) {
	requestRooms[roomId] = append(requestRooms[roomId], userID)
	if client, exists := activeClients[userID]; exists {
		var sender models.Users
		if err := db.DB.First(&sender, userID).Error; err != nil {
			return
		}
		err := client.WriteJSON(gin.H{
			"event":  "call", // Добавляем событие
			"RoomId": roomId,
		})
		if err != nil {
			log.Printf("Ошибка отправки сообщения пользователю %d: %v", userID, err)
		}
	}
}

func AcceptJoinRoom(roomId uint, userID uint) {
	if users, exists := requestRooms[roomId]; exists {
		mu.Lock()
		for i, id := range users {
			if id == userID {
				requestRooms[roomId] = append(users[:i], users[i+1:]...)
				break
			}
		}
		// Если в массиве больше нет пользователей, удаляем запись
		if len(requestRooms[roomId]) == 0 {
			delete(requestRooms, roomId)
		}
		mu.Unlock()
	}
	activeRooms[roomId] = append(activeRooms[roomId], userID)
	if userIDs, ok := activeRooms[roomId]; ok {
		mu.Lock()
		for _, userID := range userIDs {
			if client, exists := activeClients[userID]; exists {
				var sender models.Users
				if err := db.DB.First(&sender, userID).Error; err != nil {
					return
				}
				err := client.WriteJSON(gin.H{
					"event": "acceptCall", // Добавляем событие
					"User":  sender,
				})
				if err != nil {
					log.Printf("Ошибка отправки сообщения пользователю %d: %v", userID, err)
				}
			}
		}
		mu.Unlock()
	} else {
		log.Printf("Нет подписанных пользователей для чата с ID %d", roomId)
	}
}

func RejectJoinRoom(roomId uint, userID uint) {
	if users, exists := requestRooms[roomId]; exists {
		mu.Lock()
		for i, id := range users {
			if id == userID {
				requestRooms[roomId] = append(users[:i], users[i+1:]...)
				break
			}
		}
		// Если в массиве больше нет пользователей, удаляем запись
		if len(requestRooms[roomId]) == 0 {
			delete(requestRooms, roomId)
		}
		mu.Unlock()
	}
	if userIDs, ok := activeRooms[roomId]; ok {
		mu.Lock()
		for _, userID := range userIDs {
			if client, exists := activeClients[userID]; exists {
				var sender models.Users
				if err := db.DB.First(&sender, userID).Error; err != nil {
					return
				}
				err := client.WriteJSON(gin.H{
					"event": "rejectCall", // Добавляем событие
					"User":  sender,
				})
				if err != nil {
					log.Printf("Ошибка отправки сообщения пользователю %d: %v", userID, err)
				}
			}
		}
		mu.Unlock()
	} else {
		log.Printf("Нет подписанных пользователей для чата с ID %d", roomId)
	}
}

func unsubscribeRooms(userID uint) {
	// Получаем информацию о пользователе
	var user models.Users
	if err := db.DB.Preload("Chats").First(&user, userID).Error; err != nil {
		log.Println("Ошибка получения пользователя:", err)
		return
	}

	// Удаляем пользователя из каждого чата, в котором он участвовал
	for _, chat := range user.Chats {
		mu.Lock()
		// Ищем индекс пользователя в списке чата
		for i, uid := range activeRooms[chat.ID] {
			if uid == userID {
				// Удаляем пользователя из этого чата
				activeRooms[chat.ID] = append(activeRooms[chat.ID][:i], activeRooms[chat.ID][i+1:]...)
				break
			}
		}
		for i, uid := range requestRooms[chat.ID] {
			if uid == userID {
				// Удаляем пользователя из этого чата
				requestRooms[chat.ID] = append(requestRooms[chat.ID][:i], requestRooms[chat.ID][i+1:]...)
				break
			}
		}

		// Если в чате больше нет пользователей, удаляем чат из activeChats
		if len(activeRooms[chat.ID]) == 0 {
			delete(activeRooms, chat.ID)
		}
		if len(requestRooms[chat.ID]) == 0 {
			delete(requestRooms, chat.ID)
		}
		mu.Unlock()
	}
}
