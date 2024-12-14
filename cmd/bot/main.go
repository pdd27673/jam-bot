package main

import (
	"jam-bot/internal/bot"
	"log"
)

func main() {
	// init bot
	err := bot.StartBot()
	if err != nil {
		log.Fatalf("[ERROR] error starting the bot: %s", err)
	}
}
