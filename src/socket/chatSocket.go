package socket

import (
	"log"
)

func SendMessage(userID uint, chatID uint, msg string) {
	for _, aboba := range activeChats[uint(chatID)] {
		activeClients[aboba].Emit("chat.message", msg)
	}
}

func Hello() {
	log.Fatalln("123")
}
