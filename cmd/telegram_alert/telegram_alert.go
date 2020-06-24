package main

import (
	"os"
	"strconv"
	"strings"

	"github.com/stevenwilkin/treasury/telegram"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	chatId, err := strconv.Atoi(os.Getenv("TELEGRAM_CHAT_ID"))
	if err != nil {
		panic(err)
	}

	telegramBot := telegram.Telegram{
		ApiToken: os.Getenv("TELEGRAM_API_TOKEN"),
		ChatId:   chatId}

	telegramBot.Alert(strings.Join(os.Args[1:], " "))
}
