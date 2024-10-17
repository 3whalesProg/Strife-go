package socket

import "fmt"

func SendMessage(userID uint, chatID uint, msg string) {
	for _, aboba := range activeChats[uint(chatID)] {
		activeClients[aboba].Emit("chat.message", msg)
	}
}

func Hello() {
	for userID, conn := range activeClients {
		// Логируем информацию о пользователе и его соединении
		fmt.Printf("Отправка сообщения пользователю с ID: %d\n", userID)

		// Проверяем, что соединение активно
		if conn != nil {
			// Отправляем сообщение клиенту
			conn.Emit("hui", "vlad huiing")
		} else {
			// Логируем, если соединение неактивно
			fmt.Printf("Соединение для пользователя %d не активно\n", userID)
		}
	}
}
