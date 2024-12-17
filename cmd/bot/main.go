package main

import (
	"jam-bot/internal/bot"
	"jam-bot/internal/config"
	"jam-bot/internal/server"
	"log"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("[ERROR] Failed to load config: %v", err)
	}

	// Start the Discord bot in a separate goroutine
	go func() {
		err = bot.StartBot()
		if err != nil {
			log.Fatalf("[ERROR] Error starting the bot: %v", err)
		}
	}()

	// Start the HTTP server
	server.StartServer(cfg)
}
