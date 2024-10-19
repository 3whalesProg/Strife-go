package socket

import (
	"fmt"
	"log"

	"github.com/3whalesProg/Strife-go/src/db"
	"github.com/3whalesProg/Strife-go/src/models"
)

func unsubscribeChatsNotifications(userID uint) {
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
		for i, uid := range activeChats[chat.ID] {
			if uid == userID {
				// Удаляем пользователя из этого чата
				activeChats[chat.ID] = append(activeChats[chat.ID][:i], activeChats[chat.ID][i+1:]...)
				break
			}
		}

		// Если в чате больше нет пользователей, удаляем чат из activeChats
		if len(activeChats[chat.ID]) == 0 {
			delete(activeChats, chat.ID)
		}
		mu.Unlock()
	}
}

func subscribeChatNotifications(userID uint) error {
	// Получаем информацию о пользователе и его чатах
	var user models.Users
	if err := db.DB.Preload("Chats").First(&user, userID).Error; err != nil {
		log.Println("Ошибка получения пользователя:", err)
		return err
	}

	fmt.Println("Пользователь:", user)
	fmt.Println("Чаты пользователя:")

	// Подписываем пользователя на уведомления чатов
	mu.Lock()
	defer mu.Unlock()
	for _, chat := range user.Chats {
		activeChats[chat.ID] = append(activeChats[chat.ID], userID)
	}

	fmt.Println(activeChats)
	return nil
}
