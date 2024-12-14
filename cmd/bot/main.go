package main

import (
	"log"

	"github.com/pdd27673/jam-bot/internal/bot"
)

func main() {
	// init bot
	err := bot.StartBot()
	if err != nil {
		log.Fatalf("[ERROR] error starting the bot: %s", err)
	}
}
